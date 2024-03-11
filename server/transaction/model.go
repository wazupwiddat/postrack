package transaction

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	ID          uint   `gorm:"primary_key"`
	UserID      uint   `gorm:"index"`
	Account     string `gorm:"size:100"`
	Date        string `gorm:"size:50"`
	Action      string `gorm:"size:50"`
	Symbol      string `gorm:"size:50"`
	Description string `gorm:"size:250"`
	Quantity    float64
	Price       float64
	FeesComm    float64
	Amount      float64

	UniqueID string `gorm:"-:all"`
}

type TransactionFilterCond func(pos Transaction) bool

type Split struct {
	Symbol string
	Date   time.Time
	Ratio  int
}
type Splits []Split

type MergedTransactions map[string][]Transaction

type Transactions []Transaction

type ByUniqueID Transactions

func (a ByUniqueID) Len() int           { return len(a) }
func (a ByUniqueID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByUniqueID) Less(i, j int) bool { return a[i].UniqueID < a[j].UniqueID }

type ByDate []Transaction

func (a ByDate) Len() int {
	return len(a)
}

func (a ByDate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByDate) Less(i, j int) bool {
	ai, err := time.Parse("01/02/2006", a[i].Date)
	if err != nil {
		log.Println(err)
	}
	aj, err := time.Parse("01/02/2006", a[j].Date)
	if err != nil {
		log.Println(err)
	}
	// if ai.Equal(aj) {
	// 	return (a[i].Direction == "Sold" || a[i].Direction == "Bought") && (a[j].Direction == "Closed")
	// }
	return ai.Before(aj)
}

type Direction int

const (
	dirLong = iota
	dirShort
)

type Disposition int

const (
	dispOpened = iota
	dispClosed
	dispExpired
	dispAssigned
)

type Position struct {
	Account     string
	Symbol      string
	Amount      float64
	Direction   Direction
	Quantity    float64
	Disposition Disposition
	Transactions
}

type PositionFilterCond func(pos Position) bool
type PositionSummerFunc func(pos Position, sum *SumProduct)
type Positions []Position
type PostionsByDate []Position

func (a PostionsByDate) Len() int {
	return len(a)
}

func (a PostionsByDate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a PostionsByDate) Less(i, j int) bool {
	ai, err := time.Parse("01/02/2006", a[i].Transactions[0].Date)
	if err != nil {
		log.Println(err)
	}
	aj, err := time.Parse("01/02/2006", a[j].Transactions[0].Date)
	if err != nil {
		log.Println(err)
	}
	// if ai.Equal(aj) {
	// 	return (a[i].Direction == "Sold" || a[i].Direction == "Bought") && (a[j].Direction == "Closed")
	// }
	return ai.Before(aj)
}

type SumProduct struct {
	Value float64
}

type OptionSymbol struct {
	Symbol     string
	Date       string
	Price      float64
	OptionType string
}

func (t *Transactions) MergeTransactions() *MergedTransactions {
	splits := []Split{}
	splitDate, _ := time.Parse("2006-01-02", "2022-08-25")
	splits = append(splits, Split{
		Symbol: "TSLA",
		Date:   splitDate,
		Ratio:  3,
	})
	splitDate, _ = time.Parse("2006-01-02", "2020-08-31")
	splits = append(splits, Split{
		Symbol: "TSLA",
		Date:   splitDate,
		Ratio:  5,
	})

	return t.Filter(NonEmptySymbolCondition).
		Filter(ValidActionsCondition).
		ApplySplits(splits).
		WithUniqueIdentifier().
		WithMergedByUniqueID()
}

func (t *Transactions) SortByDate() *Transactions {
	sort.Sort(ByDate(*t))
	return t
}

func (t *Transactions) ApplySplits(splits Splits) *Transactions {
	for _, s := range splits {
		for idx, tran := range *t {
			sym := SymbolFromOptionSymbol(tran.Symbol)
			if sym != s.Symbol {
				continue
			}
			td, _ := time.Parse("01/02/2006", tran.Date)
			if td.Before(s.Date) {
				if IsOption(tran) {
					// TSLA 09/09/2022 800.00 P -> TSLA 09/09/2022 266.67 P
					tran.Symbol = adjustSymbolForSplit(tran.Symbol, s)
				}
				tran.Quantity = float64(s.Ratio) * tran.Quantity
				(*t)[idx] = tran
			}
		}
	}
	return t
}

