package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"lifting/config"
)

const DefaultFile = ".531bbb_memory.json"

// Snapshot stores the last saved config and timestamp.
type Snapshot struct {
	SavedAt time.Time      `json:"saved_at"`
	Config  *config.Config `json:"config"`
}

// Load reads memory from disk. If no file exists, it returns (nil, nil).
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read memory file: %w", err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse memory file: %w", err)
	}
	if snapshot.Config == nil {
		return nil, fmt.Errorf("memory file is missing config")
	}

	return &snapshot, nil
}

// Save writes config memory to disk.
func Save(path string, cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("cannot save nil config")
	}

	snapshot := Snapshot{
		SavedAt: time.Now().UTC(),
		Config:  CloneConfig(cfg),
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode memory file: %w", err)
	}

	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}

	return nil
}

// CloneConfig creates a deep copy of config.
func CloneConfig(cfg *config.Config) *config.Config {
	if cfg == nil {
		return nil
	}

	cloned := &config.Config{
		TrainingMaxes: make(config.LiftMaxes, len(cfg.TrainingMaxes)),
		LiftOrder:     append([]config.Lift{}, cfg.LiftOrder...),
		BBBPercentage: cfg.BBBPercentage,
		BBBPairing:    make(map[config.Lift]config.Lift, len(cfg.BBBPairing)),
		Accessories:   make(map[config.Lift]string, len(cfg.Accessories)),
	}

	for lift, max := range cfg.TrainingMaxes {
		cloned.TrainingMaxes[lift] = max
	}
	for lift, paired := range cfg.BBBPairing {
		cloned.BBBPairing[lift] = paired
	}
	for lift, accessory := range cfg.Accessories {
		cloned.Accessories[lift] = accessory
	}

	return cloned
}

// NextCycleConfig returns a copy with standard 5/3/1 TM increases applied.
func NextCycleConfig(cfg *config.Config) *config.Config {
	next := CloneConfig(cfg)
	if next == nil {
		return nil
	}

	next.TrainingMaxes[config.Squat] += 10
	next.TrainingMaxes[config.Deadlift] += 10
	next.TrainingMaxes[config.Bench] += 5
	next.TrainingMaxes[config.OHP] += 5

	return next
}
