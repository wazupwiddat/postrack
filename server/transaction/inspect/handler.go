package inspect

import (
	"log"
	"sort"
	"strings"

	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/quote"
	"github.com/wazupwiddat/postrack/server/transaction"
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

type InspectRequest struct {
	User *user.User
}

type InspectResponse struct {
	Accounts []string
	Symbols  []string
}

type InspectSymbolRequest struct {
	User   *user.User
	Symbol string
}

type InspectSymbolResponse struct {
	DateFrom     string
	Quote        finance.Quote
	Premium      float64
	OpenPremium  float64
	CostBasis    float64
	Quantity     float64
	Positions    []transaction.Position
	Assigned     []transaction.Transaction
	Transactions []transaction.Transaction
}

func Inspect(db *gorm.DB, req *InspectRequest) (*InspectResponse, error) {
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

	// symbols with positions
	symbols := positions.UniqueSymbols(transaction.ShortPositionCondition)
	sort.Sort(sort.StringSlice(symbols))

	return &InspectResponse{
		Accounts: accounts,
		Symbols:  symbols,
	}, nil
}

func InspectSymbol(db *gorm.DB, req *InspectSymbolRequest) (*InspectSymbolResponse, error) {
	t, err := transaction.FindAllByUser(db, req.User)
	if err != nil {
		return nil, err
	}
	acct, sym := parseSymbol(req.Symbol)
	q, err := quote.Get(sym)
	log.Println(err)

	trans := transaction.Transactions(t)
	sort.Sort(transaction.ByDate(trans))

	// filter transactions to just be the selected symbol
	trans = *trans.Filter(func(pos transaction.Transaction) bool {
		return pos.Account == acct &&
			transaction.SymbolFromOptionSymbol(pos.Symbol) == sym
	})
	dateFrom := trans[0].Date
	mergedTransactions := trans.MergeTransactions()
	positions := mergedTransactions.CollectPositions()

	// Sum all premiums (of the transactions we have)
	shorts := positions.SumProduct(transaction.PositionSummerAmount,
		transaction.ShortPositionCondition)
	open := positions.SumProduct(transaction.PositionSummerAmount,
		transaction.OpenPositionCondition,
		transaction.ShortPositionCondition)

	// b, _ := json.MarshalIndent(positions, "", "  ")
	// log.Println(string(b))
	// Collect assinged transactions
	assignedPositions := positions.Filter(func(pos transaction.Position) bool {
		return pos.Disposition == transaction.Disposition(3)
	})

	// b, _ = json.MarshalIndent(assignedPositions, "", "  ")
	// log.Println(string(b))

	assignedTransactions := trans.Filter(func(trans transaction.Transaction) bool {
		if transaction.IsOption(trans) {
			return false
		}
		for _, assignedPos := range assignedPositions {
			ps := transaction.ParseOptionSymbol(assignedPos.Symbol)
			if ps == nil {
				continue
			}
			// log.Println(trans)
			// log.Println(ps)
			match := trans.Symbol == ps.Symbol &&
				trans.Price == ps.Price &&
				((trans.Action == "Sell" && ps.OptionType == "C") || (trans.Action == "Buy" && ps.OptionType == "P"))
			if match {
				return true
			}
		}
		return false
	})

	// Sum cost basis and quantity
	assignedCollectedPositions := assignedTransactions.MergeTransactions().CollectPositions()
	costBasis := assignedCollectedPositions.SumProduct(transaction.PositionSummerAmount)
	quant := assignedPositions.SumProduct(func(pos transaction.Position, sum *transaction.SumProduct) {
		for _, tran := range pos.Transactions {
			if tran.Action == "Assigned" {
				sp := transaction.ParseOptionSymbol(tran.Symbol)
				if sp.OptionType == "C" {
					sum.Value = sum.Value - (tran.Quantity * 100)
				} else {
					sum.Value = sum.Value + (tran.Quantity * 100)
				}

			}
		}
	})

	// Filter buy/sell transactions
	buySellTrans := trans.Filter(func(trans transaction.Transaction) bool {
		if transaction.IsOption(trans) {
			return false
		}
		return trans.Action == "Sell" || trans.Action == "Buy"
	})

	// Compare that against the current price
	var quote finance.Quote
	if q != nil {
		quote = *q
	}
	return &InspectSymbolResponse{
		Quote:        quote,
		DateFrom:     dateFrom,
		Premium:      shorts[acct].Value,
		OpenPremium:  open[acct].Value,
		Assigned:     *assignedTransactions,
		CostBasis:    costBasis[acct].Value,
		Quantity:     quant[acct].Value,
		Positions:    assignedCollectedPositions,
		Transactions: *buySellTrans,
	}, nil
}

func parseSymbol(symbol string) (acct string, sym string) {
	// Brokerage - ETSY
	d := strings.Split(symbol, " - ")
	if len(d) <= 1 {
		return "", d[0]
	}
	return d[0], d[1]
}