func (t *Transactions) Filter(cond TransactionFilterCond) *Transactions {
	trans := Transactions{}
	for _, tran := range *t {
		if cond(tran) {
			trans = append(trans, tran)
		}
	}
	return &trans
}

func (t *Transactions) FindFirst(cond TransactionFilterCond) *Transaction {
	for _, tran := range *t {
		if cond(tran) {
			return &tran
		}
	}
	return nil
}

func (t *Transactions) FindPreviousFrom(startTrans *Transaction, cond TransactionFilterCond) *Transaction {
	if startTrans == nil {
		return &(*t)[0]
	}
	startFromIdx := 0
	for idx, tran := range *t {
		if tran.ID == startTrans.ID {
			startFromIdx = idx
			break
		}
	}
	log.Println("Start from index is ", startFromIdx)
	for i := startFromIdx - 1; i >= 0; i-- {
		if cond((*t)[i]) {
			log.Println("Condition met and returning", i)
			return &(*t)[i]
		}
	}
	log.Println("No Condition met returning the first element", 0)
	return &(*t)[0]
}

func NonEmptySymbolCondition(tran Transaction) bool {
	return tran.Symbol != "" && strings.TrimSpace(tran.Symbol) != ""
}

func ValidActionsCondition(tran Transaction) bool {
	removeAction := []string{"Options Frwd Split", "Journal", "Stock Split"}
	for _, ra := range removeAction {
		if tran.Action == ra {
			return false
		}
	}
	return !strings.Contains(tran.Description, "FORWARD SPLIT WITH STOCK SPLIT SHARES")
}

func (t *Transactions) WithUniqueIdentifier() *Transactions {
	trans := Transactions{}
	for _, tran := range *t {
		h := md5.New()
		io.WriteString(h, tran.Account)
		io.WriteString(h, tran.Symbol)
		tran.UniqueID = fmt.Sprintf("%x", h.Sum(nil))
		trans = append(trans, tran)
	}
	return &trans
}

func (t *Transactions) WithMergedByUniqueID() *MergedTransactions {
	sort.Sort(ByUniqueID(*t))
	mergedTransactions := MergedTransactions{}
	for _, trans := range *t {
		mergedTransaction, ok := mergedTransactions[trans.UniqueID]
		if !ok {
			merged := []Transaction{}
			merged = append(merged, trans)
			mergedTransactions[trans.UniqueID] = merged
			continue
		}
		mergedTransaction = append(mergedTransaction, trans)
		mergedTransactions[trans.UniqueID] = mergedTransaction
	}
	return &mergedTransactions
}

func (m *MergedTransactions) CollectPositions() Positions {
	result := Positions{}
	for _, merged := range *m {
		sort.Sort(ByDate(merged))
		var dir Direction = dirLong
		if IsOption(merged[0]) {
			dir = getDirection(merged[0])
		}
		pos := Position{
			Account:      merged[0].Account,
			Symbol:       merged[0].Symbol,
			Direction:    dir,
			Transactions: merged,
			Disposition:  dispClosed,
		}
		amt := 0.0
		quant := 0.0
		for _, t := range merged {
			amt += t.Amount
			if getDirection(t) == dirLong {
				quant += t.Quantity
			} else {
				quant -= t.Quantity
			}
			switch t.Action {
			case "Assigned":
				pos.Disposition = dispAssigned
			case "Expired":
				pos.Disposition = dispExpired
			}

		}
		pos.Amount = amt
		pos.Quantity = quant
		if pos.Direction == dirLong && quant > 0 {
			pos.Disposition = dispOpened
		}
		if pos.Direction == dirShort && quant < 0 {
			pos.Disposition = dispOpened
		}
		result = append(result, pos)
	}
	return result
}

