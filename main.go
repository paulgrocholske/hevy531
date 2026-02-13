package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"lifting/config"
	"lifting/export"
	"lifting/hevy"
	"lifting/memory"
	"lifting/program"
	"lifting/prompt"
)

func main() {
	reader := prompt.NewReader()

	// Gather configuration, optionally using saved memory
	snapshot, err := memory.Load(memory.DefaultFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load memory file: %v\n", err)
	}

	cfg, err := gatherConfig(reader, snapshot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gathering config: %v\n", err)
		os.Exit(1)
	}

	// Generate the program
	prog := program.Generate(cfg)

	// Ask about Hevy upload
	if reader.AskHevyUpload() {
		if err := uploadToHevy(reader, prog); err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading to Hevy: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Export to CSV
		filename := reader.GetOutputFilename()
		if err := export.ToCSV(prog, filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting CSV: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nProgram exported to %s\n", filename)
	}

	if reader.AskSaveMemory() {
		if err := memory.Save(memory.DefaultFile, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save memory: %v\n", err)
		} else {
			fmt.Printf("Saved program memory to %s\n", memory.DefaultFile)
		}
	}

	fmt.Println("\nHappy lifting!")
}

func gatherConfig(reader *prompt.Reader, snapshot *memory.Snapshot) (*config.Config, error) {
	if snapshot == nil {
		return reader.GatherConfig()
	}

	fmt.Printf("\nFound saved configuration from %s\n", snapshot.SavedAt.Local().Format(time.RFC1123))
	printTrainingMaxes(snapshot.Config.TrainingMaxes)

	switch reader.ChooseConfigStartMode() {
	case prompt.ConfigStartReuseSaved:
		fmt.Println("\nUsing saved configuration.")
		return memory.CloneConfig(snapshot.Config), nil
	case prompt.ConfigStartNextCycle:
		fmt.Println("\nApplying standard 5/3/1 training max increases for next cycle...")
		next := memory.NextCycleConfig(snapshot.Config)
		printTrainingMaxes(next.TrainingMaxes)
		return next, nil
	default:
		return reader.GatherConfig()
	}
}

func printTrainingMaxes(maxes config.LiftMaxes) {
	fmt.Println("Training maxes:")
	for _, lift := range config.AllLifts() {
		fmt.Printf("  %s: %.0f lbs\n", lift, maxes[lift])
	}
}

func uploadToHevy(reader *prompt.Reader, prog *program.Program) error {
	apiKey := reader.GetHevyAPIKey()
	client := hevy.NewClient(apiKey)

	// Fetch exercise templates
	fmt.Println("\nFetching exercise templates from Hevy...")
	templates, err := client.GetExerciseTemplates()
	if err != nil {
		return fmt.Errorf("failed to fetch exercise templates: %w", err)
	}
	fmt.Printf("Found %d exercise templates\n", len(templates))

	// Create mapper
	mapper := hevy.NewExerciseMapper(templates)

	// Convert program to Hevy routines
	fmt.Println("\nConverting program to Hevy routines...")
	routines, err := hevy.ConvertProgramToRoutines(prog, mapper)
	if err != nil {
		return fmt.Errorf("failed to convert program: %w", err)
	}

	// Fetch existing folders and routines
	fmt.Println("\nFetching existing folders and routines...")
	existingFolders, err := client.GetFolders()
	if err != nil {
		return fmt.Errorf("failed to fetch folders: %w", err)
	}
	existingRoutines, err := client.GetRoutines()
	if err != nil {
		return fmt.Errorf("failed to fetch routines: %w", err)
	}

	// Build lookup maps
	folderByName := make(map[string]int) // folder title -> ID
	for _, f := range existingFolders {
		folderByName[f.Title] = f.ID
	}
	routineByTitle := make(map[string]string) // routine title -> ID
	for _, r := range existingRoutines {
		routineByTitle[r.Title] = r.ID
	}

	// Get or create folders for each week
	fmt.Println("\nSetting up weekly folders...")
	weekFolders := make(map[int]int) // week number -> folder ID
	for week := 1; week <= 4; week++ {
		folderName := fmt.Sprintf("531 BBB Week %d", week)
		if folderID, exists := folderByName[folderName]; exists {
			weekFolders[week] = folderID
			fmt.Printf("  Found existing folder: %s\n", folderName)
		} else {
			folder, err := client.CreateFolder(folderName)
			if err != nil {
				return fmt.Errorf("failed to create folder %s: %w", folderName, err)
			}
			weekFolders[week] = folder.ID
			fmt.Printf("  Created folder: %s\n", folderName)
		}
	}

	// Upload or update each routine (with rate limiting and retry)
	fmt.Printf("\nSyncing %d routines to Hevy...\n", len(routines))
	created, updated := 0, 0
	for i, routine := range routines {
		// Determine week from routine index (4 routines per week)
		week := (i / 4) + 1
		folderID := weekFolders[week]
		routine.FolderID = &folderID

		var err error
		var isUpdate bool
		existingID, exists := routineByTitle[routine.Title]

		// Retry with exponential backoff for rate limits
		for attempt := 0; attempt < 5; attempt++ {
			if attempt > 0 {
				delay := time.Duration(10*(1<<attempt)) * time.Second // 20s, 40s, 80s, 160s
				fmt.Printf("    Rate limited, waiting %v...\n", delay)
				time.Sleep(delay)
			}

			if exists {
				// Update existing routine (folder_id not allowed in updates)
				updateRoutine := routine
				updateRoutine.FolderID = nil
				_, err = client.UpdateRoutine(existingID, updateRoutine)
				isUpdate = true
			} else {
				// Create new routine
				_, err = client.CreateRoutine(routine)
				isUpdate = false
			}

			if err == nil || !isRateLimitError(err) {
				break
			}
		}

		if err != nil {
			return fmt.Errorf("failed to sync routine %s: %w", routine.Title, err)
		}

		if isUpdate {
			fmt.Printf("  [%d/%d] Updated: %s\n", i+1, len(routines), routine.Title)
			updated++
		} else {
			fmt.Printf("  [%d/%d] Created: %s\n", i+1, len(routines), routine.Title)
			created++
		}

		// Small delay between requests to avoid rate limits
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Printf("\nSync complete! Created: %d, Updated: %d\n", created, updated)
	return nil
}

// isRateLimitError checks if an error is a rate limit error (429)
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit")
}
