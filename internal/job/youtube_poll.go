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
	"strings"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type Channel struct {
	PlaylistId  string
	ChannelName string
	Keywords    []string
	HandlerName string
	Tier        string
}

var keywords = []string{"きみが死ぬまで恋をしたい", "iwanttoloveyoutillyourdyingday", "kimigashinumadekoiwoshitai", "kimishinu", "きみ死ぬ"}

var Channels = []Channel{
	{PlaylistId: "UU2QDgMJWRQHEJYyVX3Rt08w", ChannelName: "Ichijinsha PV", Keywords: keywords, HandlerName: "ichijinsha", Tier: "fast"},
	{PlaylistId: "UUyUhtF50BuUjr2jOhxF3IjQ", ChannelName: "ReoNa Official", Keywords: keywords, HandlerName: "reona", Tier: "slow"},
	{PlaylistId: "UUevPX0DXUiKmgH8o19NrE5w", ChannelName: "Infinite Channel", Keywords: keywords, HandlerName: "infinite", Tier: "fast"},
}

const URI = "https://www.googleapis.com/youtube/v3/playlistItems?part=snippet,contentDetails&maxResults=10&playlistId="

type YoutubePollWorker struct {
	river.WorkerDefaults[YoutubePollArgs]
	ApiKey      string
	Pool        *pgxpool.Pool
	Queries     *db.Queries
	HttpClient  http.Client
	RiverClient *river.Client[pgx.Tx]
}

func YTNew(apiKey string, pool *pgxpool.Pool, queries *db.Queries, httpClient *http.Client, riverClient *river.Client[pgx.Tx]) *YoutubePollWorker {
	return &YoutubePollWorker{
		ApiKey:      apiKey,
		Pool:        pool,
		Queries:     queries,
		HttpClient:  *httpClient,
		RiverClient: riverClient,
	}
}

func (w *YoutubePollWorker) Work(ctx context.Context, job *river.Job[YoutubePollArgs]) error {
	currentTier := job.Args.Tier
	trace := "yt_kimishinu_poll_" + fmt.Sprintf("%d", job.ID)
	for _, ch := range Channels {
		if currentTier != ch.Tier {
			continue
		}
		err := w.ytKimishinuPoll(ctx, ch, trace)
		if err != nil {
			slog.Error("Error occurred while polling YouTube", slog.String("error", err.Error()))
			return err
		}
	}
	return nil
}

func (w *YoutubePollWorker) ytKimishinuPoll(ctx context.Context, ch Channel, trace string) error {
	polish := func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
	}
	logger := slog.With(slog.String("op", "yt_kimishinu_poll"), slog.String("channel", polish(ch.ChannelName)))
	logger.Debug(fmt.Sprintf("Fetching %s from Youtube API", ch.ChannelName))

	op := func(reason string) string {
		return ch.HandlerName + "_" + reason
	}

	tx, err := w.Pool.Begin(ctx)
	if err != nil {
		err = failure.Wrap("youtube.http.pool.transaction.begin", err)
		errlog.General(*logger, op("transaction_begin"), failure.InternalGeneralError, err)
		return err
	}
	defer tx.Rollback(ctx)

	qtx := w.Queries.WithTx(tx)

	finalURI := URI + ch.PlaylistId

	reqI, err := http.NewRequest(http.MethodGet, finalURI, nil)
	if err != nil {
		err = failure.Wrap("youtube.http.fetch.request.create", err)
		errlog.Internal(*logger, op("request_create"), failure.InternalGeneralError, err)
		return err
	}
	reqI.Header.Set("Accept", "application/json")
	reqI.Header.Set("X-Goog-Api-Key", w.ApiKey)

	logger.Debug("Sending request to Youtube API", slog.String("url", finalURI))
	respI, err := w.HttpClient.Do(reqI)
	if err != nil {
		err = failure.Wrap("youtube.http.fetch.request.send", err)
		errlog.Internal(*logger, op("request_send"), failure.InternalGeneralError, err)
		return err
	}
	defer respI.Body.Close()

	if respI.StatusCode != http.StatusOK {
		err = failure.InternalGeneral(op("request_status_code"), fmt.Sprintf("unexpected status code: %d", respI.StatusCode))
		err = failure.Wrap("youtube.http.fetch.request.status_code", err)
		errlog.Internal(*logger, op("request_status_code"), failure.InternalGeneralError, err)
		return err
	}

	var videoResp lib.VideoResp
	logger.Debug("Decoding response from Youtube API")
	err = sonic.ConfigDefault.NewDecoder(respI.Body).Decode(&videoResp)
	if err != nil {
		err = failure.Wrap("youtube.http.fetch.response.decode", err)
		errlog.Internal(*logger, op("response_decode"), failure.InternalGeneralError, err)
		return err
	}

	if len(videoResp.Items) == 0 {
		err = failure.InternalGeneral(op("response"), "no items found in response")
		err = failure.Wrap("youtube.http.fetch.response.no_items", err)
		errlog.Resource(*logger, op("response_no_items"), failure.NotFoundError, err)
		return err
	}

	logger.Debug("Processing items from Youtube API response", slog.Int("item_count", len(videoResp.Items)))
	pureIds, err := lib.Deduplicate(ctx, keywords, videoResp, qtx)
	if err != nil {
		err = failure.Wrap("youtube.http.fetch.deduplicate", err)
		errlog.Database(*logger, op("deduplicate"), err)
		return err
	}
	if pureIds == nil {
		logger.Debug("No new videos to send to Discord")
		return err
	}

	lib.Pretty(nil, http.StatusNoContent, nil)

	// We don't do this to tell it "Run every 'x' minutes" - this is purely made for persistence
	for _, id := range pureIds {
		_, err = w.RiverClient.InsertTx(ctx, tx, DiscordSendArgs{
			Source:     "youtube",
			Identifier: id,
			Trace:      trace,
		}, nil)
		if err != nil {
			err = failure.Wrap("youtube.river.discord.send.insert", err)
			errlog.General(*logger, op("river_setup"), failure.InternalGeneralError, err)
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		err = failure.Wrap("youtube.http.pool.transaction.commit", err)
		errlog.General(*logger, op("transaction_commit"), failure.InternalGeneralError, err)
		return err
	}

	logger.Info("successfully sent data to discord")
	return nil
}
