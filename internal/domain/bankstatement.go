package domain

import "time"

type BankStatemant struct {
	Balance          BalanceStatement
	LastTransactions []LastTransaction
}

type LastTransaction struct {
	ID           int       `json:"id"`
	Value        int       `json:"value"`
	Kind         string    `json:"kind"`
	Description  string    `json:"description"`
	Payee        string    `json:"payee"`
	Completed_at time.Time `json:"completed_at"`
}
type BalanceStatement struct {
	Balance      int       `json:"balance"`
	Completed_at time.Time `json:"completed_at"`
	Limit        int       `json:"limit"`
}
