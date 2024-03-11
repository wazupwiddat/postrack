package summary

import (
	"sort"
	"strconv"
	"time"

	"github.com/wazupwiddat/postrack/server/transaction"
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

type Request struct {
	User *user.User
}

type OpenSummary struct {
	Account string
	Value   float64
}

type ClosedSummaryByMonth struct {
	Month string
	Value []float64
}

type ClosedSummaryByYear struct {
	Year  string
	Value []float64
}

type Response struct {
	Accounts            []string               `json:"accounts"`
	OpenShorts          []OpenSummary          `json:"openedShorts"`
	ClosedShortsByMonth []ClosedSummaryByMonth `json:"closedShortsByMonth"`
	ClosedShortsByYear  []ClosedSummaryByYear  `json:"closedShortsByYear"`
}

func Summary(db *gorm.DB, req *Request) (*Response, error) {
	trans, err := transaction.FindAllByUser(db, req.User)
	if err != nil {
		return nil, err
	}
	t := transaction.Transactions(trans)
	mergedTransactions := t.MergeTransactions()
	positions := mergedTransactions.CollectPositions()

	// Accounts with positions
	accounts := positions.UniqueAccounts()
	sort.Sort(sort.StringSlice(accounts))

	// Open positions, short positions
	openShorts := positions.SumProduct(transaction.PositionSummerAmount,
		transaction.ShortPositionCondition,
		transaction.OpenPositionCondition)

	openSum := []OpenSummary{}
	for _, a := range accounts {
		os := OpenSummary{
			Account: a,
			Value:   openShorts[a].Value,
		}
		openSum = append(openSum, os)
	}

	// trailing 12 montly totals
	today := time.Now()
	todayMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, &time.Location{})
	closedSummariesByMonth := []ClosedSummaryByMonth{}
	for i := 0; i < 12; i++ {
		safeDate := todayMonth.AddDate(0, -i, 0)
		closedShorts := positions.SumProduct(transaction.PositionSummerAmount,
			transaction.ShortPositionCondition,
			func(pos transaction.Position) bool {
				td, _ := time.Parse("01/02/2006", pos.Transactions[0].Date)
				return td.Month() == safeDate.Month() && td.Year() == safeDate.Year()
			},
		)
		cs := ClosedSummaryByMonth{
			Month: safeDate.Month().String(),
		}
		total := 0.0
		for _, a := range accounts {
			total += closedShorts[a].Value
			cs.Value = append(cs.Value, closedShorts[a].Value)
		}
		cs.Value = append(cs.Value, total)
		closedSummariesByMonth = append(closedSummariesByMonth, cs)
	}

	// Last 5 years
	todayYear := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, &time.Location{})
	closedSummariesByYear := []ClosedSummaryByYear{}
	for i := 0; i < 5; i++ {
		safeDate := todayYear.AddDate(-i, 0, 0)
		closedShorts := positions.SumProduct(transaction.PositionSummerAmount,
			transaction.ShortPositionCondition,
			func(pos transaction.Position) bool {
				td, _ := time.Parse("01/02/2006", pos.Transactions[0].Date)
				return td.Year() == safeDate.Year()
			},
		)
		cs := ClosedSummaryByYear{
			Year: strconv.Itoa(safeDate.Year()),
		}
		total := 0.0
		for _, a := range accounts {
			total += closedShorts[a].Value
			cs.Value = append(cs.Value, closedShorts[a].Value)
		}
		cs.Value = append(cs.Value, total)
		closedSummariesByYear = append(closedSummariesByYear, cs)
	}
	return &Response{
		Accounts:            accounts,
		OpenShorts:          openSum,
		ClosedShortsByMonth: closedSummariesByMonth,
		ClosedShortsByYear:  closedSummariesByYear,
	}, nil
}
