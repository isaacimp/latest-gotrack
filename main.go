package main

//TODO: Add way to make recipes out of fodd you have added
//TODO: imporve money tracking so you can easily see expenses and upcoming as well as breakdowns of what you can spend etc... Maybe allow user to add a budget strategy like 40-60 or somethibg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	//"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	//"fyne.io/fyne/v2/layout"
	//"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	//"github.com/sdassow/fyne-datepicker"
)

const (
	dataFile  = "foods_data.json"
	diaryFile = "diary_data.json"
)

type Food struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Calories     float64 `json:"calories"`
	Quantity     int     `json:"quantity"` // in grams
	CalPerDollar float64 `json:"cal_per_dollar"`
	CalPer100g   float64 `json:"cal_per_100g"`
}

type DiaryEntry struct {
	ID       int     `json:"id"`
	Date     string  `json:"date"`
	FoodID   int     `json:"food_id"`
	FoodName string  `json:"food_name"`
	Quantity int     `json:"quantity"` // in grams
	Calories float64 `json:"calories"`
	Cost     float64 `json:"cost"`
}

type DailyDiary struct {
	Entries []DiaryEntry `json:"entries"`
}

var (
	foods      []Food
	dailyDiary DailyDiary
)

// Save and load functions for the food database
func saveToFile() error {
	file, err := os.Create(dataFile)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(foods); err != nil {
		return fmt.Errorf("error encoding data: %v", err)
	}
	return nil
}

func loadFromFile() error {
	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			foods = make([]Food, 0)
			return nil
		}
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&foods); err != nil {
		return fmt.Errorf("error decoding data: %v", err)
	}
	return nil
}

