package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jansemmelink/money/db"
	"github.com/jansemmelink/money/model"
)

//process standardbank cheque account statements in this format:
//
// Line    Text
//------------------------------------------------------------------
// 1       0,2645,BRANCH,0,,CENTURION LIFES,0,0
// 2       ,12319791,ACC-NO,0,,,0,0
// 3       ,0,OPEN,67009.75,OPEN BALANCE,,0,0
// .       HIST,20220127,,-980,IB-BETALING NA,AHS,00377,0
// .       HIST,20220127,,-1000.72,VERSEKERINGSPREMIE,OUTSURANCE OT11259326   SN6099,6021,0
// .       ...
// .       HIST,20220725,,-35,TJEKKAART-AANKOOP,PLANT ELITE N 5222*7143 21 JUL,6076,0
// N       ,0,CLOSE,16064.65,CLOSE BALANCE,,0,0
//------------------------------------------------------------------

type Statement struct {
	Branch       Branch
	Account      Account
	Balances     []model.Amount
	Transactions []Transaction
}
type Branch struct {
	Code string
	Name string
}

type Account struct {
	Number string
}

type Transaction struct {
	Row         int
	Date        time.Time
	Amount      model.Amount
	IsFee       bool
	What        string
	Reference   string
	Code        string
	BankAccount string
}

func (tx Transaction) Summary() string {
	return fmt.Sprintf("%s|%v|%s|%s|%s", tx.Date.Format("20060102"), tx.Amount, tx.What, tx.Reference, tx.BankAccount)
}

