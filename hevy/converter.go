package hevy

import (
	"fmt"
	"strconv"
	"strings"

	"lifting/program"
)

// ConvertDayToRoutine converts a program Day to a Hevy CreateRoutineRequest
func ConvertDayToRoutine(day program.Day, mapper *ExerciseMapper) (*CreateRoutineRequest, error) {
	title := fmt.Sprintf("531 BBB W%dD%d - %s", day.Week, day.DayNum, day.MainLift)

	exercises := []RoutineExercise{}
	currentExercise := ""
	var currentRoutineExercise *RoutineExercise

	for _, set := range day.Sets {
		// If we've moved to a new exercise, save the previous one and start a new one
		if set.Exercise != currentExercise {
			if currentRoutineExercise != nil {
				exercises = append(exercises, *currentRoutineExercise)
			}

			template, err := mapper.FindTemplate(set.Exercise)
			if err != nil {
				return nil, fmt.Errorf("failed to find template for %s: %w", set.Exercise, err)
			}

			currentRoutineExercise = &RoutineExercise{
				ExerciseTemplateID: template.ID,
				Sets:               []RoutineSet{},
			}
			currentExercise = set.Exercise
		}

		// Convert sets
		routineSets := convertSets(set)
		currentRoutineExercise.Sets = append(currentRoutineExercise.Sets, routineSets...)
	}

	// Don't forget the last exercise
	if currentRoutineExercise != nil {
		exercises = append(exercises, *currentRoutineExercise)
	}

	return &CreateRoutineRequest{
		Title:     title,
		Exercises: exercises,
	}, nil
}

// convertSets converts a program.Set to one or more RoutineSets
func convertSets(set program.Set) []RoutineSet {
	var sets []RoutineSet

	// Determine set type based on percentage (warmup sets are <=60%)
	setType := SetTypeNormal
	if set.Percentage > 0 && set.Percentage <= 60 {
		setType = SetTypeWarmup
	}

	// Check if this is an AMRAP set (ends with "+")
	isAMRAP := strings.HasSuffix(set.Reps, "+")
	repsStr := strings.TrimSuffix(set.Reps, "+")
	reps, _ := strconv.Atoi(repsStr)

	// Create the appropriate number of sets
	for i := 0; i < set.Sets; i++ {
		routineSet := RoutineSet{
			Type: setType,
		}

		// Only add weight if it's specified (accessories have 0 weight)
		if set.Weight > 0 {
			weightKg := LbsToKg(set.Weight)
			routineSet.WeightKg = &weightKg
		}

		if reps > 0 {
			if isAMRAP {
				// Use rep range for AMRAP sets: min reps to ~double
				endReps := reps * 2
				if endReps < 10 {
					endReps = 10
				}
				routineSet.RepRange = &RepRange{
					Start: &reps,
					End:   &endReps,
				}
			} else {
				routineSet.Reps = &reps
			}
		}

		sets = append(sets, routineSet)
	}

	return sets
}

// ConvertProgramToRoutines converts an entire program to Hevy routines
func ConvertProgramToRoutines(prog *program.Program, mapper *ExerciseMapper) ([]CreateRoutineRequest, error) {
	var routines []CreateRoutineRequest

	for _, day := range prog.Days {
		routine, err := ConvertDayToRoutine(day, mapper)
		if err != nil {
			return nil, err
		}
		routines = append(routines, *routine)
	}

	return routines, nil
}