// Save and load functions for the diary
func saveDiaryToFile() error {
	file, err := os.Create(diaryFile)
	if err != nil {
		return fmt.Errorf("error creating diary file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(dailyDiary); err != nil {
		return fmt.Errorf("error encoding diary data: %v", err)
	}
	return nil
}

func loadDiaryFromFile() error {
	file, err := os.Open(diaryFile)
	if err != nil {
		if os.IsNotExist(err) {
			dailyDiary.Entries = make([]DiaryEntry, 0)
			return nil
		}
		return fmt.Errorf("error opening diary file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&dailyDiary); err != nil {
		return fmt.Errorf("error decoding diary data: %v", err)
	}
	return nil
}

func readInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func searchFoods() {
	query := readInput("Enter food name to search: ")
	query = strings.ToLower(query)

	found := false
	for _, food := range foods {
		if strings.Contains(strings.ToLower(food.Name), query) {
			fmt.Printf("\nFound: %s\n", food.Name)
			fmt.Printf("Price: $%.2f\n", food.Price)
			fmt.Printf("Quantity: %dg\n", food.Quantity)
			fmt.Printf("Calories: %.0f\n", food.Calories)
			fmt.Printf("Calories per Dollar: %.0f\n", food.CalPerDollar)
			fmt.Printf("Calories per 100g: %.0f\n", food.CalPer100g)
			found = true
		}
	}

	if !found {
		fmt.Println("No foods found matching your search.")
	}
}

func viewStats() {
	if len(foods) == 0 {
		fmt.Println("No foods added yet.")
		return
	}

	// Create a copy of foods to sort
	sortedFoods := make([]Food, len(foods))
	copy(sortedFoods, foods)

	fmt.Println("\nFoods ordered by calories per dollar:")
	// Simple bubble sort by CalPerDollar
	for i := 0; i < len(sortedFoods)-1; i++ {
		for j := 0; j < len(sortedFoods)-i-1; j++ {
			if sortedFoods[j].CalPerDollar < sortedFoods[j+1].CalPerDollar {
				sortedFoods[j], sortedFoods[j+1] = sortedFoods[j+1], sortedFoods[j]
			}
		}
	}

	for i, food := range sortedFoods {
		fmt.Printf("%d. %s: %.0f calories/$\n",
			i+1, food.Name, food.CalPerDollar)
	}
}

func showSummary(myApp fyne.App) fyne.Window {
	window := myApp.NewWindow("Food Summary")

	// Get today's date and 30 days ago for monthly calculations
	today := time.Now()
	thirtyDaysAgo := today.AddDate(0, 0, -30)

	// Food and calorie calculations
	var (
		totalCalories   float64
		totalFoodCost   float64
		daysWithEntries int
		caloriesByDay   = make(map[string]float64)
		costByDay       = make(map[string]float64)
	)

	// Process diary entries
	for _, entry := range dailyDiary.Entries {
		entryDate, err := time.Parse("2006-01-02", entry.Date)
		if err != nil {
			continue
		}

		// Only consider entries from the last 30 days
		if entryDate.After(thirtyDaysAgo) || entryDate.Equal(thirtyDaysAgo) {
			totalCalories += entry.Calories
			totalFoodCost += entry.Cost

			// Aggregate by day
			caloriesByDay[entry.Date] += entry.Calories
			costByDay[entry.Date] += entry.Cost
		}
	}

	// Count unique days with entries
	daysWithEntries = len(caloriesByDay)

	// Calculate averages
	var avgDailyCalories, avgDailyFoodCost float64
	if daysWithEntries > 0 {
		avgDailyCalories = totalCalories / float64(daysWithEntries)
		avgDailyFoodCost = totalFoodCost / float64(daysWithEntries)
	}

	// Create summary text
	summaryText := fmt.Sprintf("30-Day Food & Health Summary\n\n")
	summaryText += fmt.Sprintf("Days tracked: %d out of 30\n", daysWithEntries)
	if daysWithEntries > 0 {
		summaryText += fmt.Sprintf("Average daily calories: %.0f\n", avgDailyCalories)
		summaryText += fmt.Sprintf("Average daily food cost: $%.2f\n", avgDailyFoodCost)
		summaryText += fmt.Sprintf("Total food spending: $%.2f\n", totalFoodCost)
	}

	// Calculate and add trends if enough data
	if daysWithEntries >= 7 {
		var prevWeekCost, currentWeekCost float64
		for date, cost := range costByDay {
			entryDate, _ := time.Parse("2006-01-02", date)
			daysAgo := today.Sub(entryDate).Hours() / 24
			if daysAgo <= 7 {
				currentWeekCost += cost
			} else if daysAgo <= 14 {
				prevWeekCost += cost
			}
		}

		if prevWeekCost > 0 {
			weekChange := ((currentWeekCost - prevWeekCost) / prevWeekCost) * 100
			summaryText += fmt.Sprintf("\nSpending trend: %.1f%% %s compared to previous week",
				math.Abs(weekChange),
				map[bool]string{true: "up", false: "down"}[weekChange > 0])
		}
	}

	// Create UI elements
	title := widget.NewLabel("Food Summary")
	title.TextStyle = fyne.TextStyle{Bold: true}

	stats := widget.NewTextGridFromString(summaryText)

	backBtn := widget.NewButton("Back", func() {
		window.Close()
	})

	// Layout content
	content := container.NewVBox(
		title,
		stats,
		backBtn,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(400, 300))
	window.Show()
	return window
}

func addFoodToDiary(myApp fyne.App) fyne.Window {
	window := myApp.NewWindow("Add Food to Day")

	// Create search entry
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search for food...")

	// Create list to show search results
	resultsList := widget.NewList(
		func() int { return 0 }, // Will be updated when search changes
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""), // Food name
				widget.NewLabel(""), // Price and quantity
				widget.NewLabel(""), // Calories
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {}, // Will be updated when search changes
	)

	// Create quantity entry (hidden initially)
	quantityEntry := widget.NewEntry()
	quantityEntry.SetPlaceHolder("Enter quantity in grams")
	quantityEntry.Hide()

	// Create add button (hidden initially)
	addButton := widget.NewButton("Add to Diary", nil)
	addButton.Hide()

	// Create status label for feedback
	statusLabel := widget.NewLabel("")

	var selectedFood Food
	var matchedFoods []Food

	// Update search results as user types
	searchEntry.OnChanged = func(searchText string) {
		matchedFoods = nil
		searchText = strings.ToLower(searchText)

		if searchText == "" {
			resultsList.Refresh()
			return
		}

		// Search for matching foods
		for _, food := range foods {
			if strings.Contains(strings.ToLower(food.Name), searchText) {
				matchedFoods = append(matchedFoods, food)
			}
		}

		// Update list data
		resultsList.Length = func() int {
			return len(matchedFoods)
		}

		resultsList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
			food := matchedFoods[id]
			box := item.(*fyne.Container)

			// Update labels in the container
			nameLabel := box.Objects[0].(*widget.Label)
			priceLabel := box.Objects[1].(*widget.Label)
			calLabel := box.Objects[2].(*widget.Label)

			nameLabel.SetText(food.Name)
			priceLabel.SetText(fmt.Sprintf("$%.2f/%dg", food.Price, food.Quantity))
			calLabel.SetText(fmt.Sprintf("%.0f cal", food.Calories))
		}

		resultsList.OnSelected = func(id widget.ListItemID) {
			selectedFood = matchedFoods[id]
			quantityEntry.Show()
			addButton.Show()
			statusLabel.SetText("")
		}

		resultsList.Refresh()
	}

	// Handle adding food to diary
	addButton.OnTapped = func() {
		quantityStr := quantityEntry.Text
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			statusLabel.SetText("Please enter a valid quantity")
			return
		}

		// Calculate proportional calories and cost
		ratio := float64(quantity) / float64(selectedFood.Quantity)
		calories := selectedFood.Calories * ratio
		cost := selectedFood.Price * ratio

		// Create diary entry
		entry := DiaryEntry{
			ID:       len(dailyDiary.Entries) + 1,
			Date:     time.Now().Format("2006-01-02"),
			FoodID:   selectedFood.ID,
			FoodName: selectedFood.Name,
			Quantity: quantity,
			Calories: calories,
			Cost:     cost,
		}

		// Add to diary and save
		dailyDiary.Entries = append(dailyDiary.Entries, entry)
		if err := saveDiaryToFile(); err != nil {
			statusLabel.SetText("Error saving diary")
			log.Printf("Warning: Failed to save diary: %v", err)
			return
		}

		// Show success message
		statusLabel.SetText(fmt.Sprintf("Added %s: %dg (%.0f cal, $%.2f)",
			selectedFood.Name, quantity, calories, cost))

		// Reset fields
		searchEntry.SetText("")
		quantityEntry.SetText("")
		quantityEntry.Hide()
		addButton.Hide()
		selectedFood = Food{}
	}

	// Cancel button
	cancelButton := widget.NewButton("Cancel", func() {
		window.Close()
	})

	// Layout everything
	content := container.NewVBox(
		widget.NewLabel("Search Foods"),
		searchEntry,
		resultsList,
		quantityEntry,
		addButton,
		statusLabel,
		cancelButton,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(400, 500))
	window.Show()
	return window
}

