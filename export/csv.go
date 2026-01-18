package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"lifting/program"
)

// ToCSV exports the program to a CSV file
func ToCSV(prog *program.Program, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Week", "Day", "Exercise", "Sets", "Reps", "Weight", "Percentage"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write each day's sets
	for _, day := range prog.Days {
		for _, set := range day.Sets {
			row := formatRow(day.Week, day.DayNum, set)
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("failed to write row: %w", err)
			}
		}
	}

	return nil
}

// formatRow formats a set as a CSV row
func formatRow(week, day int, set program.Set) []string {
	// Format weight and percentage (blank for accessories)
	weightStr := ""
	pctStr := ""
	if set.Weight > 0 {
		weightStr = fmt.Sprintf("%.0f", set.Weight)
	}
	if set.Percentage > 0 {
		pctStr = fmt.Sprintf("%.0f%%", set.Percentage)
	}

	return []string{
		fmt.Sprintf("%d", week),
		fmt.Sprintf("%d", day),
		set.Exercise,
		fmt.Sprintf("%d", set.Sets),
		set.Reps,
		weightStr,
		pctStr,
	}
}
