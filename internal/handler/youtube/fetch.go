package youtube

import (
	"context"
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

type Channel struct {
	PlaylistId  string
	ChannelName string
	Keywords    []string
	HandlerName string
}

var keywords = []string{"きみが死ぬまで恋をしたい", "iwanttoloveyoutillyourdyingday", "kimigashinumadekoiwoshitai", "kimishinu", "きみ死ぬ"}

var Channels = []Channel{
	{PlaylistId: "UU2QDgMJWRQHEJYyVX3Rt08w", ChannelName: "Ichijinsha PV", Keywords: keywords, HandlerName: "ichijinsha"},
	{PlaylistId: "UUyUhtF50BuUjr2jOhxF3IjQ", ChannelName: "ReoNa Official", Keywords: keywords, HandlerName: "reona"},
	{PlaylistId: "UUevPX0DXUiKmgH8o19NrE5w", ChannelName: "Infinite Channel", Keywords: keywords, HandlerName: "infinite"},
}

const URI = "https://www.googleapis.com/youtube/v3/playlistItems?part=snippet,contentDetails&maxResults=10&playlistId="

func (y *Youtube) YTKimishinu(ch Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		polish := func(s string) string {
			return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
		}
		logger := slog.With(slog.String("op", "yt_kimishinu_fetch"), slog.String("channel", polish(ch.ChannelName)))
		logger.Debug(fmt.Sprintf("Fetching %s from Youtube API", ch.ChannelName))
		trace, ok := auth.GetTrace(r.Context(), w)
		if !ok {
			return
		}

		op := func(reason string) string {
			return ch.HandlerName + "_" + reason
		}

		finalURI := URI + ch.PlaylistId

		reqI, err := http.NewRequest(http.MethodGet, finalURI, nil)
		if err != nil {
			err = failure.Wrap("youtube.http.fetch.request.create", err)
			errlog.Internal(*logger, w, op("request_create"), failure.InternalGeneralError, err, trace)
			return
		}
		reqI.Header.Set("Accept", "application/json")
		reqI.Header.Set("X-Goog-Api-Key", y.ApiKey)

		logger.Debug("Sending request to Youtube API", slog.String("url", finalURI))
		respI, err := y.HttpClient.Do(reqI)
		if err != nil {
			err = failure.Wrap("youtube.http.fetch.request.send", err)
			errlog.Internal(*logger, w, op("request_send"), failure.InternalGeneralError, err, trace)
			return
		}
		defer respI.Body.Close()

		if respI.StatusCode != http.StatusOK {
			err = failure.InternalGeneral(op("request_status_code"), fmt.Sprintf("unexpected status code: %d", respI.StatusCode))
			err = failure.Wrap("youtube.http.fetch.request.status_code", err)
			errlog.Internal(*logger, w, op("request_status_code"), failure.InternalGeneralError, err, trace)
			return
		}

		var videoResp VideoResp
		logger.Debug("Decoding response from Youtube API")
		err = sonic.ConfigDefault.NewDecoder(respI.Body).Decode(&videoResp)
		if err != nil {
			err = failure.Wrap("youtube.http.fetch.response.decode", err)
			errlog.Internal(*logger, w, op("response_decode"), failure.InternalGeneralError, err, trace)
			return
		}

		if len(videoResp.Items) == 0 {
			err = failure.InternalGeneral(op("response"), "no items found in response")
			err = failure.Wrap("youtube.http.fetch.response.no_items", err)
			errlog.Http(*logger, w, op("response_no_items"), failure.NotFoundError, err, trace)
			return
		}

		logger.Debug("Processing items from Youtube API response", slog.Int("item_count", len(videoResp.Items)))
		var inserted int
		for _, item := range videoResp.Items {
			if !lib.Match(item.Snippet.Title, keywords) && !lib.Match(item.Snippet.Description, keywords) {
				continue
			}

			insertedRows, err := y.Queries.InsertYoutubeId(context.Background(), item.ContentDetails.VideoId)
			if err != nil {
				err = failure.Wrap("youtube.database.video.id.deduplicate.check", err)
				errlog.Database(*logger, w, op("deduplicate_check"), trace, err)
				return
			}

			if insertedRows == 0 {
				logger.Debug("Video already exists in database, skipping", slog.String("video_id", item.ContentDetails.VideoId))
				continue
			}

			bodyBytes, err := discord.ConvertToDiscordFormat("youtube", item.ContentDetails.VideoId)
			if err != nil {
				err = failure.Wrap(fmt.Sprintf("youtube.json.%s.response.body.marshal", polish(ch.ChannelName)), err)
				errlog.Json(*logger, w, op("response_body_marshal"), trace, err)
				return
			}

			logger.Debug("Sending data to Discord", slog.String("video_id", item.ContentDetails.VideoId))
			d := discord.New(y.HttpClient, y.WebhookURL, *logger)

			err = d.Send(w, trace, bodyBytes)
			if err != nil {
				// discord.Send already logs the error, so we just return here
				return
			}
			inserted += 1
		}

		if inserted == 0 {
			logger.Debug("No new videos to send to Discord")
			lib.Pretty(w, http.StatusNoContent, nil)
			return
		}

		logger.Info("successfully sent data to discord")
		lib.Pretty(w, http.StatusNoContent, nil)
	}
}
