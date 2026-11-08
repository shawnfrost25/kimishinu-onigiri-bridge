package job

import (
	"context"
	"fmt"
	db "kimishinu/internal/db/sqlc"
	"net/http"

	"kimishinu/internal/errlog"
	"kimishinu/internal/failure"
	"kimishinu/internal/lib"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/bytedance/sonic"
)

type MangadexPollWorker struct {
	river.WorkerDefaults[MangadexPollArgs]
	Pool        *pgxpool.Pool
	Queries     *db.Queries
	RiverClient *river.Client[pgx.Tx]
	HttpClient  http.Client
}

const MDURI string = "https://api.mangadex.org/manga/%s/feed?limit=1&translatedLanguage%%5B%%5D=en&order%%5Bchapter%%5D=desc"

type Manga struct {
	ID          string
	Name        string
	HandlerName string
	Tier        string
}

var Mangas = []Manga{
	{ID: "0c44cf39-3d6c-4472-815b-9e163613cfe9", Name: "Kimi ga Shinu Made Koi wo Shitai", HandlerName: "kimishinu_main", Tier: "medium"},
}

func (m *MangadexPollWorker) Work(ctx context.Context, job *river.Job[MangadexPollArgs]) error {
	currentTier := job.Args.Tier
	trace := "md_kimishinu_poll_" + fmt.Sprintf("%d", job.ID)

	for _, manga := range Mangas {
		if manga.Tier != currentTier {
			continue
		}

		err := m.MDKimishinu(manga, trace)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MangadexPollWorker) MDKimishinu(manga Manga, trace string) error {
	logger := slog.With(slog.String("op", "md_kimishinu_fetch"), slog.String("manga", manga.HandlerName))

	op := func(s string) string {
		return manga.HandlerName + "_" + s
	}

	ctx := context.Background()
	tx, err := m.Pool.Begin(ctx)
	if err != nil {
		err = failure.Wrap("mangadex.http.pool.transaction.begin", err)
		errlog.General(*logger, op("transaction_begin"), failure.InternalGeneralError, err)
		return err
	}
	defer tx.Rollback(ctx)

	qtx := m.Queries.WithTx(tx)

	finalURI := fmt.Sprintf(MDURI, manga.ID)

	reqM, err := http.NewRequest(http.MethodGet, finalURI, nil)
	if err != nil {
		err = failure.Wrap("mangadex.http.fetch.request.create", err)
		errlog.General(*logger, op("request_create"), failure.InternalGeneralError, err)
		return err
	}

	respM, err := m.HttpClient.Do(reqM)
	if err != nil {
		err = failure.Wrap("mangadex.http.fetch.request.send", err)
		errlog.General(*logger, op("request_send"), failure.InternalGeneralError, err)
		return err
	}
	defer respM.Body.Close()

	if respM.StatusCode != http.StatusOK {
		err := failure.InternalGeneral(op("response"), fmt.Sprintf("unexpected status code: %d", respM.StatusCode))
		err = failure.Wrap("mangadex.http.fetch.response.status_code", err)
		errlog.General(*logger, op("response_status_code"), failure.InternalGeneralError, err)
		return err
	}

	var mangaResp lib.MangaDexResp
	err = sonic.ConfigDefault.NewDecoder(respM.Body).Decode(&mangaResp)
	if err != nil {
		err = failure.Wrap("mangadex.http.fetch.response.decode", err)
		errlog.General(*logger, op("response_decode"), failure.InternalGeneralError, err)
		return err
	}

	if mangaResp.Result != "ok" {
		err := failure.InternalGeneral(op("response"), "result is not ok")
		err = failure.Wrap("mangadex.http.fetch.response.result_not_ok", err)
		errlog.General(*logger, op("response_result_not_ok"), failure.InternalGeneralError, err)
		return err
	}

	if len(mangaResp.Data) == 0 {
		err := failure.InternalGeneral(op("response"), "no data found in response")
		err = failure.Wrap("mangadex.http.fetch.response.no_data", err)
		errlog.Resource(*logger, op("response_no_data"), failure.NotFoundError, err)
		return err
	}

	pureIds, err := lib.Deduplicate(ctx, nil, mangaResp.Data, qtx)
	if err != nil {
		err = failure.Wrap("mangadex.http.fetch.deduplicate", err)
		errlog.Database(*logger, op("deduplicate"), err)
		return err
	}
	if pureIds == nil {
		logger.Debug("No new mangas to send to Discord")
		return nil
	}

	for _, id := range pureIds {
		_, err := m.RiverClient.InsertTx(ctx, tx, &DiscordSendArgs{
			Source:     "mangadex",
			Identifier: id,
			Trace:      trace,
		}, nil)
		if err != nil {
			err = failure.Wrap("mangadex.river.discord.send.insert", err)
			errlog.General(*logger, op("river_setup"), failure.InternalGeneralError, err)
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		err = failure.Wrap("mangadex.http.pool.transaction.commit", err)
		errlog.General(*logger, op("transaction_commit"), failure.InternalGeneralError, err)
		return err
	}

	logger.Info("successfully send data to discord")
	return nil
}
