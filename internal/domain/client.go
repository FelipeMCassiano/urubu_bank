package domain

type Client struct {
	ID       int    `json:"id"`
	Fullname string `json:"fullname"`
	Birth    string `json:"birth"`
	Limit    int    `json:"limit"`
	Balance  int    `json:"balance"`
	UrubuKey int    `json:"urubukey"`
}

type ClientConsult struct {
	Fullname string `json:"fullname"`
	UrubuKey int    `json:"urubukey"`
}
