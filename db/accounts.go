package db

import (
	"github.com/go-msvc/errors"
	"github.com/google/uuid"
	"github.com/jansemmelink/money/model"
)

type Account struct {
	ID   string            `db:"id"`
	Name string            `db:"name"`
	Type model.AccountType `db:"type"`
}

func (dbAcc Account) Model() model.Account {
	return model.Account{
		ID:   dbAcc.ID,
		Name: dbAcc.Name,
		Type: dbAcc.Type,
	}
}

func GetAccounts(filter model.AccountFilter, sort []string, limit int) ([]model.Account, error) {
	log.Debugf("GetAccounts(filter: %+v, sort: %+v, limit: %d)", filter, sort, limit)
	var rows []Account

	query := "SELECT `id`,`name`,`type` FROM `accounts`"
	args := []interface{}{}
	if filter.Name != "" {
		query += " WHERE `name` like ?"
		args = append(args, "%"+filter.Name+"%")
	}

	query += " ORDER BY `name`"

	if limit <= 0 {
		limit = 10
	}
	query += " LIMIT ?"
	args = append(args, limit)

	if err := db.Select(&rows, query, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to query list of accounts")
	}
	accounts := make([]model.Account, len(rows))
	for i, row := range rows {
		accounts[i] = row.Model()
	}
	return accounts, nil
}

func AddAccount(newAccount model.Account) (*model.Account, error) {
	if newAccount.Name == "" {
		return nil, errors.Errorf("missing name")
	}

	var account model.Account
	account = newAccount
	account.ID = uuid.New().String()
	if _, err := db.NamedExec(
		"INSERT INTO accounts SET id=:id,name=:name,type=:type",
		map[string]interface{}{
			"id":   account.ID,
			"name": account.Name,
			"type": account.Type,
		},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to create account")
	}
	return &account, nil
}

func GetAccount(id string) (*model.Account, error) {
	var account model.Account
	if err := NamedGet(
		&account,
		"SELECT id,name,type from accounts where id=:id",
		map[string]interface{}{
			"id": id,
		},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to get account")
	}
	return &account, nil
}
