package job

type TwitterPollArgs struct {
	Tier string
}

func (TwitterPollArgs) Kind() string {
	return "twitter_poll"
}
