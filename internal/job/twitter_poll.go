package job

import (
	"context"
	"fmt"
	db "kimishinu/internal/db/sqlc"
	"kimishinu/internal/errlog"
	"kimishinu/internal/failure"
	"kimishinu/internal/lib"
	"log/slog"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type TwitterPollWorker struct {
	river.WorkerDefaults[TwitterPollArgs]
	Pool        *pgxpool.Pool
	Queries     *db.Queries
	RiverClient *river.Client[pgx.Tx]
	HttpClient  http.Client
	ApiKey      string
}

type User struct {
	Username    string
	Keywords    []string
	Handlername string
	Tier        string
}

const TWTURI string = "https://api.twitterapi.io/twitter/user/last_tweets?userName="

var twtKeywords = []string{"きみが死ぬまで恋をしたい", "きみ死ぬ", "きみ死ぬアニメ", "kimigashinumadekoiwoshitai", "kimishinu", "kimishinuanime"}

var Users = []User{
	{Username: "aooont", Handlername: "aooont", Tier: "fast"},
	{Username: "kimishinu", Handlername: "kimishinu", Tier: "fast"},
	{Username: "anime_kimishinu", Handlername: "anime_kimishinu", Tier: "fast"},
	{Username: "kimishinu_event", Handlername: "kimishinu_event", Tier: "medium"},
	{Username: "mup_contents", Handlername: "mup_contents", Tier: "slow"},
	{Username: "kotobukiyas", Keywords: twtKeywords, Handlername: "kotobukiyas", Tier: "medium"},
	{Username: "VSN_CHILLfigg", Keywords: twtKeywords, Handlername: "vsn_chillfigg", Tier: "slow"},
	{Username: "YuFu_toyomimasu", Keywords: twtKeywords, Handlername: "yufu", Tier: "slow"},
	{Username: "roll2chan", Keywords: twtKeywords, Handlername: "roll2chan", Tier: "medium"},
	{Username: "ROLL2__", Keywords: twtKeywords, Handlername: "roll2", Tier: "medium"},
}

func (t *TwitterPollWorker) Work(ctx context.Context, job *river.Job[TwitterPollArgs]) error {
	currentTier := job.Args.Tier

	trace := "twt_kimishinu_poll_" + fmt.Sprintf("%d", job.ID)

	for _, user := range Users {
		if user.Tier != currentTier {
			continue
		}

		time.Sleep(5 * time.Second)
		err := t.TWTKimishinu(user, trace)
		if err != nil {
			slog.Error("Error occurred while polling Twitter", slog.String("error", err.Error()))
			return err
		}
	}
	return nil
}

func (t *TwitterPollWorker) TWTKimishinu(user User, trace string) error {
	logger := slog.With(slog.String("op", "twt_kimishinu_fetch"), slog.String("account", user.Handlername))

	finalURI := TWTURI + user.Username

	op := func(reason string) string {
		return user.Handlername + "_" + reason
	}

	ctx := context.Background()
	ctx = errlog.SetLogger(ctx, logger)
	tx, err := t.Pool.Begin(ctx)
	if err != nil {
		err = failure.Wrap("twitter.http.pool.transaction.begin", err)
		errlog.General(ctx, op("transaction_begin"), failure.InternalGeneralError, err)
		return err
	}
	defer tx.Rollback(ctx)

	qtx := t.Queries.WithTx(tx)

	reqT, err := http.NewRequest(http.MethodGet, finalURI, nil)
	if err != nil {
		err = failure.Wrap("twitter.http.fetch.request.create", err)
		errlog.Internal(ctx, op("request_create"), failure.InternalGeneralError, err)
		return err
	}
	reqT.Header.Set("Authorization", "Bearer "+t.ApiKey)

	var TweetsResponse lib.TwitterResp
	respT, err := t.HttpClient.Do(reqT)
	if err != nil {
		err = failure.Wrap("twitter.http.fetch.request.send", err)
		errlog.Internal(ctx, op("request_send"), failure.InternalGeneralError, err)
		return err
	}
	defer respT.Body.Close()

	if respT.StatusCode != http.StatusOK {
		err = failure.InternalGeneral(op("request_status_code"), "unexpected status code: "+respT.Status)
		err = failure.Wrap("twitter.http.fetch.request.status_code", err)
		errlog.Internal(ctx, op("request_status_code"), failure.InternalGeneralError, err)
		return err
	}

	err = sonic.ConfigDefault.NewDecoder(respT.Body).Decode(&TweetsResponse)
	if err != nil {
		err = failure.Wrap("twitter.json.fetch.response.decode", err)
		errlog.Internal(ctx, op("response_decode"), failure.InternalGeneralError, err)
		return err
	}

	if len(TweetsResponse.Data.Tweets) == 0 {
		err := failure.NotFound(op("response_tweets"))
		err = failure.Wrap("twitter.http.fetch.response.no_tweets", err)
		errlog.Resource(ctx, op("response_tweets"), failure.NotFoundError, err)
		return err
	}

	pureIds, err := lib.Deduplicate(ctx, keywords, TweetsResponse.Data.Tweets, qtx)
	if err != nil {
		err = failure.Wrap("twitter.http.fetch.deduplicate", err)
		errlog.Database(ctx, op("deduplicate"), err)
		return err
	}
	if pureIds == nil {
		logger.Debug("No new tweets to send to Discord")
		return nil
	}

	for _, id := range pureIds {
		gift := user.Username + "_" + id
		_, err := t.RiverClient.InsertTx(ctx, tx, &DiscordSendArgs{
			Source:     "twitter",
			Identifier: gift,
			Trace:      trace,
		}, nil)
		if err != nil {
			err = failure.Wrap("twitter.river.discord.send.insert", err)
			errlog.General(ctx, op("river_setup"), failure.InternalGeneralError, err)
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		err = failure.Wrap("twitter.http.pool.transaction.commit", err)
		errlog.General(ctx, op("transaction_commit"), failure.InternalGeneralError, err)
		return err
	}

	logger.Info("Successfully sent data to Discord")
	return nil
}