func showFoodWindow(myApp fyne.App) {
	window := myApp.NewWindow("Food Menu")

	addFoodBtn := widget.NewButton("Add Food", func() {
		addFoodToDatabase(myApp)
	})

	addFoodDiaryBtn := widget.NewButton("Add new food to diary", func() {
		addFoodToDiary(myApp)
	})

	viewFoodBtn := widget.NewButton("View Food Diary", func() {
		viewDiary(myApp)
	})

	searchFoodBtn := widget.NewButton("Search Foods", func() {

	})

	viewStatsBtn := widget.NewButton("View Stats", func() {

	})

	content := container.NewVBox(
		widget.NewLabel("Food Menu"),
		addFoodBtn,
		addFoodDiaryBtn,
		viewFoodBtn,
		searchFoodBtn,
		viewStatsBtn,
		widget.NewButton("Back", func() {
			window.Close()
		}),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(300, 200))
	window.Show()
}

func viewDiary(myApp fyne.App) fyne.Window {
	window := myApp.NewWindow("Food Diary")
	window.Resize(fyne.NewSize(500, 700))

	dateInput := widget.NewEntry()
	dateInput.SetPlaceHolder("YYYY/MM/DD")

	entriesText := widget.NewTextGrid()
	entriesText.Resize(fyne.NewSize(450, 500)) // Set a specific size for the text grid

	entriesScroll := container.NewScroll(entriesText)
	entriesScroll.Resize(fyne.NewSize(480, 500)) // Make sure this is big enough to show multiple lines
	entriesScroll.SetMinSize(fyne.NewSize(480, 300))

	totalCaloriesLabel := widget.NewLabel("")
	totalCostLabel := widget.NewLabel("")

	// Function to update display
	updateDisplay := func(dateStr string) {
		if dateStr == "" {
			entriesText.SetText("Please select a date")
			totalCaloriesLabel.SetText("")
			totalCostLabel.SetText("")
			return
		}

		var displayText strings.Builder
		var totalCals float64
		var totalCost float64

		// Find entries for selected date
		formattedDate := strings.ReplaceAll(dateStr, "/", "-")
		hasEntries := false

		for _, entry := range dailyDiary.Entries {
			if entry.Date == formattedDate {
				hasEntries = true
				displayText.WriteString(fmt.Sprintf("%s: %dg\n", entry.FoodName, entry.Quantity))
				displayText.WriteString(fmt.Sprintf("Calories: %.0f, Cost: $%.2f\n\n", entry.Calories, entry.Cost))
				totalCals += entry.Calories
				totalCost += entry.Cost
			}
		}

		if !hasEntries {
			entriesText.SetText("No entries for this date")
			totalCaloriesLabel.SetText("")
			totalCostLabel.SetText("")
		} else {
			entriesText.SetText(displayText.String())
			totalCaloriesLabel.SetText(fmt.Sprintf("Total Calories: %.0f", totalCals))
			totalCostLabel.SetText(fmt.Sprintf("Total Cost: $%.2f", totalCost))
		}
	}

	// Handle date input changes
	dateInput.OnChanged = func(dateStr string) {
		updateDisplay(dateStr)
	}

	// Create calendar button
	calendarBtn := widget.NewButton("📅", func() {
		// Create spinners for year, month, and day
		yearSpin := widget.NewEntry()
		monthSpin := widget.NewEntry()
		daySpin := widget.NewEntry()

		// Set current date as default
		now := time.Now()
		yearSpin.SetText(fmt.Sprintf("%d", now.Year()))
		monthSpin.SetText(fmt.Sprintf("%02d", now.Month()))
		daySpin.SetText(fmt.Sprintf("%02d", now.Day()))

		// Create date selection form
		dateForm := widget.NewForm(
			widget.NewFormItem("Year", yearSpin),
			widget.NewFormItem("Month", monthSpin),
			widget.NewFormItem("Day", daySpin),
		)

		// Show dialog
		dialog.ShowCustom("Select Date", "OK", dateForm, window)

		// Update date when OK is clicked
		dateForm.OnSubmit = func() {
			dateStr := fmt.Sprintf("%s/%s/%s", yearSpin.Text, monthSpin.Text, daySpin.Text)
			dateInput.SetText(dateStr)
		}
	})

	// Create header with larger text
	header := widget.NewLabelWithStyle("Food Diary", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create date selection container
	dateContainer := container.NewBorder(
		nil, nil, nil, calendarBtn,
		dateInput,
	)

	// Add padding around date container
	paddedDateContainer := container.NewPadded(dateContainer)

	// Create back button
	backBtn := widget.NewButton("Back", func() {
		window.Close()
	})

	// Create summary container with padding and larger text
	summaryContainer := container.NewVBox(
		widget.NewSeparator(),
		container.NewPadded(
			container.NewVBox(
				totalCaloriesLabel,
				totalCostLabel,
			),
		),
	)

	// Layout everything with padding
	content := container.NewVBox(
		container.NewPadded(header),
		paddedDateContainer,
		widget.NewSeparator(),
		entriesScroll,
		summaryContainer,
		container.NewPadded(backBtn),
	)

	window.SetContent(content)

	// Set initial date to today
	now := time.Now()
	dateInput.SetText(fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day()))

	window.Show()
	return window
}

