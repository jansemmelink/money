package model

import (
	"strings"

	"github.com/go-msvc/errors"
)

//Account types are one of the following
//Assets and expenses increase when you debit the accounts and decrease when you credit them.
//Liabilities, equity, and revenue increase when you credit the accounts and decrease when you debit them.

type AccountType int

const (
	AccountTypeAsset AccountType = iota
	AccountTypeLiability
	AccountTypeExpense
	AccountTypeRevenue //=income
	AccountTypeEquity
)

var accountTypeStringValues = map[AccountType]string{
	AccountTypeAsset:     "asset",
	AccountTypeLiability: "liability",
	AccountTypeExpense:   "expense",
	AccountTypeRevenue:   "revenue",
	AccountTypeEquity:    "quity",
}

var accountTypeStringToValue map[string]AccountType

func init() {
	accountTypeStringToValue = map[string]AccountType{}
	for t, s := range accountTypeStringValues {
		accountTypeStringToValue[s] = t
	}
}

func (accType AccountType) String() string {
	if s, ok := accountTypeStringValues[accType]; ok {
		return s
	}
	return ""
}

func (accType AccountType) MarshalJSON() ([]byte, error) {
	if s, ok := accountTypeStringValues[accType]; ok {
		return []byte("\"" + s + "\""), nil
	}
	return nil, errors.Errorf("invalid account type %v", accType)
}

func (accType *AccountType) UnmarshalJSON(value []byte) error {
	s := string(value)
	if !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
		return errors.Errorf("AccountType with unquoted value")
	}
	s = s[1 : len(s)-1]
	if t, ok := accountTypeStringToValue[s]; !ok {
		return errors.Errorf("unknown AccountType(%s)", s)
	} else {
		*accType = t
	}
	return nil
}
