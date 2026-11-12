package job

// Since river relies on a interface, we need to satisfy the contract, so we create an useful strurcture and Kind()
// By declaring this, we will declare "Kind" too - here we store data that changes and stuff (which River must absolutely store)
type DiscordSendArgs struct {
	Source     string
	Identifier string
}

// We use Kind() to satisfy this interface
/*
type JobArgs interface {
	Kind() string
*/
func (DiscordSendArgs) Kind() string {
	return "discord_send"
}
