package twitter

import (
	"fmt"
	"kimishinu/internal/auth"
	"kimishinu/internal/discord"
	"kimishinu/internal/errlog"
	"kimishinu/internal/failure"
	"kimishinu/internal/lib"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
)

type User struct {
	Username    string
	Keywords    []string
	Handlername string
}

const URI string = "https://api.twitterapi.io/twitter/user/last_tweets?userName="

var keywords = []string{"きみが死ぬまで恋をしたい", "きみ死ぬ", "kimigashinumadekoiwoshitai", "kimigashinumadekoioshitai", "iwanttoloveyoutillyourdyingday", "iwanttoloveyouuntilyoudie", "kimishinu", "kimishinuanime", "kimidieanime", "kimishineanime"}

var Users = []User{
	{Username: "kimishinu", Handlername: "kimishinu"},
	{Username: "anime_kimishinu", Handlername: "kimishinu_anime"},
	{Username: "kimishinu_event", Handlername: "kimishinu_event"},
	{Username: "kotobukiyas", Keywords: keywords, Handlername: "kotobukiyas"},
	{Username: "VSN_CHILLfigg", Keywords: keywords, Handlername: "vsn_chillfigg"},
	{Username: "ROLL2__", Keywords: keywords, Handlername: "roll2"},
}

func match(text string, keywords []string) bool {
	normalizedText := lib.Normalize(text)
	for _, keyword := range keywords {
		if strings.Contains(normalizedText, keyword) {
			return true
		}
	}

	return false
}

func (t *Twitter) TWTKimishinu(user User) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.With(slog.String("op", "twt_kimishinu_fetch"), slog.String("account", user.Handlername))

		trace, ok := auth.GetTrace(r.Context(), w)
		if !ok {
			return
		}

		finalURI := URI + user.Username

		op := func(reason string) string {
			return user.Handlername + "_" + reason
		}

		reqT, err := http.NewRequest(http.MethodGet, finalURI, nil)
		if err != nil {
			err = failure.Wrap("twitter.http.fetch.request.create", err)
			errlog.Internal(*logger, w, op("request_create"), failure.InternalGeneralError, err, trace)
			return
		}
		reqT.Header.Set("Authorization", "Bearer "+t.ApiKey)

		var TweetsResponse TwitterResp
		respT, err := t.HttpClient.Do(reqT)
		if err != nil {
			err = failure.Wrap("twitter.http.fetch.request.send", err)
			errlog.Internal(*logger, w, op("request_send"), failure.InternalGeneralError, err, trace)
			return
		}
		defer respT.Body.Close()

		if respT.StatusCode != http.StatusOK {
			err = failure.InternalGeneral(op("request_status_code"), "unexpected status code: "+respT.Status)
			err = failure.Wrap("twitter.http.fetch.request.status_code", err)
			errlog.Internal(*logger, w, op("request_status_code"), failure.InternalGeneralError, err, trace)
			return
		}

		err = sonic.ConfigDefault.NewDecoder(respT.Body).Decode(&TweetsResponse)
		if err != nil {
			err = failure.Wrap("twitter.json.fetch.response.decode", err)
			errlog.Internal(*logger, w, op("response_decode"), failure.InternalGeneralError, err, trace)
			return
		}

		if len(TweetsResponse.Data.Tweets) == 0 {
			err := failure.NotFound(op("response_tweets"))
			err = failure.Wrap("twitter.http.fetch.response.no_tweets", err)
			errlog.Http(*logger, w, op("response_tweets"), failure.NotFoundError, err, trace)
			return
		}

		for _, tweet := range TweetsResponse.Data.Tweets {
			normalizedText := lib.Normalize(tweet.Text)
			if match(normalizedText, user.Keywords) {
				continue
			}

			insertedRows, err := t.Queries.InsertTwitterId(r.Context(), tweet.ID)
			if err != nil {
				err = failure.Wrap("twitter.database.twitter.id.deduplicate.check", err)
				errlog.Database(*logger, w, op("deduplicate_check"), trace, err)
				return
			}

			if insertedRows == 0 {
				logger.Debug("Tweet already exists in database, skipping", slog.String("tweet_id", tweet.ID))
				continue
			}

			gift := user.Username + "_" + tweet.ID
			bodyBytes, err := discord.ConvertToDiscordFormat("twitter", gift)
			if err != nil {
				err = failure.Wrap(fmt.Sprintf("twitter.json.%s.response.body.marshal", user.Handlername), err)
				errlog.Json(*logger, w, op("response_body_marshal"), trace, err)
				return
			}

			logger.Debug("Sending data to Discord", slog.String("tweet_id", tweet.ID))
			d := discord.New(t.HttpClient, t.WebhookURL, *logger)

			err = d.Send(w, trace, bodyBytes)
			if err != nil {
				return
			}
		}

		logger.Info("Successfully sent data to Discord")
		lib.Pretty(w, http.StatusNoContent, nil)
	})
}
