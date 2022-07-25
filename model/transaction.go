package model

import "time"

type Transaction struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Debit     Account   `json:"debit"`
	Credit    Account   `json:"credit"`
	Amount    Amount    `json:"amount"`
	Details   string    `json:"details"`
}

type TransactionFilter struct {
	TimeAfter  *time.Time `json:"time_after,omitempty"`
	TimeBefore *time.Time `json:"time_before,omitempty"`
	Debit      *Account   `json:"debit,omitempty"`
	Credit     *Account   `json:"credit,omitempty"`
	AmountMin  *Amount    `json:"amount_min,omitempty"`
	AmountMax  *Amount    `json:"amount_max,omitempty"`
}
