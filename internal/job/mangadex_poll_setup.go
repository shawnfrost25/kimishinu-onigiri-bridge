package job

type MangadexPollArgs struct {
	Tier string
}

func (MangadexPollArgs) Kind() string {
	return "mangadex_poll"
}
