package domain

type Costumer struct {
	ID       int      `json:"id"`
	Fullname string   `json:"fullname"`
	Birth    string   `json:"birth"`
	Limit    int      `json:"limit"`
	Balance  int      `json:"balance"`
	Password string   `json:"password"`
	UrubuKey UrubuKey `json:"urubukey"`
}

type CostumerConsult struct {
	Fullname string   `json:"fullname"`
	UrubuKey UrubuKey `json:"urubukey"`
}

type CreateCostumer struct {
	Fullname string `json:"fullname"`
	Birth    string `json:"birth"`
	Limit    int    `json:"limit"`
	Password string `json:"password"`
}
type CreatedCostumer struct {
	ID       int      `json:"id"`
	Fullname string   `json:"fullname"`
	Limit    int      `json:"limit"`
	UrubuKey UrubuKey `json:"urubukey"`
}

type UrubuKey string
