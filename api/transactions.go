package main

import (
	"context"
	"time"

	"github.com/go-msvc/errors"
	"github.com/jansemmelink/money/db"
	"github.com/jansemmelink/money/model"
)

type TransactionLstRequest struct{}

type TransactionLstResponse struct{}

func txLst(ctx context.Context, req TransactionLstRequest) (*TransactionLstResponse, error) {
	return nil, errors.Errorf("NYI")
}

type TransactionAddRequest struct {
	Timestamp time.Time    `json:"timestamp"`
	Debit     string       `json:"debit"`
	Credit    string       `json:"credit"`
	Amount    model.Amount `json:"amount"`
}

func (req *TransactionAddRequest) Validate() error {
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}
	if req.Debit == "" {
		return errors.Errorf("missing debit")
	}
	if req.Credit == "" {
		return errors.Errorf("missing credit")
	}
	if req.Amount == 0 {
		return errors.Errorf("missing amount")
	}
	if req.Amount < 0 {
		return errors.Errorf("negative amount")
	}
	return nil
}

type TransactionAddResponse struct{}

func txAdd(ctx context.Context, req model.Transaction) (*model.Transaction, error) {
	tx, err := db.AddTransaction(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add transaction")
	}
	return tx, nil
}

type TransactionGetRequest struct{}

type TransactionGetResponse struct{}

func txGet(ctx context.Context, req TransactionGetRequest) (*TransactionGetResponse, error) {
	return nil, errors.Errorf("NYI")
}

type TransactionUpdRequest struct{}

type TransactionUpdResponse struct{}

func txUpd(ctx context.Context, req TransactionUpdRequest) (*TransactionUpdResponse, error) {
	return nil, errors.Errorf("NYI")
}

type TransactionDelRequest struct{}

type TransactionDelResponse struct{}

func txDel(ctx context.Context, req TransactionDelRequest) (*TransactionDelResponse, error) {
	return nil, errors.Errorf("NYI")
}
