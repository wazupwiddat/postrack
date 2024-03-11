package transaction_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/piquette/finance-go/quote"
	"github.com/wazupwiddat/postrack/server/transaction"
)

func TestWithNonEmptySymbol(t *testing.T) {
	trans := &transaction.Transactions{
		transaction.Transaction{Symbol: "AAPL", Price: 100},
		transaction.Transaction{Symbol: "", Price: 200},
		transaction.Transaction{Symbol: "GOOG", Price: 300},
		transaction.Transaction{Symbol: " ", Price: 200},
	}
	merged := *trans.Filter(transaction.NonEmptySymbolCondition)
	if len(merged) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(merged))
	}
	if merged[0].Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", merged[0].Symbol)
	}
	if merged[1].Symbol != "GOOG" {
		t.Errorf("Expected symbol GOOG, got %s", merged[1].Symbol)
	}
}

func TestWithUniqueIdentifier(t *testing.T) {
	transactions := transaction.Transactions{
		transaction.Transaction{Account: "Acct1", Symbol: "AAPL"},
		transaction.Transaction{Account: "Acct2", Symbol: "GOOG"},
		transaction.Transaction{Account: "Acct1", Symbol: "AAPL"},
	}
	uniqueTransactions := transactions.WithUniqueIdentifier()
	if len(*uniqueTransactions) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(*uniqueTransactions))
	}
	if (*uniqueTransactions)[0].UniqueID != (*uniqueTransactions)[2].UniqueID {
		t.Errorf("Expected unique IDs, got same ID")
	}
}

func TestWithMergedByUniqueID(t *testing.T) {
	transactions := transaction.Transactions{
		transaction.Transaction{UniqueID: "1", Amount: 10},
		transaction.Transaction{UniqueID: "2", Amount: 20},
		transaction.Transaction{UniqueID: "1", Amount: 30},
	}
	summedTransactions := transactions.WithMergedByUniqueID()
	if len(*summedTransactions) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(*summedTransactions))
	}
	if (*summedTransactions)["1"][0].Amount+(*summedTransactions)["1"][1].Amount != 40 {
		t.Errorf("Expected summed amount of 40, got %f", (*summedTransactions)["1"][0].Amount+(*summedTransactions)["1"][1].Amount)
	}
}

func TestAssignedTransactions(t *testing.T) {
	jsonFile, err := os.Open("../../test_data/underlyingTransactions.json")
	if err != nil {
		t.Errorf("Unable to open file %s", "./test_data/underlyingTransactions.json")
	}
	defer jsonFile.Close()

	b, _ := ioutil.ReadAll(jsonFile)

	var trans transaction.Transactions
	json.Unmarshal(b, &trans)
	acct, sym := "IRA", "SHOP"

	sort.Sort(transaction.ByDate(trans))
	trans = *trans.Filter(func(pos transaction.Transaction) bool {
		return pos.Account == acct &&
			transaction.SymbolFromOptionSymbol(pos.Symbol) == sym
	})
	mergedTransactions := trans.MergeTransactions()
	positions := mergedTransactions.CollectPositions()

	// Sum all premiums (of the transactions we have)
	shorts := positions.SumProduct(transaction.PositionSummerAmount,
		transaction.ShortPositionCondition)
	// Collect assinged transactions
	assignedPositions := positions.Filter(func(pos transaction.Position) bool {
		return pos.Disposition == transaction.Disposition(3)
	})
	assignedTransactions := trans.Filter(func(trans transaction.Transaction) bool {
		if transaction.IsOption(trans) {
			return false
		}
		for _, assignedPos := range assignedPositions {
			ps := transaction.ParseOptionSymbol(assignedPos.Symbol)
			if ps == nil {
				continue
			}
			match := trans.Symbol == ps.Symbol &&
				trans.Date == ps.Date &&
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

	if len(shorts) == 0 {
		t.Errorf("Shorts should not be 0")
	}
	if len(assignedPositions) != 3 {
		t.Errorf("assigned positions should be 3")
	}
	if len(*assignedTransactions) != 3 {
		t.Errorf("assigned transactions should be 3")
	}

	if len(assignedCollectedPositions) != 1 {
		t.Errorf("assigned transactions should be 1")
	}
	if costBasis[acct].Value != -60001.509999999995 {
		t.Errorf("cost basis should be 60001.59")
	}
	if quant[acct].Value != 2000 {
		t.Errorf("quantity should be 2000")
	}
}

func TestCollectPositionFromFileAll(t *testing.T) {
	jsonFile, err := os.Open("../../test_data/transactions.json")
	if err != nil {
		t.Errorf("Unable to open file %s", "./test_data/transactions.json")
	}
	defer jsonFile.Close()

	b, _ := ioutil.ReadAll(jsonFile)

	var transactions transaction.Transactions
	json.Unmarshal(b, &transactions)

	mergedTransactions := transactions.MergeTransactions()
	positions := mergedTransactions.CollectPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 positions transactions, but got %d", len(positions))
	}
	if positions[0].Symbol != "SQ" && positions[0].Account != "IRA" {
		t.Errorf("Error with the collectPosition")
	}
	if positions[0].Quantity != -121 {
		t.Errorf("Position quantity is wrong")
	}
	if positions[0].Disposition != transaction.Disposition(1) {
		t.Errorf("Position should be closed")
	}

	quantity := positions.SumProduct(transaction.PositionSummerQuantity,
		func(pos transaction.Position) bool {
			return pos.Symbol == "SQ" // this should just be the underlying position quantity
		},
	)
	if quantity["IRA"].Value != -121 {
		t.Errorf("Error summing quantity")
	}
}

func TestCollectPositionFromFileShortClosed(t *testing.T) {
	jsonFile, err := os.Open("../../test_data/shortClosedTransactions.json")
	if err != nil {
		t.Errorf("Unable to open file %s", "./test_data/shortClosedTransactions.json")
	}
	defer jsonFile.Close()

	b, _ := ioutil.ReadAll(jsonFile)

	var transactions transaction.Transactions
	json.Unmarshal(b, &transactions)

	mergedTransactions := transactions.MergeTransactions()
	positions := mergedTransactions.CollectPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 positions transactions, but got %d", len(positions))
	}
	if positions[0].Quantity != 0 {
		t.Errorf("Position quantity is wrong")
	}
	if positions[0].Disposition != transaction.Disposition(1) {
		t.Errorf("Position should be closed")
	}
}

func TestCollectPositionFromFileShortExpired(t *testing.T) {
	jsonFile, err := os.Open("../../test_data/shortExpiredTransactions.json")
	if err != nil {
		t.Errorf("Unable to open file %s", "./test_data/shortExpiredTransactions.json")
	}
	defer jsonFile.Close()

	b, _ := ioutil.ReadAll(jsonFile)

	var transactions transaction.Transactions
	json.Unmarshal(b, &transactions)

	mergedTransactions := transactions.MergeTransactions()
	positions := mergedTransactions.CollectPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 positions transactions, but got %d", len(positions))
	}
	if positions[0].Quantity != 0 {
		t.Errorf("Position quantity is wrong")
	}
	if positions[0].Disposition != transaction.Disposition(2) {
		t.Errorf("Position should be closed")
	}
}

