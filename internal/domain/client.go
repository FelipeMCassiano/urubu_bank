package domain

type Client struct {
	ID       int      `json:"id"`
	Fullname string   `json:"fullname"`
	Birth    string   `json:"birth"`
	Limit    int      `json:"limit"`
	Balance  int      `json:"balance"`
	UrubuKey UrubuKey `json:"urubukey"`
}

type ClientConsult struct {
	Fullname string   `json:"fullname"`
	UrubuKey UrubuKey `json:"urubukey"`
}

type CreateCLient struct {
	Fullname string   `json:"fullname"`
	Birth    string   `json:"birth"`
	Limit    int      `json:"limit"`
	UrubuKey UrubuKey `json:"urubukey"`
}

type UrubuKey string
