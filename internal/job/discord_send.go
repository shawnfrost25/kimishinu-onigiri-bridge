package job

import (
	"context"
	"kimishinu/internal/discord"
	"net/http"

	"github.com/riverqueue/river"
)

// Think of this as Args, but here we store data that never changes, as example, the webhook, the logger is pretty much identical, etc...
type DiscordSendWorker struct {
	// WorkerDefaults expect inside '[]' an interface, the name is 'JobArgs' - the interface expects Kind() - exactly what we do, we pass xxxArgs as the struct (which hold data that we can use), while Kind() satisfy the contract
	river.WorkerDefaults[DiscordSendArgs]
	WebhookURL string
	HttpClient http.Client
}

// river.Job let us fetch the data it already stored when starting this, meaning that is "identifier" and "source" got filled with data, it will automatically get them - meaning you can use them here
func (w *DiscordSendWorker) Work(ctx context.Context, job *river.Job[DiscordSendArgs]) error {
	identifier := job.Args.Identifier
	source := job.Args.Source
	trace := job.Args.Trace

	bodyBytes, err := discord.ConvertToDiscordFormat(source, identifier)
	if err != nil {
		return err
	}

	d := discord.New(w.HttpClient, w.WebhookURL)

	err = d.Send(trace, bodyBytes)
	if err != nil {
		return err
	}

	return nil
}
