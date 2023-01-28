package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/jansemmelink/money/db"
	"github.com/jansemmelink/money/model"
)

func main() {
	for {
		acc, err := selectAccount()
		if err != nil {
			panic(fmt.Sprintf("failed to select an account: %+v", err))
		}
		if acc == nil {
			os.Exit(0)
		}

		for {
			//show most recent transactions
			clear()
			fmt.Printf("Account: %+v\n", *acc)
			txs, err := db.GetTransactions(model.TransactionFilter{}, []string{}, 20)
			if err != nil {
				fmt.Printf("Failed to list transactions: %+v", err)
			}
			for i, tx := range txs {
				fmt.Printf("%2d) %+v %10.2f %30v %30v\n", i+1, tx.Timestamp.Format("2006-01-02"), tx.Amount, tx.Debit.Name, tx.Credit.Name)
			}
			var input string
			fmt.Scanf("%s", &input)
			i64, err := strconv.ParseInt(input, 10, 64)
			if err == nil && i64 >= 1 && int(i64) <= len(txs) {
				//selected a transaction
				tx := txs[i64-1]

				//show selected transaction
				clear()
				fmt.Printf("Transaction: %+v\n", tx.ID)
				fmt.Printf("\n")
				fmt.Printf("%+v %10.2f %30v %30v\n", tx.Timestamp.Format("2006-01-02"), tx.Amount, tx.Debit.Name, tx.Credit.Name)
				fmt.Printf("Details: %s\n", tx.Details)
				fmt.Printf("\n")
				var input string
				fmt.Scanf("%s", &input)

			} else {
				break //show accounts again
			}
		}
	}
}

func selectAccount() (*model.Account, error) {
	accounts := []model.Account{}
	if err := db.Db().Select(&accounts,
		"SELECT * FROM `accounts`",
	); err != nil {
		panic(fmt.Sprintf("failed to list accounts: %+v", err))
	} else {
		for _, acc := range accounts {
			fmt.Printf("ACC: %+v\n", acc)
		}
	}
	for {
		clear()
		fmt.Printf("Select an account:\n")
		for i := 0; i < len(accounts); i++ {
			fmt.Printf("%d) %s\n", i+1, accounts[i].Name)
		}
		var input string
		fmt.Scanf("%s", &input)
		if i64, err := strconv.ParseInt(input, 10, 64); err == nil && i64 > 0 && int(i64) <= len(accounts) {
			return &accounts[i64-1], nil
		}
	}
}

var clear func() //create a map for storing clear funcs

func init() {
	switch runtime.GOOS {
	case "windows":
		clear = func() {
			cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	default:
		clear = func() {
			cmd := exec.Command("clear") //Linux example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	}
}
