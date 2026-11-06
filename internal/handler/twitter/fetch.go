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

		var TweetsResponse lib.TwitterResp
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

		pureIds, err := lib.Deduplicate(r.Context(), keywords, TweetsResponse.Data.Tweets, t.Queries)
		if err != nil {
			err = failure.Wrap("twitter.http.fetch.deduplicate", err)
			errlog.Database(*logger, w, op("deduplicate"), trace, err)
			return
		}
		if pureIds == nil {
			logger.Debug("No new tweets to send to Discord")
			lib.Pretty(w, http.StatusNoContent, nil)
			return
		}

		lib.Pretty(w, http.StatusNoContent, nil)

		go func(l []string) {
			for _, id := range l {
				gift := user.Username + "_" + id
				bodyBytes, err := discord.ConvertToDiscordFormat("twitter", gift)
				if err != nil {
					err = failure.Wrap(fmt.Sprintf("twitter.json.%s.response.body.marshal", user.Handlername), err)
					errlog.Json(*logger, w, op("response_body_marshal"), trace, err)
					return
				}

				logger.Debug("Sending data to Discord", slog.String("tweet_id", id))
				d := discord.New(t.HttpClient, t.WebhookURL, *logger)

				// Whatever, we will fake it until we make it. When I'll buy the premium version of Twitter API, I will fix this. For now, we will just send the latest tweet every time.
				err = d.Send(trace, bodyBytes)
				if err != nil {
					return
				}
			}
		}(pureIds)

		logger.Info("Successfully sent data to Discord")
	})
}
