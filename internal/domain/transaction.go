package domain

import (
	"time"
)

type TransactionDebit struct {
	ID            int       `json:"id"`
	Client_Id     int       `json:"client_id"`
	Value         int       `json:"value"`
	Kind          string    `json:"kind"`
	Description   string    `json:"description"`
	Payor         string    `json:"payor"`
	PayeeUrubuKey string    `json:"payeeurubukey"`
	Completed_at  time.Time `json:"completed_at"`
}

type TransactionCredit struct {
	ID           int       `json:"id"`
	Client_Id    int       `json:"client_id"`
	Value        int       `json:"value"`
	Kind         string    `json:"kind"`
	Description  string    `json:"description"`
	Completed_at time.Time `json:"completed_at"`
}

type TransactionResponseDebit struct {
	Value        int       `json:"value"`
	Kind         string    `json:"kind"`
	Description  string    `json:"description"`
	Payor        string    `json:"payor"`
	Payee        string    `json:"payee"`
	Balance      int       `json:"balance"`
	Completed_at time.Time `json:"completed_at"`
}

type TransactionResponseCredit struct {
	Newbalance   int       `json:"newbalance"`
	Completed_at time.Time `json:"completed_at"`
}