func getDirection(t Transaction) Direction {
	d := strings.Split(t.Action, " to ")
	if len(d) <= 1 {
		if d[0] == "Buy" {
			return dirLong
		}
		if d[0] == "Assigned" || d[0] == "Expired" {
			return dirLong
		}
		return dirShort
	}
	if strings.Contains(d[0], "Sell") && strings.Contains(d[1], "Open") {
		return dirShort
	}
	if strings.Contains(d[0], "Buy") && strings.Contains(d[1], "Open") {
		return dirLong
	}
	if strings.Contains(d[0], "Sell") && strings.Contains(d[1], "Close") {
		return dirShort
	}
	if strings.Contains(d[0], "Buy") && strings.Contains(d[1], "Close") {
		return dirLong
	}

	return dirLong
}

func (p Positions) Filter(cond PositionFilterCond) Positions {
	result := Positions{}
	for _, pos := range p {
		if cond(pos) {
			result = append(result, pos)
		}
	}
	return result
}

func (p Positions) SumProduct(summer PositionSummerFunc, cond ...PositionFilterCond) map[string]SumProduct {
	result := map[string]SumProduct{}
	for _, pos := range p {
		if checkCondition(cond, pos) {
			sum, ok := result[pos.Account]
			if !ok {
				sum = SumProduct{}
			}
			summer(pos, &sum)
			result[pos.Account] = sum
		}
	}
	return result
}

func PositionSummerAmount(pos Position, sum *SumProduct) {
	sum.Value = sum.Value + pos.Amount
}

func PositionSummerQuantity(pos Position, sum *SumProduct) {
	sum.Value = sum.Value + pos.Quantity
}

func checkCondition(cond []PositionFilterCond, pos Position) bool {
	for _, c := range cond {
		if !c(pos) {
			return false
		}
	}
	return true
}

func OpenPositionCondition(pos Position) bool {
	return pos.Disposition == dispOpened
}

func ClosedPositionCondition(pos Position) bool {
	return pos.Disposition != dispOpened
}

func ShortPositionCondition(pos Position) bool {
	return pos.Direction == dirShort
}

func LongPositionCondition(pos Position) bool {
	return pos.Direction == dirLong
}

func (p Positions) UniqueAccounts() []string {
	result := []string{}
	accts := map[string]string{}
	for _, pos := range p {
		_, ok := accts[pos.Account]
		if !ok {
			accts[pos.Account] = "1"
			result = append(result, pos.Account)
		}
	}
	return result
}

func (p Positions) UniqueSymbols(cond ...PositionFilterCond) []string {
	result := []string{}
	symbols := map[string]string{}
	for _, pos := range p {
		if checkCondition(cond, pos) {
			sym := SymbolFromOptionSymbol(pos.Symbol)
			key := fmt.Sprintf("%s - %s", pos.Account, sym)
			_, ok := symbols[key]
			if !ok {
				symbols[key] = "1"
				result = append(result, key)
			}
		}
	}
	return result
}

func IsOption(tran Transaction) bool {
	d := strings.Split(tran.Symbol, " ")
	if len(d) <= 3 {
		return false
	}
	return true
}

func adjustSymbolForSplit(symbol string, s Split) string {
	d := strings.Split(symbol, " ")
	if len(d) <= 2 {
		return symbol
	}
	strike, err := strconv.ParseFloat(d[2], 64)
	if err != nil {
		return symbol
	}
	newStrike := strike / float64(s.Ratio)
	d[2] = fmt.Sprintf("%.2f", newStrike)
	return strings.Join(d, " ")
}

func SymbolFromOptionSymbol(sym string) string {
	d := strings.Split(sym, " ")
	if len(d) <= 1 {
		return sym
	}
	return d[0]
}

func ParseOptionSymbol(sym string) *OptionSymbol {
	d := strings.Split(sym, " ")
	if len(d) != 4 {
		return nil
	}
	strike, err := strconv.ParseFloat(d[2], 64)
	if err != nil {
		return nil
	}
	os := OptionSymbol{
		Symbol:     d[0],
		Date:       d[1],
		Price:      strike,
		OptionType: d[3],
	}
	return &os
}
