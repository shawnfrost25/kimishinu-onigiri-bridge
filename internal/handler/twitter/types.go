package twitter

type Tweet struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Text string `json:"text"`
}

type Data struct {
	Tweets []Tweet `json:"tweets"`
}

type TwitterResp struct {
	Data Data `json:"data"`
}
