package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"lifting/config"
)

// Reader handles interactive prompts
type Reader struct {
	scanner *bufio.Scanner
}

// NewReader creates a new prompt reader
func NewReader() *Reader {
	return &Reader{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// readLine reads a line of input from the user
func (r *Reader) readLine() string {
	r.scanner.Scan()
	return strings.TrimSpace(r.scanner.Text())
}

// readFloat reads a float from the user with validation
func (r *Reader) readFloat(prompt string) float64 {
	for {
		fmt.Print(prompt)
		input := r.readLine()
		val, err := strconv.ParseFloat(input, 64)
		if err != nil || val <= 0 {
			fmt.Println("Please enter a valid positive number.")
			continue
		}
		return val
	}
}

// readYesNo reads a yes/no response from the user
func (r *Reader) readYesNo(prompt string) bool {
	for {
		fmt.Print(prompt + " (y/n): ")
		input := strings.ToLower(r.readLine())
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Please enter 'y' or 'n'.")
	}
}

// readChoice reads a numbered choice from the user
func (r *Reader) readChoice(prompt string, options []string) int {
	for {
		fmt.Println(prompt)
		for i, opt := range options {
			fmt.Printf("  %d. %s\n", i+1, opt)
		}
		fmt.Print("Enter choice (1-" + strconv.Itoa(len(options)) + "): ")
		input := r.readLine()
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(options) {
			fmt.Println("Please enter a valid choice.")
			continue
		}
		return choice - 1 // Return 0-indexed
	}
}

// GatherConfig interactively gathers all configuration from the user
func (r *Reader) GatherConfig() (*config.Config, error) {
	cfg := config.NewDefaultConfig()

	fmt.Print("\n=== 5/3/1 BBB Program Generator ===\n\n")

	// Step 1: Get 1RM values and determine if they're true 1RM or training max
	fmt.Println("Enter your max for each lift.")
	isTrueMax := r.readYesNo("Are these your TRUE 1RMs? (I'll calculate training max at 90%)")
	fmt.Println()

	for _, lift := range config.AllLifts() {
		prompt := fmt.Sprintf("Enter %s max (lbs): ", lift)
		maxVal := r.readFloat(prompt)
		if isTrueMax {
			cfg.TrainingMaxes[lift] = config.CalculateTrainingMax(maxVal)
		} else {
			cfg.TrainingMaxes[lift] = maxVal
		}
	}

	if isTrueMax {
		fmt.Println("\nTraining maxes (90% of true 1RM):")
		for _, lift := range config.AllLifts() {
			fmt.Printf("  %s: %.0f lbs\n", lift, cfg.TrainingMaxes[lift])
		}
	}

	// Step 2: Customize lift order
	fmt.Println("\n--- Lift Order ---")
	fmt.Println("Default order: Day 1 = Squat, Day 2 = Bench, Day 3 = Deadlift, Day 4 = OHP")
	if r.readYesNo("Would you like to customize the lift order?") {
		cfg.LiftOrder = r.gatherLiftOrder()
	}

	// Step 3: BBB percentage
	fmt.Println("\n--- BBB Configuration ---")
	fmt.Printf("Default BBB percentage is 50%%.\n")
	if r.readYesNo("Would you like to change the BBB percentage?") {
		cfg.BBBPercentage = r.readFloat("Enter BBB percentage (e.g., 50 for 50%): ")
	}

	// Step 4: BBB pairing
	fmt.Println("\nDefault BBB pairing: same lift (e.g., Squat day does BBB Squats)")
	if r.readYesNo("Would you like to use opposite lift pairing?") {
		cfg.BBBPairing = r.gatherBBBPairing(cfg.LiftOrder)
	}

	// Step 5: Accessories
	fmt.Println("\n--- Accessory Selection ---")
	for _, lift := range cfg.LiftOrder {
		options := config.AccessoryPresets[lift]
		fmt.Printf("\n%s day accessory:\n", lift)
		choice := r.readChoice("Select an accessory:", options)
		cfg.Accessories[lift] = options[choice]
	}

	return cfg, nil
}

// gatherLiftOrder lets the user specify custom lift order
func (r *Reader) gatherLiftOrder() []config.Lift {
	order := make([]config.Lift, 4)
	available := make(map[config.Lift]bool)
	for _, lift := range config.AllLifts() {
		available[lift] = true
	}

	for day := 1; day <= 4; day++ {
		var options []string
		var lifts []config.Lift
		for _, lift := range config.AllLifts() {
			if available[lift] {
				options = append(options, string(lift))
				lifts = append(lifts, lift)
			}
		}
		choice := r.readChoice(fmt.Sprintf("Select lift for Day %d:", day), options)
		selectedLift := lifts[choice]
		order[day-1] = selectedLift
		delete(available, selectedLift)
	}

	return order
}

// gatherBBBPairing lets the user specify BBB lift pairings
func (r *Reader) gatherBBBPairing(liftOrder []config.Lift) map[config.Lift]config.Lift {
	pairing := make(map[config.Lift]config.Lift)

	fmt.Println("\nFor each main lift, select the BBB lift:")
	for _, mainLift := range liftOrder {
		var options []string
		for _, lift := range config.AllLifts() {
			options = append(options, string(lift))
		}
		choice := r.readChoice(fmt.Sprintf("BBB lift for %s day:", mainLift), options)
		pairing[mainLift] = config.AllLifts()[choice]
	}

	return pairing
}

// GetOutputFilename prompts for the output filename
func (r *Reader) GetOutputFilename() string {
	fmt.Print("\nEnter output filename (default: 531_bbb.csv): ")
	filename := r.readLine()
	if filename == "" {
		return "531_bbb.csv"
	}
	if !strings.HasSuffix(filename, ".csv") {
		filename += ".csv"
	}
	return filename
}

// AskHevyUpload asks if the user wants to upload to Hevy
func (r *Reader) AskHevyUpload() bool {
	fmt.Println("\n--- Export Options ---")
	return r.readYesNo("Would you like to upload routines to Hevy?")
}

// GetHevyAPIKey prompts for the Hevy API key
func (r *Reader) GetHevyAPIKey() string {
	fmt.Print("Enter your Hevy API key: ")
	return r.readLine()
}

// ReadString reads a single line of input (exported for general use)
func (r *Reader) ReadString(prompt string) string {
	fmt.Print(prompt)
	return r.readLine()
}