func main() {
	bankAccountNamePtr := flag.String("acc", "", "Name in this system for the account that represents this bank account")
	bankFeesAccountNamePtr := flag.String("fees", "", "Name in this system for account that represents fees charged by this bank account")
	filenamePtr := flag.String("f", "", "File to import")
	flag.Parse()
	if *filenamePtr == "" {
		fmt.Fprintf(os.Stderr, "ERROR: Missing option -file <filename>\n")
		flag.Usage()
	}

	f, err := os.Open(*filenamePtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot open file %s: %+v\n", *filenamePtr, err)
		flag.Usage()
	}
	defer f.Close()

	r := csv.NewReader(f)
	row := 0
	localLocation := time.Now().Location()

	s := Statement{
		Balances:     []model.Amount{0, 0},
		Transactions: []Transaction{},
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(fmt.Sprintf("failed to read: %+v\n", err))
		}

		row++
		if row == 1 {
			//header row, example: "0,2645,BRANCH,0,,CENTURION LIFES,0,0"
			//columns:
			//0		"0"
			//1		<Numeric Branch code>
			//2		"BRANCH"
			//3		"0"
			//4		""
			//5		<Branch name>
			//6		"0"
			//7		"0"
			if len(record) != 8 {
				panic(fmt.Sprintf("row %d has %d columns, expected 8", row, len(record)))
			}
			s.Branch.Code = record[1]
			s.Branch.Name = record[5]
			continue
		}

		if row == 2 {
			//header row:
			//,12319791,ACC-NO,0,,,0,0
			//columns:
			//0	""
			//1	<account number>
			//2	"ACC-NO"
			//3	"0"
			//4	""
			//5	""
			//6	"0"
			//7	"0"
			if len(record) != 8 {
				panic(fmt.Sprintf("row %d has %d columns, expected 8", row, len(record)))
			}
			s.Account.Number = record[1]
			continue
		}

		if row == 3 {
			//header row:
			//,0,OPEN,67009.75,OPEN BALANCE,,0,0
			//columns:
			//0	""
			//1	"0"
			//2	"OPEN"
			//3	<opening balance>
			//4	"OPEN BALANCE"
			//5	""
			//6	"0"
			//7 "0"
			if len(record) != 8 {
				panic(fmt.Sprintf("row %d has %d columns, expected 8", row, len(record)))
			}
			if record[2] != "OPEN" {
				panic(fmt.Sprintf("row %d col 3 \"%s\" != \"OPEN\"", row, record[2]))
			}
			if record[4] != "OPEN BALANCE" {
				panic(fmt.Sprintf("row %d col 5 \"%s\" != \"OPEN BALANCE\"", row, record[4]))
			}
			var a model.Amount
			if err := a.Parse(record[3]); err != nil {
				panic(fmt.Sprintf("row %d col 3 \"%s\" cannot parse as amount: %+v", row, record[3], err))
			}
			s.Balances[0] = a
			continue
		}

		if len(record) >= 2 && record[2] == "CLOSE" {
			//last row: closing balance
			//,0,CLOSE,16064.65,CLOSE BALANCE,,0,0
			//columns:
			//0	""
			//1	"0"
			//2	"CLOSE"
			//3	<closing balance>
			//4	"CLOSE BALANCE"
			//5	""
			//6	"0"
			//7 "0"
			if len(record) != 8 {
				panic(fmt.Sprintf("row %d has %d columns, expected 8", row, len(record)))
			}
			if record[2] != "CLOSE" {
				panic(fmt.Sprintf("row %d col 3 \"%s\" != \"CLOSE\"", row, record[2]))
			}
			if record[4] != "CLOSE BALANCE" {
				panic(fmt.Sprintf("row %d col 5 \"%s\" != \"CLOSE BALANCE\"", row, record[4]))
			}
			var a model.Amount
			if err := a.Parse(record[3]); err != nil {
				panic(fmt.Sprintf("row %d col 3 \"%s\" cannot parse as amount: %+v", row, record[3], err))
			}
			s.Balances[1] = a
			continue
		}

		//transaction rows:
		// .       HIST,20220127,,-1000.72,VERSEKERINGSPREMIE,OUTSURANCE OT11259326   SN6099,6021,0
		//columns:
		//0	"HIST"
		//1 <date> CCYYMMDD
		//2 "##" when this is banking fee paid to the bank
		//3 <amount>
		//4 <what>
		//5 <reference>
		//6 <code>
		//7 "0"
		if record[0] != "HIST" {
			panic(fmt.Sprintf("row %d col 1 \"%s\" != \"HIST\"", row, record[0]))
		}
		if record[7] != "0" {
			panic(fmt.Sprintf("row %d col 8 \"%s\" != \"HIST\"", row, record[0]))
		}
		tx := Transaction{
			Row:         row,
			BankAccount: s.Account.Number,
		}
		tx.Date, err = time.ParseInLocation("20060102", record[1], localLocation)
		if err != nil {
			panic(fmt.Sprintf("row %d col 2 \"%s\" invalid date, expecting CCYYMMDD", row, record[1]))
		}
		switch record[2] {
		case "":
		case "##":
			tx.IsFee = true
		default:
			panic(fmt.Sprintf("row %d col 3 \"%s\" expected blank or ## only", row, record[2]))
		}
		if err := tx.Amount.Parse(record[3]); err != nil {
			panic(fmt.Sprintf("row %d col 4 \"%s\" invalid amount", row, record[3]))
		}
		if math.Abs(float64(tx.Amount)) < 0.01 {
			panic(fmt.Sprintf("row %d col 4 \"%s\" zero amount", row, record[3]))
		}
		if record[4] == "" {
			panic(fmt.Sprintf("row %d col 5 \"%s\" missing what value", row, record[4]))
		}
		tx.What = record[4]
		tx.Reference = record[5] //allowed to be empty
		tx.Code = record[6]

		s.Transactions = append(s.Transactions, tx)
	} //for each line

	diff := model.Amount(0)
	for _, tx := range s.Transactions {
		diff += tx.Amount
	}

	fmt.Printf("%d transactions\n", len(s.Transactions))
	fmt.Printf("Open:  %10.2f\n", s.Balances[0])
	fmt.Printf("Diff:  %10.2f\n", diff)
	fmt.Printf("Close: %10.2f\n", s.Balances[1])

	//compare to nearest cent only
	if fmt.Sprintf("%.2f", (s.Balances[0]+diff)) != fmt.Sprintf("%.2f", s.Balances[1]) {
		panic(fmt.Sprintf("Does not add up: open(%v) + diff(%v) = %v != close(%v)", s.Balances[0], diff, s.Balances[0]+diff, s.Balances[1]))
	}

	//get bank account
	if *bankAccountNamePtr == "" {
		*bankAccountNamePtr = fmt.Sprintf("bank account (acc: %s)", s.Account.Number)
	}
	thisBankAccount := getAccount(*bankAccountNamePtr, model.AccountTypeAsset)

	//get fees account
	if *bankFeesAccountNamePtr == "" {
		*bankFeesAccountNamePtr = fmt.Sprintf("bank fees (acc: %s)", s.Account.Number)
	}
	bankFeesAccount := getAccount(*bankFeesAccountNamePtr, model.AccountTypeExpense)

	//unknown expenses and incomes used for everything else
	unknownExpense := getAccount("unknown expenses", model.AccountTypeExpense)
	unknownIncome := getAccount("unknown income", model.AccountTypeEquity)

	//import into db - rejecting duplicates
	importID := uuid.New().String()
	importedCount := 0
	duplicates := []Transaction{}
	for _, tx := range s.Transactions {
		var otherAccountID = ""
		if tx.IsFee {
			otherAccountID = bankFeesAccount.ID
		}
		values := map[string]interface{}{
			"ts":        db.SqlDate(tx.Date),
			"amount":    float64(tx.Amount),
			"details":   tx.What,
			"reference": tx.Reference,
			"summary":   tx.Summary(),
			"import_id": importID,
		}
		if tx.Amount > 0 {
			if otherAccountID == "" {
				otherAccountID = unknownIncome.ID
			}
			values["ct_acc_id"] = otherAccountID
			values["dt_acc_id"] = thisBankAccount.ID
		} else {
			if otherAccountID == "" {
				otherAccountID = unknownExpense.ID
			}
			values["dt_acc_id"] = otherAccountID
			values["ct_acc_id"] = thisBankAccount.ID
		}

		if _, err := db.Db().NamedExec(
			"INSERT INTO `period_transactions`"+
				" SET `timestamp`=:ts,`dt_account_id`=:dt_acc_id,`ct_account_id`=:ct_acc_id,`amount`=:amount,`details`=:details,`summary`=:summary,`import_id`=:import_id",
			values,
		); err != nil {
			if mysqlError, ok := err.(*mysql.MySQLError); ok {
				switch mysqlError.Number {
				case 1062:
					//duplicate key
					duplicates = append(duplicates, tx)
				default:
					panic(fmt.Sprintf("failed to insert transation %+v: %+v", tx, err))
				}
			}
		} else {
			importedCount++
		}
	}

	//insert opening and closing balance
	if len(s.Transactions) > 0 {
		for index, bal := range s.Balances {
			var ts time.Time
			if index == 0 {
				//opening balance is (at the end of) the day before first transaction in this statement
				ts = s.Transactions[0].Date
			} else {
				//closing balance is (at the end of) the day of the last transaction
				ts = s.Transactions[len(s.Transactions)-1].Date
			}

			if _, err := db.Db().NamedExec(
				"INSERT INTO `account_expected_balances` SET account_id=:account_id,timestamp=:ts,expected_balance=:exp_bal,import_id=:import_id",
				map[string]interface{}{
					"account_id": thisBankAccount.ID,
					"ts":         ts,
					"exp_bal":    bal,
					"import_id":  importID,
				},
			); err != nil {
				//undo the import
				if _, delErr := db.Db().Exec("DELETE from `period_transactions` WHERE `import_id`=?", importID); delErr != nil {
					panic(fmt.Sprintf("failed to delete transactions with import id after balance entries failed \"%s\": %+v", importID, delErr))
				}
				fmt.Printf("Note: cleaned up by deleting imported records\n")
				fmt.Printf("ERROR: failed to insert expected balance: %+v\n", err)
				os.Exit(1)
			}
		}
	}

	fmt.Printf("Imported %10d new transactions\n", importedCount)
	fmt.Printf("Ignored  %10d duplicate transactions\n", len(duplicates))
	for _, dup := range duplicates {
		fmt.Printf("  duplicate: row %5d: %s\n", dup.Row, dup.Summary())
	}

	if len(duplicates) > 0 && importedCount > 0 {
		fmt.Printf("==================================================================================\n")
		fmt.Printf("With some records imported and some that causes duplicates, you should consider to\n")
		fmt.Printf("delete the imported records, update the source file and then try again.\n")
		fmt.Printf("==================================================================================\n")
		answer := ""
		for answer == "" {
			fmt.Printf("Do you want to keep or delete the %d imported records now (k=keep or d=delete)?", importedCount)
			fmt.Scanf("%s", &answer)
			switch answer {
			case "k":
			case "d":
			default:
				fmt.Printf("Answer only k to keep or d to delete the records that were imported.\n")
				answer = ""
			}
		}
		if answer == "d" {
			if _, err := db.Db().Exec("DELETE from `period_transactions` WHERE `import_id`=?", importID); err != nil {
				panic(fmt.Sprintf("failed to delete transactions with import id \"%s\": %+v", importID, err))
			}
			fmt.Printf("Deleted import ID \"%s\"\n", importID)
			os.Exit(1)
		}
	}
}

func getAccount(name string, accType model.AccountType) *model.Account {
	acc, err := db.GetAccountByName(name)
	if err == nil {
		if acc.Type != accType {
			panic(fmt.Sprintf("existing account(%s) has type %s != %s", name, acc.Type, accType))
		}
		return acc
	}

	fmt.Printf("account (%s) not found: %+v\n", name, err)
	var answer string
	for answer == "" {
		fmt.Printf("Can I create the account? (y/n) [y]")
		fmt.Scanf("%s", &answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "" {
			answer = "y"
		}
		switch answer {
		case "y":
		case "n":
			fmt.Fprintf(os.Stderr, "STATEMENT NOT IMPORTED\n")
			os.Exit(1)
		default:
			answer = "" //ask again
			fmt.Printf("Please answer only y or n\n")
		}
	}

	acc, err = db.AddAccount(model.Account{
		ID:   uuid.New().String(),
		Name: name,
		Type: accType,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create account: %+v", err))
	}
	return acc
} //getAccount()
