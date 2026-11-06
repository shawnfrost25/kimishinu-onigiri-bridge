package mangadex

type Data struct {
	ID string `json:"id"`
}

type MangaDexResp struct {
	Result string `json:"result"`
	Data   []Data `json:"data"`
}
