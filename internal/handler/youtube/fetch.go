package youtube

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

type Channel struct {
	PlaylistId  string
	ChannelName string
	Keywords    []string
	HandlerName string
}

var keywords = []string{"きみが死ぬまで恋をしたい", "I Want To Love You Till Your Dying Day", "Kimi ga Shinu made Koi wo Shitai", "KimiShinu", "きみ死ぬ"}

var Channels = []Channel{
	{PlaylistId: "UU2QDgMJWRQHEJYyVX3Rt08w", ChannelName: "Ichijinsha PV", Keywords: keywords, HandlerName: "ichijinsha"},
	{PlaylistId: "UUyUhtF50BuUjr2jOhxF3IjQ", ChannelName: "ReoNa Official", Keywords: keywords, HandlerName: "reona"},
	{PlaylistId: "UUevPX0DXUiKmgH8o19NrE5w", ChannelName: "Infinite Channel", Keywords: keywords, HandlerName: "infinite"},
}

func match(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(text), strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

const URI = "https://www.googleapis.com/youtube/v3/playlistItems?part=snippet,contentDetails&maxResults=10&playlistId="

func (y *Youtube) FetchKimishinu(ch Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		polish := func(s string) string {
			return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
		}
		logger := slog.With(slog.String("op", "fetch_kimishinu"), slog.String("channel", polish(ch.ChannelName)))
		logger.Debug(fmt.Sprintf("Fetching %s from Youtube API", ch.ChannelName))
		trace, ok := auth.GetTrace(r.Context(), w)
		if !ok {
			return
		}

		op := func(reason string) string {
			return polish(ch.ChannelName) + reason
		}

		finalURI := URI + ch.PlaylistId

		reqI, err := http.NewRequest(http.MethodGet, finalURI, nil)
		if err != nil {
			err = failure.Wrap("youtube.http.fetch.request.create", err)
			errlog.Internal(*logger, w, op("_request_create"), failure.InternalGeneralError, err, trace)
			return
		}
		reqI.Header.Set("X-Goog-Api-Key", y.ApiKey)

		logger.Debug("Sending request to Youtube API", slog.String("url", finalURI))
		respI, err := y.HttpClient.Do(reqI)
		if err != nil {
			err = failure.Wrap("youtube.http.fetch.request.send", err)
			errlog.Internal(*logger, w, op("_request_send"), failure.InternalGeneralError, err, trace)
			return
		}
		defer respI.Body.Close()

		if respI.StatusCode != http.StatusOK {
			err = failure.InternalGeneral(op("_request_status_code"), fmt.Sprintf("unexpected status code: %d", respI.StatusCode))
			err = failure.Wrap("youtube.http.fetch.request.status_code", err)
			errlog.Internal(*logger, w, op("_request_status_code"), failure.InternalGeneralError, err, trace)
			return
		}

		var videoResp VideoResp
		logger.Debug("Decoding response from Youtube API")
		err = sonic.ConfigDefault.NewDecoder(respI.Body).Decode(&videoResp)
		if err != nil {
			err = failure.Wrap("youtube.http.fetch.response.decode", err)
			errlog.Internal(*logger, w, op("_response_decode"), failure.InternalGeneralError, err, trace)
			return
		}

		if len(videoResp.Items) == 0 {
			err = failure.InternalGeneral(op("_response"), "no items found in response")
			err = failure.Wrap("youtube.http.fetch.response.no_items", err)
			errlog.Internal(*logger, w, op("_response_no_items"), failure.InternalGeneralError, err, trace)
			return
		}

		logger.Debug("Processing items from Youtube API response", slog.Int("item_count", len(videoResp.Items)))
		for _, item := range videoResp.Items {
			if !match(item.Snippet.Title, ch.Keywords) {
				continue
			}

			bodyBytes, err := discord.ConvertToDiscordFormat("youtube", item.ContentDetails.VideoId)
			if err != nil {
				err = failure.Wrap(fmt.Sprintf("youtube.json.%s.response.body.marshal", polish(ch.ChannelName)), err)
				errlog.Json(*logger, w, op("_response_body_marshal"), trace, err)
				return
			}

			logger.Debug("Sending data to Discord", slog.String("video_id", item.ContentDetails.VideoId))
			d := discord.New(y.HttpClient, y.WebhookURL, *logger)

			err = d.Send(w, trace, bodyBytes)
			if err != nil {
				// discord.Send already logs the error, so we just return here
				return
			}
		}

		logger.Info("successfully sent data to discord")
		lib.Pretty(w, http.StatusOK, nil)
	}
}
