package discord

import (
	"bytes"
	"kimishinu/internal/errlog"
	"kimishinu/internal/failure"
	"log/slog"
	"net/http"
)

func (d *Discord) Send(w http.ResponseWriter, trace string, body []byte) error {
	logger := d.Logger.With(slog.String("op", "discord_post_to_webhook"))
	b := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, d.WebhookURL, b)
	if err != nil {
		err = failure.Wrap("discord.http.post.request.create", err)
		errlog.Internal(*logger, w, "discord_request_create", failure.InternalGeneralError, err, trace)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.HttpClient.Do(req)
	if err != nil {
		err = failure.Wrap("discord.http.post.request.send", err)
		errlog.Internal(*logger, w, "discord_request_send", failure.InternalGeneralError, err, trace)
		return err
	}
	defer resp.Body.Close()

	if 200 >= resp.StatusCode || resp.StatusCode >= 300 {
		err = failure.InternalGeneral("discord_response", "unexpected status code: "+resp.Status)
		err = failure.Wrap("discord.http.post.request.status_code", err)
		errlog.Internal(*logger, w, "discord_request_status_code", failure.InternalGeneralError, err, trace)
		return err
	}

	return nil
}
