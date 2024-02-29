package domain

import (
	"time"
)

type Transaction struct {
	ID            int       `json:"id"`
	Client_Id     int       `json:"client_id"`
	Value         int       `json:"value"`
	Kind          string    `json:"kind"`
	Description   string    `json:"description"`
	Payor         string    `json:"payor"`
	PayeeUrubuKey string    `json:"payeeurubukey"`
	Completed_at  time.Time `json:"completed_at"`
}

type TransactionResponse struct {
	Value        int       `json:"value"`
	Kind         string    `json:"kind"`
	Payor        string    `json:"payor"`
	Payee        string    `json:"payee"`
	Balance      int       `json:"balance"`
	Completed_at time.Time `json:"completed_at"`
}