func TestCollectPositionFromFileShortAssignedNotWorking(t *testing.T) {
	jsonFile, err := os.Open("../../test_data/shortAssignedTransactionNotWorking.json")
	if err != nil {
		t.Errorf("Unable to open file %s", "./test_data/shortAssignedTransactionNotWorking.json")
	}
	defer jsonFile.Close()

	b, _ := ioutil.ReadAll(jsonFile)

	var transactions transaction.Transactions
	json.Unmarshal(b, &transactions)

	mergedTransactions := transactions.MergeTransactions()
	positions := mergedTransactions.CollectPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 positions transactions, but got %d", len(positions))
	}
	if positions[0].Quantity != -2 {
		t.Errorf("Position quantity is wrong")
	}
	if positions[0].Disposition != transaction.Disposition(0) {
		t.Errorf("Position should be closed")
	}

}

func TestCollectPositionFromFileShortAssigned(t *testing.T) {
	jsonFile, err := os.Open("../../test_data/shortAssignedTransactions.json")
	if err != nil {
		t.Errorf("Unable to open file %s", "./test_data/shortAssignedTransactions.json")
	}
	defer jsonFile.Close()

	b, _ := ioutil.ReadAll(jsonFile)

	var transactions transaction.Transactions
	json.Unmarshal(b, &transactions)

	mergedTransactions := transactions.MergeTransactions()
	positions := mergedTransactions.CollectPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 positions transactions, but got %d", len(positions))
	}
	if positions[0].Quantity != 0 {
		t.Errorf("Position quantity is wrong")
	}
	if positions[0].Disposition != transaction.Disposition(3) {
		t.Errorf("Position should be closed")
	}

}

func TestCollectPositionFromFileOpened(t *testing.T) {
	jsonFile, err := os.Open("../../test_data/openedTransactions.json")
	if err != nil {
		t.Errorf("Unable to open file %s", "./test_data/openedTransactions.json")
	}
	defer jsonFile.Close()

	b, _ := ioutil.ReadAll(jsonFile)

	var transactions transaction.Transactions
	json.Unmarshal(b, &transactions)

	mergedTransactions := transactions.MergeTransactions()
	positions := mergedTransactions.CollectPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 positions transactions, but got %d", len(positions))
	}
	if positions[0].Symbol != "SQ" && positions[0].Account != "IRA" {
		t.Errorf("Error with the collectPosition")
	}
	if positions[0].Quantity != 1279 {
		t.Errorf("Position quantity is wrong, %f", positions[0].Quantity)
	}
	if positions[0].Disposition != transaction.Disposition(0) {
		t.Errorf("Position should be Opened")
	}

	quantity := positions.SumProduct(transaction.PositionSummerQuantity,
		func(pos transaction.Position) bool {
			return pos.Symbol == "SQ" // this should just be the underlying position quantity
		},
	)
	if quantity["IRA"].Value != 1279 {
		t.Errorf("Error summing quantity, %f", quantity["IRA"].Value)
	}
}

