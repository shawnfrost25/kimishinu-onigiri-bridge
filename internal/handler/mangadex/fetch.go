package mangadex

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

const URI string = "https://api.mangadex.org/manga/%s/feed?limit=1&translatedLanguage%%5B%%5D=en&order%%5Bchapter%%5D=desc"

type Manga struct {
	ID          string
	Name        string
	HandlerName string
}

var Mangas = []Manga{
	{ID: "0c44cf39-3d6c-4472-815b-9e163613cfe9", Name: "Kimi ga Shinu Made Koi wo Shitai", HandlerName: "kimishinu_main"},
}

func (m *MangaDex) MDKimishinu(manga Manga) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.With(slog.String("op", "md_kimishinu_fetch"), slog.String("manga", manga.HandlerName))
		trace, ok := auth.GetTrace(r.Context(), w)
		if !ok {
			return
		}

		op := func(s string) string {
			return manga.HandlerName + "_" + s
		}

		finalURI := fmt.Sprintf(URI, manga.ID)

		reqM, err := http.NewRequest(http.MethodGet, finalURI, nil)
		if err != nil {
			err = failure.Wrap("mangadex.http.fetch.request.create", err)
			errlog.General(*logger, w, op("request_create"), failure.InternalGeneralError, err, trace)
			return
		}

		respM, err := http.DefaultClient.Do(reqM)
		if err != nil {
			err = failure.Wrap("mangadex.http.fetch.request.send", err)
			errlog.General(*logger, w, op("request_send"), failure.InternalGeneralError, err, trace)
			return
		}
		defer respM.Body.Close()

		if respM.StatusCode != http.StatusOK {
			err := failure.InternalGeneral(op("response"), fmt.Sprintf("unexpected status code: %d", respM.StatusCode))
			err = failure.Wrap("mangadex.http.fetch.response.status_code", err)
			errlog.General(*logger, w, op("response_status_code"), failure.InternalGeneralError, err, trace)
			return
		}

		var mangaResp MangaDexResp
		err = sonic.ConfigDefault.NewDecoder(respM.Body).Decode(&mangaResp)
		if err != nil {
			err = failure.Wrap("mangadex.http.fetch.response.decode", err)
			errlog.General(*logger, w, op("response_decode"), failure.InternalGeneralError, err, trace)
			return
		}

		if mangaResp.Result != "ok" {
			err := failure.InternalGeneral(op("response"), "result is not ok")
			err = failure.Wrap("mangadex.http.fetch.response.result_not_ok", err)
			errlog.General(*logger, w, op("response_result_not_ok"), failure.InternalGeneralError, err, trace)
			return
		}

		if len(mangaResp.Data) == 0 {
			err := failure.InternalGeneral(op("response"), "no data found in response")
			err = failure.Wrap("mangadex.http.fetch.response.no_data", err)
			errlog.General(*logger, w, op("response_no_data"), failure.NotFoundError, err, trace)
			return
		}

		var inserted int
		for _, data := range mangaResp.Data {
			logger.Debug("Processing manga data", slog.String("manga_id", data.ID))

			insertedRows, err := m.Queries.InsertMangaDexId(r.Context(), data.ID)
			if err != nil {
				err = failure.Wrap("mangadex.database.manga.id.deduplicate.check", err)
				errlog.Database(*logger, w, op("deduplicate_check"), trace, err)
				return
			}

			if insertedRows == 0 {
				logger.Debug("Manga ID already exists in database, skipping", slog.String("manga_id", data.ID))
				continue
			}

			bodyBytes, err := discord.ConvertToDiscordFormat("mangadex", data.ID)
			if err != nil {
				err = failure.Wrap(fmt.Sprintf("mangadex.json.%s.response.body.marshal", manga.HandlerName), err)
				errlog.Json(*logger, w, op("response_body_marshal"), trace, err)
				return
			}

			logger.Debug("Sending data to Discord", slog.String("manga_id", data.ID))
			d := discord.New(m.HttpClient, m.WebhookURL, *logger)

			err = d.Send(w, trace, bodyBytes)
			if err != nil {
				return
			}
			inserted += 1
		}

		if inserted == 0 {
			logger.Debug("No new mangas to send to Discord")
			lib.Pretty(w, http.StatusNoContent, nil)
			return
		}

		logger.Info("successfully send data to discord")
		lib.Pretty(w, http.StatusNoContent, nil)
	})
}
