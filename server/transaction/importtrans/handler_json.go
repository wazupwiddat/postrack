package importtrans

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/wazupwiddat/postrack/server/config"
	"github.com/wazupwiddat/postrack/server/transaction"
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

type JSONTransactions struct {
	FromDate                string `json:"FromDate"`
	ToDate                  string `json:"ToDate"`
	TotalTransactionsAmount string `json:"TotalTransactionsAmount"`
	BrokerageTransactions   []struct {
		Date        string `json:"Date"`
		Action      string `json:"Action"`
		Symbol      string `json:"Symbol"`
		Description string `json:"Description"`
		Quantity    string `json:"Quantity"`
		Price       string `json:"Price"`
		FeesComm    string `json:"Fees & Comm"`
		Amount      string `json:"Amount"`
	} `json:"BrokerageTransactions"`
}

func ImportUploadedJSONFiles(db *gorm.DB, cfg *config.Config, u *user.User) {
	// Delete all user transactions first
	db.Unscoped().Delete(&transaction.Transaction{}, "user_id = ?", u.ID)

	files := loadTransactionFiles(cfg.Import.DownloadPath)
	for _, f := range files {
		filename := fmt.Sprintf("%s/%s", cfg.Import.DownloadPath, f.Name())
		log.Println("JSON Files to be read: ", filename)

		jsonFile, err := os.Open(filename)
		if err != nil {
			log.Println("Couldn't open the json file", err)
		}

		byteValue, _ := io.ReadAll(jsonFile)
		var jsonTransactions JSONTransactions
		json.Unmarshal(byteValue, &jsonTransactions)

		var accountName string

		// alias the account instead of what they provide
		if strings.Contains(filename, "XXX953") {
			accountName = "Brokerage"
		}
		if strings.Contains(filename, "XXX286") {
			accountName = "IRA"
		}

		transactions := []transaction.Transaction{}
		for _, jt := range jsonTransactions.BrokerageTransactions {
			// create DB record add to transaction list
			t := transaction.Transaction{
				UserID:      u.ID,
				Account:     accountName,
				Date:        transactionDateFromDate(jt.Date),
				Action:      jt.Action,
				Symbol:      jt.Symbol,
				Description: jt.Description,
				Quantity:    safeStringToFloat(jt.Quantity),
				Price:       safeStringToFloat(jt.Price),
				FeesComm:    safeStringToFloat(jt.FeesComm),
				Amount:      safeStringToFloat(jt.Amount),
			}

			transactions = append(transactions, t)
		}
		if err := transaction.CreateMany(db, transactions); err != nil {
			log.Println(err)
		}
		if e := os.Remove(filename); e != nil {
			log.Println(err)
		}
	}
}