func TestCollectPositions(t *testing.T) {
	transactions := transaction.Transactions{
		transaction.Transaction{Account: "A", Symbol: "S1", Action: "Buy", Quantity: 16, Amount: -10.0, Date: "12/01/2022"},
		transaction.Transaction{Account: "A", Symbol: "S1", Action: "Sell", Quantity: 16, Amount: 20.0, Date: "12/02/2022"},
		transaction.Transaction{Account: "B", Symbol: "S1", Quantity: 16, Amount: 5.0, Date: "12/03/2022"},
		transaction.Transaction{Account: "B", Symbol: "S2", Quantity: 16, Amount: 15.0, Date: "12/04/2022"},
	}
	mergedTransactions := transactions.MergeTransactions()
	positions := mergedTransactions.CollectPositions()

	if len(positions) != 3 {
		t.Errorf("Expected 3 positions transactions, but got %d", len(positions))
	}

	for _, s := range positions {
		if s.Symbol == "S1" && s.Account == "A" {
			if s.Amount != 10.0 {
				t.Errorf("Expected a position amount of 10.0 for Account A and Symbol S1, but got %f", s.Amount)
			}
			if s.Quantity != 0 {
				t.Errorf("Expected a position quantity of 0 for Account A and Symbol S1, but got %f", s.Quantity)
			}
		} else if s.Symbol == "S1" && s.Account == "B" {
			if s.Amount != 5.0 {
				t.Errorf("Expected a position amount of 5.0 for Account B and Symbol S1, but got %f", s.Amount)
			}
		} else if s.Symbol == "S2" && s.Account == "B" {
			if s.Amount != 15.0 {
				t.Errorf("Expected a position amount of 15.0 for Account B and Symbol S2, but got %f", s.Amount)
			}
		}
	}
}

func TestApplySplits(t *testing.T) {
	// given
	transactions := transaction.Transactions{
		{Symbol: "TSLA", Date: "08/09/2022", Quantity: 100},
		{Symbol: "AAPL", Date: "09/09/2022", Quantity: 200},
		{Symbol: "TSLA 09/10/2022 800.00 P", Date: "08/09/2022", Quantity: 10},
	}
	splits := transaction.Splits{
		{Symbol: "TSLA", Date: time.Date(2022, 9, 9, 0, 0, 0, 0, time.Local), Ratio: 3},
	}
	expected := transaction.Transactions{
		{Symbol: "TSLA", Date: "08/09/2022", Quantity: 300},
		{Symbol: "AAPL", Date: "09/09/2022", Quantity: 200},
		{Symbol: "TSLA 09/10/2022 266.67 P", Date: "08/09/2022", Quantity: 30},
	}

	// when
	result := transactions.ApplySplits(splits)

	// then
	if len(*result) != len(expected) {
		t.Errorf("Expected length %d but got %d", len(expected), len(*result))
	}
	for i, tran := range *result {
		if tran.Symbol != expected[i].Symbol || tran.Date != expected[i].Date || tran.Quantity != expected[i].Quantity {
			t.Errorf("Expected %v but got %v", expected[i], tran)
		}
	}
}

func TestUniqueAccounts(t *testing.T) {
	positions := transaction.Positions{
		{Account: "Acct1", Symbol: "TSLA", Quantity: 100},
		{Account: "Acct2", Symbol: "AAPL", Quantity: 200},
		{Account: "Acct1", Symbol: "GOOG", Quantity: 300},
	}
	expected := []string{"Acct1", "Acct2"}
	result := positions.UniqueAccounts()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

func TestSumProduct(t *testing.T) {
	positions := transaction.Positions{
		{Account: "A", Symbol: "APPL", Amount: 100.0},
		{Account: "A", Symbol: "GOOG", Amount: 200.0},
		{Account: "B", Symbol: "APPL", Amount: 300.0},
		{Account: "B", Symbol: "GOOG", Amount: 400.0},
	}

	cond := func(pos transaction.Position) bool {
		return pos.Symbol == "APPL"
	}

	result := positions.SumProduct(transaction.PositionSummerAmount, cond)
	expectedResult := map[string]transaction.SumProduct{
		"A": {Value: 100.0},
		"B": {Value: 300.0},
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected result to be %v but got %v", expectedResult, result)
	}

	cond1 := func(pos transaction.Position) bool {
		return pos.Symbol == "APPL"
	}
	cond2 := func(pos transaction.Position) bool {
		return pos.Account == "B"
	}
	result1 := positions.SumProduct(transaction.PositionSummerAmount, cond1, cond2)
	expectedResult1 := map[string]transaction.SumProduct{
		"B": {Value: 300.0},
	}

	if !reflect.DeepEqual(result1, expectedResult1) {
		t.Errorf("Expected result to be %v but got %v", expectedResult1, result1)
	}
}

func TestQuote(t *testing.T) {
	q, err := quote.Get("AAPL")
	if err != nil {
		// Uh-oh.
		panic(err)
	}

	// Success!
	fmt.Println(q)
}
