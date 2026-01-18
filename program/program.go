package program

import (
	"lifting/config"
)

// Set represents a single set in the program
type Set struct {
	Exercise   string
	Sets       int
	Reps       string // string to support "5+" notation
	Weight     float64
	Percentage float64
}

// Day represents a training day
type Day struct {
	Week     int
	DayNum   int
	MainLift config.Lift
	Sets     []Set
}

// Program represents the full 4-week program
type Program struct {
	Days []Day
}

// WeekScheme defines the percentages and reps for a week's main lift work
type WeekScheme struct {
	Percentages []float64
	Reps        []string
}

// Warmup scheme: 40% x5, 50% x5, 60% x3
var WarmupScheme = WeekScheme{
	Percentages: []float64{40, 50, 60},
	Reps:        []string{"5", "5", "3"},
}

// Working set schemes for weeks 1-3
var WorkingSchemes = map[int]WeekScheme{
	1: {
		Percentages: []float64{65, 75, 85},
		Reps:        []string{"5", "5", "5+"},
	},
	2: {
		Percentages: []float64{70, 80, 90},
		Reps:        []string{"3", "3", "3+"},
	},
	3: {
		Percentages: []float64{75, 85, 95},
		Reps:        []string{"5", "3", "1+"},
	},
}

// Deload scheme (week 4): 40% x5, 50% x5, 60% x5
var DeloadScheme = WeekScheme{
	Percentages: []float64{40, 50, 60},
	Reps:        []string{"5", "5", "5"},
}

// Generate creates a full 4-week 5/3/1 BBB program from the given config
func Generate(cfg *config.Config) *Program {
	program := &Program{
		Days: make([]Day, 0, 16), // 4 weeks x 4 days
	}

	for week := 1; week <= 4; week++ {
		for dayIdx, mainLift := range cfg.LiftOrder {
			day := Day{
				Week:     week,
				DayNum:   dayIdx + 1,
				MainLift: mainLift,
				Sets:     make([]Set, 0),
			}

			trainingMax := cfg.TrainingMaxes[mainLift]

			// Add sets based on whether it's a deload week
			if week == 4 {
				// Deload week - just the deload sets (no warmup, they're the same)
				day.Sets = append(day.Sets, generateMainSets(mainLift, trainingMax, DeloadScheme)...)
			} else {
				// Regular week - warmup + working sets
				day.Sets = append(day.Sets, generateMainSets(mainLift, trainingMax, WarmupScheme)...)
				day.Sets = append(day.Sets, generateMainSets(mainLift, trainingMax, WorkingSchemes[week])...)
			}

			// BBB sets (5x10 at configured percentage)
			bbbLift := cfg.BBBPairing[mainLift]
			bbbTrainingMax := cfg.TrainingMaxes[bbbLift]
			day.Sets = append(day.Sets, generateBBBSets(bbbLift, bbbTrainingMax, cfg.BBBPercentage)...)

			// Accessory (5x10, no weight)
			if accessory, ok := cfg.Accessories[mainLift]; ok && accessory != "" {
				day.Sets = append(day.Sets, Set{
					Exercise:   accessory,
					Sets:       5,
					Reps:       "10",
					Weight:     0,
					Percentage: 0,
				})
			}

			program.Days = append(program.Days, day)
		}
	}

	return program
}

// generateMainSets creates sets for main lift work (warmup or working sets)
func generateMainSets(lift config.Lift, trainingMax float64, scheme WeekScheme) []Set {
	sets := make([]Set, len(scheme.Percentages))
	for i, pct := range scheme.Percentages {
		weight := config.RoundToNearest5(trainingMax * pct / 100)
		sets[i] = Set{
			Exercise:   string(lift),
			Sets:       1,
			Reps:       scheme.Reps[i],
			Weight:     weight,
			Percentage: pct,
		}
	}
	return sets
}

// generateBBBSets creates the 5x10 BBB sets
func generateBBBSets(lift config.Lift, trainingMax float64, percentage float64) []Set {
	weight := config.RoundToNearest5(trainingMax * percentage / 100)
	return []Set{
		{
			Exercise:   string(lift),
			Sets:       5,
			Reps:       "10",
			Weight:     weight,
			Percentage: percentage,
		},
	}
}
