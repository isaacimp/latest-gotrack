package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const financeFile = "finances_data.json"
const transactionFile = "transactions_data.json"

type TransactionType string
type AssetType string

const (
	Income   TransactionType = "income"
	Expense  TransactionType = "expense"
	Transfer TransactionType = "transfer"
)

const (
	Cash     AssetType = "cash"
	Bank     AssetType = "bank"
	Invest   AssetType = "investment"
	Property AssetType = "property"
	Vehicle  AssetType = "vehicle"
	Other    AssetType = "other"
)

type Asset struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        AssetType `json:"asset_type"`
	Value       float64   `json:"value"`
	LastUpdated string    `json:"last_updated"`
	Notes       string    `json:"notes,omitempty"`
}

type Transaction struct {
	ID              int             `json:"id"`
	Date            string          `json:"date"`
	Type            TransactionType `json:"type"`
	Category        string          `json:"category"`
	Amount          float64         `json:"amount"`
	FromAssetID     int             `json:"from_asset_id,omitempty"`
	ToAssetID       int             `json:"to_asset_id,omitempty"`
	Description     string          `json:"description"`
	IsRecurring     bool            `json:"is_recurring"`
	RecurringPeriod string          `json:"recurring_period,omitempty"`
}

type FinanceTracker struct {
	Assets       []Asset       `json:"assets"`
	Transactions []Transaction `json:"transactions"`
}

var financeTracker FinanceTracker

