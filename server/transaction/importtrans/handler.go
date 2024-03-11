package importtrans

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wazupwiddat/postrack/server/config"
	"github.com/wazupwiddat/postrack/server/transaction"
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

const (
	csvTxDate = iota
	csvAction
	csvSymbol
	csvDescription
	csvQuantity
	csvPrice
	csvFees
	csvAmount
)

func ImprotUploadedFiles(db *gorm.DB, cfg *config.Config, u *user.User) {
	// Delete all user transactions first
	db.Unscoped().Delete(&transaction.Transaction{}, "user_id = ?", u.ID)

	files := loadTransactionFiles(cfg.Import.DownloadPath)
	for _, f := range files {
		filename := fmt.Sprintf("%s/%s", cfg.Import.DownloadPath, f.Name())
		log.Println("Files to be read: ", filename)
		csvFile, err := os.Open(filename)
		if err != nil {
			log.Println("Couldn't open the csv file", err)
		}

		// skip first line of file
		var accountName string
		// bio := bufio.NewReader(csvFile)
		// acct, _, _ := bio.ReadLine()

		// alias the account instead of what they provide
		if strings.Contains(filename, "XXX953") {
			accountName = "Brokerage"
		}
		if strings.Contains(filename, "XXX286") {
			accountName = "IRA"
		}
		// _, err = csvFile.Seek(int64(len(acct)), io.SeekStart)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		transactions := []transaction.Transaction{}
		reader := csv.NewReader(csvFile)
		// Header
		header, err := reader.Read()
		log.Println("headers", header)

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println(err)
				continue
			}

			tdate := record[csvTxDate]
			if strings.Contains(tdate, "Transactions") {
				continue
			}

			// write to DB
			t := transaction.Transaction{
				UserID:      u.ID,
				Account:     accountName,
				Date:        transactionDateFromDate(record[csvTxDate]),
				Action:      record[csvAction],
				Symbol:      record[csvSymbol],
				Description: record[csvDescription],
				Quantity:    safeStringToFloat(record[csvQuantity]),
				Price:       safeStringToFloat(record[csvPrice]),
				FeesComm:    safeStringToFloat(record[csvFees]),
				Amount:      safeStringToFloat(record[csvAmount]),
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

func loadTransactionFiles(directory string) []os.FileInfo {
	files := []os.FileInfo{}
	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		if info.IsDir() {
			return nil
		}
		log.Printf("File to be prepared: %s\n", info.Name())
		files = append(files, info)
		return nil
	})
	return files
}

func safeStringToFloat(str string) float64 {
	if str == "" {
		return 0.0
	}
	replacer := strings.NewReplacer("$", "", ",", "")
	s := replacer.Replace(str)
	c, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return c
}

// 05/24/2021 as of 05/21/2021
func transactionDateFromDate(date string) string {
	d := strings.Split(date, " as of ")
	if len(d) > 1 {
		return d[1]
	}
	return d[0]
}
