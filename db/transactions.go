package db

import (
	"fmt"
	"time"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
	"github.com/jansemmelink/money/model"
)

type transaction struct {
	ID            string       `db:"id"`
	Timestamp     SqlTime      `db:"timestamp"`
	DtAccountID   string       `db:"dt_account_id"`
	DtAccountName string       `db:"dt_account_name"`
	CtAccountID   string       `db:"ct_account_id"`
	CtAccountName string       `db:"ct_account_name"`
	Amount        model.Amount `db:"amount"`
	Details       string       `db:"details"`
	Summary       string       `db:"summary"`
}

func (dbTx transaction) Model() model.Transaction {
	return model.Transaction{
		ID:        dbTx.ID,
		Timestamp: time.Time(dbTx.Timestamp),
		Debit: model.Account{
			ID:   dbTx.DtAccountID,
			Name: dbTx.DtAccountName,
		},
		Credit: model.Account{
			ID:   dbTx.CtAccountID,
			Name: dbTx.CtAccountName,
		},
		Amount:  dbTx.Amount,
		Details: dbTx.Details,
	}
}

func GetTransactions(filter model.TransactionFilter, sort []string, limit int) ([]model.Transaction, error) {
	log.Debugf("GetTransactions(filter: %+v, sort: %+v, limit: %d)", filter, sort, limit)

	query := "SELECT tx.id,tx.timestamp,tx.amount,tx.dt_account_id,dtacc.name AS dt_account_name,tx.ct_account_id,ctacc.name AS ct_account_name,tx.details FROM period_transactions AS tx" +
		" JOIN accounts AS dtacc ON tx.dt_account_id=dtacc.id" +
		" JOIN accounts AS ctacc ON tx.ct_account_id=ctacc.id"
	args := []interface{}{}

	whereClause := ""
	if filter.TimeAfter != nil {
		whereClause += " AND `timestamp` > ?"
		args = append(args, filter.TimeAfter)
	}
	if filter.TimeBefore != nil {
		whereClause += " AND `timestamp` < ?"
		args = append(args, filter.TimeBefore)
	}
	if whereClause != "" {
		query += " WHERE " + whereClause[5:] //5: to skip over first " AND "
	}

	query += " ORDER BY `timestamp`"

	if limit <= 0 {
		limit = 10
	}
	query += " LIMIT ?"
	args = append(args, limit)

	var rows []transaction
	if err := db.Select(&rows, query, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to query list of transactions")
	}
	transactions := make([]model.Transaction, len(rows))
	for i, row := range rows {
		transactions[i] = row.Model()
	}
	return transactions, nil
}

func AddTransaction(tx model.Transaction) (*model.Transaction, error) {
	if tx.Amount <= 0 {
		return nil, errors.Errorf("negative or zero amount")
	}

	if tx.Timestamp.IsZero() {
		tx.Timestamp = time.Now()
	}

	//get dt and ct account details so can be included in the summary for future searching of transactions
	dtAcc, err := GetAccount(tx.Debit.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get debit account details")
	}
	tx.Debit = *dtAcc

	ctAcc, err := GetAccount(tx.Credit.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get credit account details")
	}
	tx.Credit = *ctAcc

	id := uuid.New().String()
	if _, err := db.NamedExec("INSERT INTO period_transactions"+
		" SET id=:id,timestamp=:ts,dt_account_id=:dt,ct_account_id=:ct,amount=:amount,details=:details,summary=:summary",
		map[string]interface{}{
			"id":      id,
			"ts":      SqlTime(tx.Timestamp),
			"dt":      tx.Debit.ID,
			"ct":      tx.Credit.ID,
			"amount":  tx.Amount,
			"details": tx.Details,
			"summary": fmt.Sprintf("%s;%s;%s;%v;%s", tx.Timestamp, dtAcc.Name, ctAcc.Name, tx.Amount, tx.Details),
		}); err != nil {
		return nil, errors.Wrapf(err, "failed to create transaction")
	}
	tx.ID = id
	return &tx, nil
}

func GetTransaction(id string) (*model.Transaction, error) {
	var dbTx transaction
	if err := NamedGet(
		&dbTx,
		"SELECT tx.id,tx.timestamp,tx.amount,tx.dt_account_id,dtacc.name AS dt_account_name,tx.ct_account_id,ctacc.name AS ct_account_name,tx.details FROM period_transactions AS tx"+
			" JOIN accounts AS dtacc ON tx.dt_account_id=dtacc.id"+
			" JOIN accounts AS ctacc ON tx.ct_account_id=ctacc.id"+
			" where tx.id=:id",
		map[string]interface{}{
			"id": id,
		},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to get account")
	}
	modelTx := dbTx.Model()
	return &modelTx, nil
}

func GetAccountByName(name string) (*model.Account, error) {
	var account model.Account
	if err := NamedGet(
		&account,
		"SELECT id,name,type from accounts where name=:name",
		map[string]interface{}{
			"name": name,
		},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to get account")
	}
	return &account, nil
}