func showFoodMenu() {
	fmt.Println("\n=== Food Tracker Menu ===")
	fmt.Println("1. Add new food to database")
	fmt.Println("2. Add food to today's diary")
	fmt.Println("3. View diary")
	fmt.Println("4. Search foods")
	fmt.Println("5. View stats")
	fmt.Println("6. Return to Main Menu")
	fmt.Print("Choose an option: ")
}

func showMainMenu() {
	fmt.Println("\n=== Main Menu ===")
	fmt.Println("1. Show quick Summary")
	fmt.Println("2. Food Tracking")
	fmt.Println("3. Symptom Tracking")
	fmt.Println("4. Compare")
	fmt.Println("5. Finances")
	fmt.Println("6. Exit")
	fmt.Print("Choose an Option by typing the number: ")
}

func compareTrackMenu() {
	fmt.Println("\n=== Compare Track Menu ===")
	fmt.Println("What would you like to compare:\n ")
	fmt.Println("1. Compare diet and symptoms")
	fmt.Print("Choose an Option by typing the number: ")
}

func compareDietSymptoms() {

}

func handleCompareMenu() {
	for {
		compareTrackMenu()
		choice := readInput("")

		switch choice {
		case "1":
			compareDietSymptoms()
		}
	}
}

func handleFoodMenu() {
	for {
		showFoodMenu()
		choice := readInput("")

		switch choice {
		case "1":

		case "2":

		case "3":

		case "4":
			searchFoods()
		case "5":
			viewStats()
		case "6":
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func loadInitialData() {
	// Load symptom data
	if err := loadSymptomData(); err != nil {
		log.Printf("Warning: Failed to load existing symptom data: %v", err)
		symptomDiary.Symptoms = make([]Symptom, 0)
	}

	if err := loadSymptomDiaryData(); err != nil {
		log.Printf("Warning: Failed to load existing symptom diary data: %v", err)
		symptomDiary.Entries = make([]SymptomEntry, 0)
	}

	// Load food data
	if err := loadFromFile(); err != nil {
		log.Printf("Warning: Failed to load existing food data: %v", err)
		foods = make([]Food, 0)
	}

	if err := loadDiaryFromFile(); err != nil {
		log.Printf("Warning: Failed to load existing diary data: %v", err)
		dailyDiary.Entries = make([]DiaryEntry, 0)
	}
}

func addFoodToDatabase(myApp fyne.App) fyne.Window {
	window := myApp.NewWindow("Add Food")

	// Rest of the function remains the same...
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Food Name")

	priceEntry := widget.NewEntry()
	priceEntry.SetPlaceHolder("Price ($)")

	quantityEntry := widget.NewEntry()
	quantityEntry.SetPlaceHolder("Quantity (grams)")

	caloriesEntry := widget.NewEntry()
	caloriesEntry.SetPlaceHolder("Total Calories")

	// Create result label for feedback
	resultLabel := widget.NewLabel("")

	// Save button with validation and processing
	saveBtn := widget.NewButton("Save Food", func() {
		// Validate and process inputs
		var food Food

		// Get name
		food.Name = nameEntry.Text
		if food.Name == "" {
			resultLabel.SetText("Please enter a food name")
			return
		}

		// Parse and validate price
		price, err := strconv.ParseFloat(priceEntry.Text, 64)
		if err != nil {
			resultLabel.SetText("Invalid price. Please enter a number")
			return
		}
		food.Price = price

		// Parse and validate quantity
		quantity, err := strconv.Atoi(quantityEntry.Text)
		if err != nil {
			resultLabel.SetText("Invalid quantity. Please enter a number")
			return
		}
		food.Quantity = quantity

		// Parse and validate calories
		calories, err := strconv.ParseFloat(caloriesEntry.Text, 64)
		if err != nil {
			resultLabel.SetText("Invalid calories. Please enter a number")
			return
		}
		food.Calories = calories

		// Calculate derived values
		food.CalPerDollar = food.Calories / food.Price
		food.CalPer100g = (food.Calories / float64(food.Quantity)) * 100

		// Set ID based on existing foods
		maxID := 0
		for _, f := range foods {
			if f.ID > maxID {
				maxID = f.ID
			}
		}
		food.ID = maxID + 1

		// Add to foods slice and save to file
		foods = append(foods, food)
		if err := saveToFile(); err != nil {
			resultLabel.SetText("Error saving food: " + err.Error())
			return
		}

		// Show success message with details
		resultText := fmt.Sprintf("Added: %s\nPrice: $%.2f\nQuantity: %dg\nCalories: %.0f\nCalories per Dollar: %.0f\nCalories per 100g: %.0f",
			food.Name,
			food.Price,
			food.Quantity,
			food.Calories,
			food.CalPerDollar,
			food.CalPer100g)
		resultLabel.SetText(resultText)

		// Clear input fields
		nameEntry.SetText("")
		priceEntry.SetText("")
		quantityEntry.SetText("")
		caloriesEntry.SetText("")
	})

	// Create back button
	backBtn := widget.NewButton("Back", func() {
		window.Close()
	})

	// Create layout
	content := container.NewVBox(
		widget.NewLabel("Add New Food"),
		nameEntry,
		priceEntry,
		quantityEntry,
		caloriesEntry,
		saveBtn,
		resultLabel,
		backBtn,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(300, 400))
	window.Show()
	return window
}

func createMainWindow(myApp fyne.App) fyne.Window {
	window := myApp.NewWindow("Health Tracker")

	// Create buttons for each menu option
	summaryBtn := widget.NewButton("View Summary", func() {
		showSummary(myApp)
	})

	foodBtn := widget.NewButton("Food Menu", func() {
		showFoodWindow(myApp)
	})

	symptomBtn := widget.NewButton("Symptom Menu", func() {
		showSymptomWindow(myApp)
	})

	compareBtn := widget.NewButton("Compare Menu", func() {
		showCompareWindow(myApp)
	})

	financeBtn := widget.NewButton("Finance Menu", func() {
		showFinanceWindow(myApp)
	})

	quitBtn := widget.NewButton("Quit", func() {
		window.Close()
	})

	// Create a vertical container for the buttons
	content := container.NewVBox(
		widget.NewLabel("Health Tracker Menu"),
		summaryBtn,
		foodBtn,
		symptomBtn,
		compareBtn,
		financeBtn,
		quitBtn,
	)

	window.SetContent(content)
	return window
}

func showSymptomWindow(myApp fyne.App) {
	window := myApp.NewWindow("Symptom Menu")

	addSymptomBtn := widget.NewButton("Add Symptom", func() {
		// Implement add symptom functionality
	})

	viewSymptomBtn := widget.NewButton("View Symptom Diary", func() {
		// Implement view symptom diary functionality
	})

	content := container.NewVBox(
		widget.NewLabel("Symptom Menu"),
		addSymptomBtn,
		viewSymptomBtn,
		widget.NewButton("Back", func() {
			window.Close()
		}),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(300, 200))
	window.Show()
}

func showCompareWindow(myApp fyne.App) {
	window := myApp.NewWindow("Compare Menu")

	compareBtn := widget.NewButton("Compare Data", func() {
		// Implement comparison functionality
	})

	content := container.NewVBox(
		widget.NewLabel("Compare Menu"),
		compareBtn,
		widget.NewButton("Back", func() {
			window.Close()
		}),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(300, 200))
	window.Show()
}

func showFinanceWindow(myApp fyne.App) {
	window := myApp.NewWindow("Finance Menu")

	addExpenseBtn := widget.NewButton("Add Expense", func() {
		// Implement add expense functionality
	})

	viewExpensesBtn := widget.NewButton("View Expenses", func() {
		// Implement view expenses functionality
	})

	content := container.NewVBox(
		widget.NewLabel("Finance Menu"),
		addExpenseBtn,
		viewExpensesBtn,
		widget.NewButton("Back", func() {
			window.Close()
		}),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(300, 200))
	window.Show()
}

func main() {
	loadInitialData()

	myApp := app.New()
	mainWindow := createMainWindow(myApp)
	mainWindow.Resize(fyne.NewSize(300, 400))
	mainWindow.ShowAndRun()
}
