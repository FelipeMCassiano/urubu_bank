package domain

import "time"

type BankStatemant struct {
	Balance          int       `json:"balance"`
	Completed_at     time.Time `json:"completed_at"`
	Limit            int       `json:"limit"`
	LastTransactions []LastTransaction
}

type LastTransaction struct {
	ID           int       `json:"id"`
	Value        int       `json:"value"`
	Kind         string    `json:"kind"`
	Description  string    `json:"description"`
	Payor        string    `json:"payor"`
	Payee        string    `json:"payee"`
	Completed_at time.Time `json:"completed_at"`
}
