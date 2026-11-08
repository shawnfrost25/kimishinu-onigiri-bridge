package job

// We pass the Tier from main.go to the job, which later will get passed to the worker, which will then filter the channels based on the Tier
type YoutubePollArgs struct {
	Tier string
}

func (YoutubePollArgs) Kind() string {
	return "youtube_poll"
}
