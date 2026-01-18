package config

// Lift represents one of the four main lifts
type Lift string

const (
	Squat    Lift = "Squat"
	Bench    Lift = "Bench Press"
	Deadlift Lift = "Deadlift"
	OHP      Lift = "Overhead Press"
)

// DefaultLiftOrder is the standard 5/3/1 day order
var DefaultLiftOrder = []Lift{Squat, Bench, Deadlift, OHP}

// AllLifts returns all available main lifts
func AllLifts() []Lift {
	return []Lift{Squat, Bench, Deadlift, OHP}
}

// AccessoryPresets maps each main lift to appropriate accessory options
var AccessoryPresets = map[Lift][]string{
	Squat:    {"Leg Curl", "Lunges", "Leg Press", "Bulgarian Split Squat"},
	Bench:    {"Dumbbell Press", "Dumbbell Row", "Dips", "Tricep Pushdown", "Cable Fly"},
	Deadlift: {"Barbell Row", "Good Morning", "Hanging Leg Raise", "Back Extension"},
	OHP:      {"Lateral Raise", "Face Pull", "Rear Delt Fly", "Pull-up"},
}

// LiftMaxes holds the training max for each lift
type LiftMaxes map[Lift]float64

// Config holds all configuration for generating a 5/3/1 BBB program
type Config struct {
	// Training maxes for each lift (already calculated if input was true 1RM)
	TrainingMaxes LiftMaxes

	// Order of lifts for each training day (Day 1 = index 0, etc.)
	LiftOrder []Lift

	// BBB percentage (default 50)
	BBBPercentage float64

	// BBB lift pairing - maps main lift to its BBB lift (same or opposite)
	BBBPairing map[Lift]Lift

	// Selected accessory for each main lift day
	Accessories map[Lift]string
}

// NewDefaultConfig creates a config with sensible defaults
func NewDefaultConfig() *Config {
	liftOrder := DefaultLiftOrder

	// Default BBB pairing is same lift
	bbbPairing := make(map[Lift]Lift)
	for _, lift := range liftOrder {
		bbbPairing[lift] = lift
	}

	return &Config{
		TrainingMaxes: make(LiftMaxes),
		LiftOrder:     liftOrder,
		BBBPercentage: 50.0,
		BBBPairing:    bbbPairing,
		Accessories:   make(map[Lift]string),
	}
}

// CalculateTrainingMax returns 90% of the true 1RM, rounded to nearest 5
func CalculateTrainingMax(true1RM float64) float64 {
	return RoundToNearest5(true1RM * 0.9)
}

// RoundToNearest5 rounds a weight to the nearest 5 lbs
func RoundToNearest5(weight float64) float64 {
	return float64(int((weight+2.5)/5) * 5)
}
