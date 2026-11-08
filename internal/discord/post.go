package discord

import (
	"bytes"
	"kimishinu/internal/failure"
	"net/http"
)

func (d *Discord) Send(trace string, body []byte) error {
	b := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, d.WebhookURL, b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if 200 >= resp.StatusCode || resp.StatusCode >= 300 {
		err = failure.InternalGeneral("discord_response", "unexpected status code: "+resp.Status)
		return err
	}

	return nil
}
