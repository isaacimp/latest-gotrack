package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const symptomFile = "symptoms_data.json"
const symptomDiaryFile = "symptom_diary_data.json"

type TrackingType string

const (
	SeverityScale TrackingType = "severity"
	YesNo         TrackingType = "yesno"
	Counter       TrackingType = "counter"
	Notes         TrackingType = "notes"
)

type Symptom struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	TrackingType TrackingType `json:"tracking_type"`
	ScaleMin     int          `json:"scale_min,omitempty"`
	ScaleMax     int          `json:"scale_max,omitempty"`
}

type SymptomEntry struct {
	ID            int    `json:"id"`
	Date          string `json:"date"`
	SymptomID     int    `json:"symptom_id"`
	SymptomName   string `json:"symptom_name"`
	SeverityValue int    `json:"severity_value,omitempty"`
	YesNoValue    bool   `json:"yes_no_value,omitempty"`
	CountValue    int    `json:"count_value,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

type SymptomDiary struct {
	Symptoms []Symptom      `json:"symptoms"`
	Entries  []SymptomEntry `json:"entries"`
}

var symptomDiary SymptomDiary

// Save and load functions for the symptom system
func saveSymptomData() error {
	file, err := os.Create(symptomFile)
	if err != nil {
		return fmt.Errorf("error creating symptom file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(symptomDiary.Symptoms); err != nil {
		return fmt.Errorf("error encoding symptom data: %v", err)
	}
	return nil
}

func loadSymptomData() error {
	file, err := os.Open(symptomFile)
	if err != nil {
		if os.IsNotExist(err) {
			symptomDiary.Symptoms = make([]Symptom, 0)
			return nil
		}
		return fmt.Errorf("error opening symptom file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&symptomDiary.Symptoms); err != nil {
		return fmt.Errorf("error decoding symptom data: %v", err)
	}
	return nil
}

func saveSymptomDiaryData() error {
	file, err := os.Create(symptomDiaryFile)
	if err != nil {
		return fmt.Errorf("error creating symptom diary file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(symptomDiary.Entries); err != nil {
		return fmt.Errorf("error encoding symptom diary data: %v", err)
	}
	return nil
}

func loadSymptomDiaryData() error {
	file, err := os.Open(symptomDiaryFile)
	if err != nil {
		if os.IsNotExist(err) {
			symptomDiary.Entries = make([]SymptomEntry, 0)
			return nil
		}
		return fmt.Errorf("error opening symptom diary file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&symptomDiary.Entries); err != nil {
		return fmt.Errorf("error decoding symptom diary data: %v", err)
	}
	return nil
}

func showSymptomMenu() {
	fmt.Println("\n=== Symptom Tracker Menu ===")
	fmt.Println("1. Add symptom to track")
	fmt.Println("2. Remove symptom from track")
	fmt.Println("3. Add symptom to diary")
	fmt.Println("4. View Symptom Diary")
	fmt.Println("5. Return to Main Menu")
	fmt.Print("Choose an option: ")
}

func addSymptom() {
	fmt.Println("\n=== Add New Symptom ===")
	name := readInput("Enter symptom name: ")

	fmt.Println("\nHow would you like to track this symptom?")
	fmt.Println("1. Severity Scale (e.g., 1-10)")
	fmt.Println("2. Yes/No")
	fmt.Println("3. Counter (number of occurrences)")
	fmt.Println("4. Notes only")

	trackingChoice := readInput("Choose tracking method: ")

	// Find max ID
	maxID := 0
	for _, s := range symptomDiary.Symptoms {
		if s.ID > maxID {
			maxID = s.ID
		}
	}

	newSymptom := Symptom{
		ID:   maxID + 1,
		Name: name,
	}

	switch trackingChoice {
	case "1":
		newSymptom.TrackingType = SeverityScale
		scaleMinStr := readInput("Enter minimum scale value: ")
		scaleMin, err := strconv.Atoi(scaleMinStr)
		if err != nil {
			fmt.Println("Invalid minimum value. Using default of 1")
			scaleMin = 1
		}

		scaleMaxStr := readInput("Enter maximum scale value: ")
		scaleMax, err := strconv.Atoi(scaleMaxStr)
		if err != nil {
			fmt.Println("Invalid maximum value. Using default of 10")
			scaleMax = 10
		}

		newSymptom.ScaleMin = scaleMin
		newSymptom.ScaleMax = scaleMax
	case "2":
		newSymptom.TrackingType = YesNo
	case "3":
		newSymptom.TrackingType = Counter
	case "4":
		newSymptom.TrackingType = Notes
	default:
		fmt.Println("Invalid choice. Defaulting to severity scale (1-10)")
		newSymptom.TrackingType = SeverityScale
		newSymptom.ScaleMin = 1
		newSymptom.ScaleMax = 10
	}

	symptomDiary.Symptoms = append(symptomDiary.Symptoms, newSymptom)
	if err := saveSymptomData(); err != nil {
		log.Printf("Warning: Failed to save symptom data: %v", err)
	}

	fmt.Printf("\nAdded symptom: %s\n", name)
	fmt.Printf("Tracking type: %s\n", newSymptom.TrackingType)
	if newSymptom.TrackingType == SeverityScale {
		fmt.Printf("Scale range: %d-%d\n", newSymptom.ScaleMin, newSymptom.ScaleMax)
	}
}

func addSymptomEntry() {
	if len(symptomDiary.Symptoms) == 0 {
		fmt.Println("No symptoms configured to track. Please add a symptom first.")
		return
	}

	fmt.Println("\n=== Available Symptoms ===")
	for _, s := range symptomDiary.Symptoms {
		fmt.Printf("%d. %s (%s)\n", s.ID, s.Name, s.TrackingType)
	}

	idStr := readInput("Enter symptom ID: ")
	symptomID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Invalid symptom ID.")
		return
	}

	var symptom Symptom
	found := false
	for _, s := range symptomDiary.Symptoms {
		if s.ID == symptomID {
			symptom = s
			found = true
			break
		}
	}

	if !found {
		fmt.Println("Symptom not found")
		return
	}

	entry := SymptomEntry{
		ID:          len(symptomDiary.Entries) + 1,
		Date:        time.Now().Format("2006-01-02"),
		SymptomID:   symptomID,
		SymptomName: symptom.Name,
	}

	switch symptom.TrackingType {
	case SeverityScale:
		severityStr := readInput(fmt.Sprintf("Enter severity (%d-%d): ", symptom.ScaleMin, symptom.ScaleMax))
		severity, err := strconv.Atoi(severityStr)
		if err != nil || severity < symptom.ScaleMin || severity > symptom.ScaleMax {
			fmt.Println("Invalid severity value")
			return
		}
		entry.SeverityValue = severity

	case YesNo:
		response := readInput("Did you experience this symptom today? (y/n): ")
		entry.YesNoValue = response == "y" || response == "Y"

	case Counter:
		countStr := readInput("How many times did this occur? ")
		count, err := strconv.Atoi(countStr)
		if err != nil || count < 0 {
			fmt.Println("Invalid count value")
			return
		}
		entry.CountValue = count

	case Notes:
		entry.Notes = readInput("Enter notes about this symptom: ")
	}

	symptomDiary.Entries = append(symptomDiary.Entries, entry)
	if err := saveSymptomDiaryData(); err != nil {
		log.Printf("Warning: Failed to save symptom diary data: %v", err)
	}

	fmt.Println("\nSymptom entry added successfully")
}

func removeSymptom() {
	if len(symptomDiary.Symptoms) == 0 {
		fmt.Println("No symptoms configured to remove")
		return
	}

	fmt.Println("\n=== Current Symptoms ===")
	for _, s := range symptomDiary.Symptoms {
		fmt.Printf("%d. %s\n", s.ID, s.Name)
	}

	idStr := readInput("Enter symptom ID to remove: ")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Invalid symptom ID.")
		return
	}

	newSymptoms := make([]Symptom, 0)
	found := false
	for _, s := range symptomDiary.Symptoms {
		if s.ID != id {
			newSymptoms = append(newSymptoms, s)
		} else {
			found = true
		}
	}

	if !found {
		fmt.Println("Symptom not found")
		return
	}

	symptomDiary.Symptoms = newSymptoms
	if err := saveSymptomData(); err != nil {
		log.Printf("Warning: Failed to save symptom data: %v", err)
	}

	fmt.Println("Symptom removed successfully")
}

func viewSymptomDiary() {
	fmt.Println("\n=== Symptom Diary Viewer ===")
	dateStr := readInput("Enter date (YYYY-MM-DD) or press Enter for today: ")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	var dayEntries []SymptomEntry
	for _, entry := range symptomDiary.Entries {
		if entry.Date == dateStr {
			dayEntries = append(dayEntries, entry)
		}
	}

	if len(dayEntries) == 0 {
		fmt.Printf("No entries found for %s\n", dateStr)
		return
	}

	fmt.Printf("\nSymptom Diary for %s:\n", dateStr)
	fmt.Println("----------------------------------------")
	for _, entry := range dayEntries {
		fmt.Printf("Symptom: %s\n", entry.SymptomName)

		switch getSymptomType(entry.SymptomID) {
		case SeverityScale:
			fmt.Printf("  Severity: %d\n", entry.SeverityValue)
		case YesNo:
			fmt.Printf("  Experienced: %v\n", entry.YesNoValue)
		case Counter:
			fmt.Printf("  Count: %d\n", entry.CountValue)
		case Notes:
			fmt.Printf("  Notes: %s\n", entry.Notes)
		}
	}
	fmt.Println("----------------------------------------")
}

func getSymptomType(id int) TrackingType {
	for _, s := range symptomDiary.Symptoms {
		if s.ID == id {
			return s.TrackingType
		}
	}
	return ""
}

func HandleSymptomMenu() {
	for {
		showSymptomMenu()
		choice := readInput("")

		switch choice {
		case "1":
			addSymptom()
		case "2":
			removeSymptom()
		case "3":
			addSymptomEntry()
		case "4":
			viewSymptomDiary()
		case "5":
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}