// Save and load functions
func saveFinanceData() error {
	file, err := os.Create(financeFile)
	if err != nil {
		return fmt.Errorf("error creating finance file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(financeTracker.Assets); err != nil {
		return fmt.Errorf("error encoding finance data: %v", err)
	}
	return nil
}

func loadFinanceData() error {
	file, err := os.Open(financeFile)
	if err != nil {
		if os.IsNotExist(err) {
			financeTracker.Assets = make([]Asset, 0)
			return nil
		}
		return fmt.Errorf("error opening finance file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&financeTracker.Assets); err != nil {
		return fmt.Errorf("error decoding finance data: %v", err)
	}
	return nil
}

func saveTransactionData() error {
	file, err := os.Create(transactionFile)
	if err != nil {
		return fmt.Errorf("error creating transaction file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(financeTracker.Transactions); err != nil {
		return fmt.Errorf("error encoding transaction data: %v", err)
	}
	return nil
}

func loadTransactionData() error {
	file, err := os.Open(transactionFile)
	if err != nil {
		if os.IsNotExist(err) {
			financeTracker.Transactions = make([]Transaction, 0)
			return nil
		}
		return fmt.Errorf("error opening transaction file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&financeTracker.Transactions); err != nil {
		return fmt.Errorf("error decoding transaction data: %v", err)
	}
	return nil
}

func showFinanceMenu() {
	fmt.Println("\n=== Finance Tracker Menu ===")
	fmt.Println("1. Add/Update Asset")
	fmt.Println("2. Remove Asset")
	fmt.Println("3. Add Transaction")
	fmt.Println("4. View Assets")
	fmt.Println("5. View Transactions")
	fmt.Println("6. View Financial Summary")
	fmt.Println("7. Return to Main Menu")
	fmt.Print("Choose an option: ")
}

func addAsset() {
	fmt.Println("\n=== Add/Update Asset ===")
	name := readInput("Enter asset name: ")

	fmt.Println("\nAsset Types:")
	fmt.Println("1. Cash")
	fmt.Println("2. Bank Account")
	fmt.Println("3. Investment")
	fmt.Println("4. Property")
	fmt.Println("5. Vehicle")
	fmt.Println("6. Other")

	typeChoice := readInput("Choose asset type: ")

	var assetType AssetType
	switch typeChoice {
	case "1":
		assetType = Cash
	case "2":
		assetType = Bank
	case "3":
		assetType = Invest
	case "4":
		assetType = Property
	case "5":
		assetType = Vehicle
	case "6":
		assetType = Other
	default:
		fmt.Println("Invalid choice. Defaulting to Other")
		assetType = Other
	}

	valueStr := readInput("Enter current value: ")
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		fmt.Println("Invalid value. Please enter a number.")
		return
	}

	notes := readInput("Enter any notes (optional): ")

	// Find max ID
	maxID := 0
	for _, a := range financeTracker.Assets {
		if a.ID > maxID {
			maxID = a.ID
		}
	}

	newAsset := Asset{
		ID:          maxID + 1,
		Name:        name,
		Type:        assetType,
		Value:       value,
		LastUpdated: time.Now().Format("2006-01-02"),
		Notes:       notes,
	}

	financeTracker.Assets = append(financeTracker.Assets, newAsset)
	if err := saveFinanceData(); err != nil {
		log.Printf("Warning: Failed to save finance data: %v", err)
	}

	fmt.Printf("\nAdded asset: %s\n", name)
	fmt.Printf("Type: %s\n", assetType)
	fmt.Printf("Value: %.2f\n", value)
}

func addTransaction() {
	if len(financeTracker.Assets) == 0 {
		fmt.Println("No assets configured. Please add an asset first.")
		return
	}

	fmt.Println("\n=== Add Transaction ===")
	fmt.Println("Transaction Types:")
	fmt.Println("1. Income")
	fmt.Println("2. Expense")
	fmt.Println("3. Transfer between assets")
	fmt.Println("4. Investment")

	typeChoice := readInput("Choose transaction type: ")

	var transType TransactionType
	switch typeChoice {
	case "1":
		transType = Income
	case "2":
		transType = Expense
	case "3":
		transType = Transfer
	default:
		fmt.Println("Invalid choice")
		return
	}

	amountStr := readInput("Enter amount: ")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		fmt.Println("Invalid amount")
		return
	}

	category := readInput("Enter category (e.g., Salary, Food, Rent): ")
	description := readInput("Enter description: ")

	var fromAssetID, toAssetID int

	if transType == Transfer {
		fmt.Println("\nAvailable Assets:")
		for _, a := range financeTracker.Assets {
			fmt.Printf("%d. %s (%.2f)\n", a.ID, a.Name, a.Value)
		}

		fromStr := readInput("Transfer from asset ID: ")
		fromAssetID, _ = strconv.Atoi(fromStr)
		toStr := readInput("Transfer to asset ID: ")
		toAssetID, _ = strconv.Atoi(toStr)
	} else {
		fmt.Println("\nSelect affected asset:")
		for _, a := range financeTracker.Assets {
			fmt.Printf("%d. %s (%.2f)\n", a.ID, a.Name, a.Value)
		}
		assetStr := readInput("Enter asset ID: ")
		fromAssetID, _ = strconv.Atoi(assetStr)
	}

	isRecurringStr := readInput("Is this a recurring transaction? (y/n): ")
	isRecurring := isRecurringStr == "y" || isRecurringStr == "Y"

	var recurringPeriod string
	if isRecurring {
		recurringPeriod = readInput("Enter recurring period (daily/weekly/monthly/yearly): ")
	}

	// Create transaction
	transaction := Transaction{
		ID:              len(financeTracker.Transactions) + 1,
		Date:            time.Now().Format("2006-01-02"),
		Type:            transType,
		Category:        category,
		Amount:          amount,
		FromAssetID:     fromAssetID,
		ToAssetID:       toAssetID,
		Description:     description,
		IsRecurring:     isRecurring,
		RecurringPeriod: recurringPeriod,
	}

	// Update asset values
	for i, asset := range financeTracker.Assets {
		if asset.ID == fromAssetID {
			if transType == Expense || transType == Transfer {
				financeTracker.Assets[i].Value -= amount
			} else if transType == Income {
				financeTracker.Assets[i].Value += amount
			}
			financeTracker.Assets[i].LastUpdated = time.Now().Format("2006-01-02")
		}
		if asset.ID == toAssetID && transType == Transfer {
			financeTracker.Assets[i].Value += amount
			financeTracker.Assets[i].LastUpdated = time.Now().Format("2006-01-02")
		}
	}

	financeTracker.Transactions = append(financeTracker.Transactions, transaction)

	if err := saveTransactionData(); err != nil {
		log.Printf("Warning: Failed to save transaction data: %v", err)
	}
	if err := saveFinanceData(); err != nil {
		log.Printf("Warning: Failed to save finance data: %v", err)
	}

	fmt.Println("\nTransaction added successfully")
}

func removeAsset() {
	if len(financeTracker.Assets) == 0 {
		fmt.Println("No assets configured to remove")
		return
	}

	fmt.Println("\n=== Current Assets ===")
	for _, a := range financeTracker.Assets {
		fmt.Printf("%d. %s (%.2f)\n", a.ID, a.Name, a.Value)
	}

	idStr := readInput("Enter asset ID to remove: ")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Invalid asset ID.")
		return
	}

	newAssets := make([]Asset, 0)
	found := false
	for _, a := range financeTracker.Assets {
		if a.ID != id {
			newAssets = append(newAssets, a)
		} else {
			found = true
		}
	}

	if !found {
		fmt.Println("Asset not found")
		return
	}

	financeTracker.Assets = newAssets
	if err := saveFinanceData(); err != nil {
		log.Printf("Warning: Failed to save finance data: %v", err)
	}

	fmt.Println("Asset removed successfully")
}

func ViewAssets() {
	if len(financeTracker.Assets) == 0 {
		fmt.Println("No assets found")
		return
	}

	fmt.Println("\n=== Current Assets ===")
	fmt.Println("----------------------------------------")

	var totalValue float64
	for _, a := range financeTracker.Assets {
		fmt.Printf("Asset: %s (ID: %d)\n", a.Name, a.ID)
		fmt.Printf("Type: %s\n", a.Type)
		fmt.Printf("Value: %.2f\n", a.Value)
		fmt.Printf("Last Updated: %s\n", a.LastUpdated)
		if a.Notes != "" {
			fmt.Printf("Notes: %s\n", a.Notes)
		}
		fmt.Println("----------------------------------------")
		totalValue += a.Value
	}

	fmt.Printf("\nTotal Assets Value: %.2f\n", totalValue)
}

func viewTransactions() {
	fmt.Println("\n=== Transaction History ===")
	dateStr := readInput("Enter date (YYYY-MM-DD) or press Enter for all transactions: ")

	var transactions []Transaction
	if dateStr == "" {
		transactions = financeTracker.Transactions
	} else {
		for _, t := range financeTracker.Transactions {
			if t.Date == dateStr {
				transactions = append(transactions, t)
			}
		}
	}

	if len(transactions) == 0 {
		fmt.Printf("No transactions found\n")
		return
	}

	fmt.Println("\nTransactions:")
	fmt.Println("----------------------------------------")
	for _, t := range transactions {
		fmt.Printf("Date: %s\n", t.Date)
		fmt.Printf("Type: %s\n", t.Type)
		fmt.Printf("Category: %s\n", t.Category)
		fmt.Printf("Amount: %.2f\n", t.Amount)
		fmt.Printf("Description: %s\n", t.Description)
		if t.IsRecurring {
			fmt.Printf("Recurring: %s\n", t.RecurringPeriod)
		}
		fmt.Println("----------------------------------------")
	}
}

func viewFinancialSummary() {
	fmt.Println("\n=== Financial Summary ===")

	// Calculate total assets
	var totalAssets float64
	assetsByType := make(map[AssetType]float64)
	for _, a := range financeTracker.Assets {
		totalAssets += a.Value
		assetsByType[a.Type] += a.Value
	}

	// Calculate income and expenses for the current month
	currentMonth := time.Now().Format("2006-01")
	var monthlyIncome, monthlyExpenses float64
	for _, t := range financeTracker.Transactions {
		if strings.HasPrefix(t.Date, currentMonth) {
			switch t.Type {
			case Income:
				monthlyIncome += t.Amount
			case Expense:
				monthlyExpenses += t.Amount
			}
		}
	}

	fmt.Println("\nAssets Breakdown:")
	fmt.Println("----------------------------------------")
	for assetType, value := range assetsByType {
		fmt.Printf("%s: %.2f (%.1f%%)\n", assetType, value, (value/totalAssets)*100)
	}
	fmt.Printf("\nTotal Assets: %.2f\n", totalAssets)

	fmt.Println("\nCurrent Month Summary:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Total Income: %.2f\n", monthlyIncome)
	fmt.Printf("Total Expenses: %.2f\n", monthlyExpenses)
	fmt.Printf("Net: %.2f\n", monthlyIncome-monthlyExpenses)

	fmt.Println("\nCategory Breakdown:")
	fmt.Println("----------------------------------------")
	categorySummary := ""
	for category, amount := range categorySummary {
		fmt.Printf("%s: %.2f\n", category, amount)
	}

	// Show recurring transactions
	fmt.Println("\nUpcoming Recurring Transactions:")
	fmt.Println("----------------------------------------")
	for _, t := range financeTracker.Transactions {
		if t.IsRecurring {
			fmt.Printf("%s (%s): %.2f - %s\n",
				t.Category,
				t.RecurringPeriod,
				t.Amount,
				t.Description)
		}
	}
}
func getAssetByID(id int) *Asset {
	for i := range financeTracker.Assets {
		if financeTracker.Assets[i].ID == id {
			return &financeTracker.Assets[i]
		}
	}
	return nil
}

func HandleFinanceMenu() {
	// Load data when starting
	if err := loadFinanceData(); err != nil {
		log.Printf("Warning: Failed to load finance data: %v", err)
	}
	if err := loadTransactionData(); err != nil {
		log.Printf("Warning: Failed to load transaction data: %v", err)
	}

	for {
		showFinanceMenu()
		choice := readInput("")

		switch choice {
		case "1":
			addAsset()
		case "2":
			removeAsset()
		case "3":
			addTransaction()
		case "4":
			ViewAssets()
		case "5":
			viewTransactions()
		case "6":
			viewFinancialSummary()
		case "7":
			// Save data before exiting
			if err := saveFinanceData(); err != nil {
				log.Printf("Warning: Failed to save finance data: %v", err)
			}
			if err := saveTransactionData(); err != nil {
				log.Printf("Warning: Failed to save transaction data: %v", err)
			}
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

// Helper function to validate currency input
func validateCurrency(input string) (float64, error) {
	// Remove any currency symbols and commas
	input = strings.ReplaceAll(input, "$", "")
	input = strings.ReplaceAll(input, ",", "")
	input = strings.TrimSpace(input)

	return strconv.ParseFloat(input, 64)
}

// Helper function to calculate recurring transaction next date
func calculateNextRecurringDate(transaction Transaction) time.Time {
	lastDate, _ := time.Parse("2006-01-02", transaction.Date)

	switch transaction.RecurringPeriod {
	case "daily":
		return lastDate.AddDate(0, 0, 1)
	case "weekly":
		return lastDate.AddDate(0, 0, 7)
	case "monthly":
		return lastDate.AddDate(0, 1, 0)
	case "yearly":
		return lastDate.AddDate(1, 0, 0)
	default:
		return lastDate
	}
}

// Initialize function to be called at program start
func InitializeFinanceTracker() {
	financeTracker = FinanceTracker{
		Assets:       make([]Asset, 0),
		Transactions: make([]Transaction, 0),
	}

	// Load existing data
	if err := loadFinanceData(); err != nil {
		log.Printf("Warning: Failed to load finance data: %v", err)
	}
	if err := loadTransactionData(); err != nil {
		log.Printf("Warning: Failed to load transaction data: %v", err)
	}
}
