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

const URI = "https://www.googleapis.com/youtube/v3/playlistItems?part=snippet,contentDetails&playlistId=UU2QDgMJWRQHEJYyVX3Rt08w&maxResults=10&key="

func (y *Youtube) IchijinshaPV(w http.ResponseWriter, r *http.Request) {
	logger := slog.With(slog.String("op", "ichijinsha_pv_fetch_kimishinu"))
	logger.Debug("Fetching Ichijinsha PV from Youtube API")
	trace, ok := auth.GetTrace(r.Context(), w)
	if !ok {
		return
	}

	finalURI := URI + y.ApiKey

	reqI, err := http.NewRequest(http.MethodGet, finalURI, nil)
	if err != nil {
		err = failure.Wrap("youtube.http.fetch.request.create", err)
		errlog.Internal(*logger, w, "ichijinsha_pv_request_create", failure.InternalGeneralError, err, trace)
		return
	}

	logger.Debug("Sending request to Youtube API", slog.String("url", finalURI))
	respI, err := y.HttpClient.Do(reqI)
	if err != nil {
		err = failure.Wrap("youtube.http.fetch.request.send", err)
		errlog.Internal(*logger, w, "ichijinsha_pv_request_send", failure.InternalGeneralError, err, trace)
		return
	}
	defer respI.Body.Close()

	if respI.StatusCode != http.StatusOK {
		err = failure.InternalGeneral("ichijinsha_response", fmt.Sprintf("unexpected status code: %d", respI.StatusCode))
		err = failure.Wrap("youtube.http.fetch.request.status_code", err)
		errlog.Internal(*logger, w, "ichijinsha_pv_request_status_code", failure.InternalGeneralError, err, trace)
		return
	}

	var videoResp VideoResp
	logger.Debug("Decoding response from Youtube API")
	err = sonic.ConfigDefault.NewDecoder(respI.Body).Decode(&videoResp)
	if err != nil {
		err = failure.Wrap("youtube.http.fetch.response.decode", err)
		errlog.Internal(*logger, w, "ichijinsha_pv_response_decode", failure.InternalGeneralError, err, trace)
		return
	}

	if len(videoResp.Items) == 0 {
		err = failure.InternalGeneral("ichijinsha_response", "no items found in response")
		err = failure.Wrap("youtube.http.fetch.response.no_items", err)
		errlog.Internal(*logger, w, "ichijinsha_pv_response_no_items", failure.InternalGeneralError, err, trace)
		return
	}

	logger.Debug("Processing items from Youtube API response", slog.Int("item_count", len(videoResp.Items)))
	for _, item := range videoResp.Items {
		if !strings.Contains(item.Snippet.Title, "きみが死ぬまで恋をしたい") && !strings.Contains(item.Snippet.Description, "きみが死ぬまで恋をしたい") {
			continue
		}

		bodyBytes, err := YTDiscordConvert(&item)
		if err != nil {
			err = failure.Wrap("youtube.json.ichijinsha.response.body.marshal", err)
			errlog.Json(*logger, w, "ichijinsha_response_body", trace, err)
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
