package main

import (
	"context"

	"github.com/go-msvc/errors"
	"github.com/jansemmelink/money/db"
	"github.com/jansemmelink/money/model"
)

type AccountsResponse struct {
	Accounts []model.Account `json:"accounts"`
}

func accLst(ctx context.Context, filter model.AccountFilter) (*AccountsResponse, error) {
	acc, err := db.GetAccounts(filter, []string{"name"}, 50)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get accounts")
	}
	return &AccountsResponse{
		Accounts: acc,
	}, nil
}

type AccountAddRequest struct {
	Name string            `json:"name"`
	Type model.AccountType `json:"type"`
}

func (req *AccountAddRequest) Validate() error {
	if req.Name == "" {
		return errors.Errorf("missing name")
	}
	// if err := req.Type.Validate(); err != nil {
	// 	return errors.Wrapf(err, "invalid type")
	// }
	return nil
}

func accAdd(ctx context.Context, req model.Account) (*model.Account, error) {
	acc, err := db.AddAccount(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add accounts")
	}
	return acc, nil
}

type AccountGetRequest struct{}

type AccountGetResponse struct{}

func accGet(ctx context.Context, req AccountGetRequest) (*AccountGetResponse, error) {
	return nil, errors.Errorf("NYI")
}

type AccountUpdRequest struct{}

type AccountUpdResponse struct{}

func accUpd(ctx context.Context, req AccountUpdRequest) (*AccountUpdResponse, error) {
	return nil, errors.Errorf("NYI")
}

type AccountDelRequest struct{}

type AccountDelResponse struct{}

func accDel(ctx context.Context, req AccountDelRequest) (*AccountDelResponse, error) {
	return nil, errors.Errorf("NYI")
}
