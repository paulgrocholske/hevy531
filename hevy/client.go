package hevy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseURL = "https://api.hevyapp.com/v1"

// Client is a Hevy API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Hevy API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// ExerciseTemplate represents a Hevy exercise template
type ExerciseTemplate struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Type              string `json:"type"`
	PrimaryMuscleGroup string `json:"primary_muscle_group"`
	IsCustom          bool   `json:"is_custom"`
}

// ExerciseTemplatesResponse is the response from GET /exercise_templates
type ExerciseTemplatesResponse struct {
	PageCount         int                `json:"page_count"`
	ExerciseTemplates []ExerciseTemplate `json:"exercise_templates"`
}

// SetType represents the type of a set
type SetType string

const (
	SetTypeWarmup  SetType = "warmup"
	SetTypeNormal  SetType = "normal"
	SetTypeFailure SetType = "failure"
	SetTypeDropset SetType = "dropset"
)

// RepRange represents a rep range for AMRAP sets
type RepRange struct {
	Start *int `json:"start,omitempty"`
	End   *int `json:"end,omitempty"`
}

// RoutineSet represents a set in a routine exercise
type RoutineSet struct {
	Type     SetType   `json:"type,omitempty"`
	WeightKg *float64  `json:"weight_kg,omitempty"`
	Reps     *int      `json:"reps,omitempty"`
	RepRange *RepRange `json:"rep_range,omitempty"`
}

// RoutineExercise represents an exercise in a routine
type RoutineExercise struct {
	ExerciseTemplateID string       `json:"exercise_template_id"`
	SupersetID         *int         `json:"superset_id,omitempty"`
	RestSeconds        *int         `json:"rest_seconds,omitempty"`
	Notes              *string      `json:"notes,omitempty"`
	Sets               []RoutineSet `json:"sets"`
}

// CreateRoutineRequest is the request body for POST/PUT /routines
type CreateRoutineRequest struct {
	Title     string            `json:"title"`
	FolderID  *int              `json:"folder_id,omitempty"` // omit for updates, include for creates
	Notes     *string           `json:"notes,omitempty"`
	Exercises []RoutineExercise `json:"exercises"`
}

// Routine represents a created routine
type Routine struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// CreateRoutineResponse is the response from POST /routines
type CreateRoutineResponse struct {
	Routine Routine `json:"routine"`
}

// Folder represents a routine folder
type Folder struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// CreateFolderRequest is the request body for POST /routine_folders
type CreateFolderRequest struct {
	Title string `json:"title"`
}

// RoutineFull represents a routine with all details from GET /routines
type RoutineFull struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	FolderID *int   `json:"folder_id"`
}

// RoutinesResponse is the response from GET /routines
type RoutinesResponse struct {
	PageCount int           `json:"page_count"`
	Routines  []RoutineFull `json:"routines"`
}

// FoldersResponse is the response from GET /routine_folders
type FoldersResponse struct {
	PageCount      int      `json:"page_count"`
	RoutineFolders []Folder `json:"routine_folders"`
}

// GetExerciseTemplates fetches all exercise templates (paginated)
func (c *Client) GetExerciseTemplates() ([]ExerciseTemplate, error) {
	var allTemplates []ExerciseTemplate
	page := 1

	for {
		url := fmt.Sprintf("%s/exercise_templates?page=%d&pageSize=100", baseURL, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("api-key", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch exercise templates: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		var result ExerciseTemplatesResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allTemplates = append(allTemplates, result.ExerciseTemplates...)

		if page >= result.PageCount {
			break
		}
		page++
	}

	return allTemplates, nil
}

// CreateRoutine creates a new routine
func (c *Client) CreateRoutine(routine CreateRoutineRequest) (*Routine, error) {
	url := fmt.Sprintf("%s/routines", baseURL)

	// API expects the routine wrapped in a "routine" key
	wrapper := map[string]CreateRoutineRequest{"routine": routine}
	body, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create routine: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// The API returns {"routine": {...}} on success
	var result struct {
		Routine Routine `json:"routine"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// If parsing fails, the routine was still created - return a placeholder
		return &Routine{Title: routine.Title}, nil
	}

	return &result.Routine, nil
}

// CreateFolder creates a new routine folder
func (c *Client) CreateFolder(title string) (*Folder, error) {
	url := fmt.Sprintf("%s/routine_folders", baseURL)

	// API expects the folder wrapped in a "routine_folder" key
	wrapper := map[string]CreateFolderRequest{"routine_folder": {Title: title}}
	body, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response to get folder ID
	var result struct {
		RoutineFolder Folder `json:"routine_folder"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode folder response: %w", err)
	}

	return &result.RoutineFolder, nil
}

// GetFolders fetches all routine folders
func (c *Client) GetFolders() ([]Folder, error) {
	var allFolders []Folder
	page := 1

	for {
		url := fmt.Sprintf("%s/routine_folders?page=%d&pageSize=10", baseURL, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("api-key", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch folders: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		var result FoldersResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allFolders = append(allFolders, result.RoutineFolders...)

		if page >= result.PageCount {
			break
		}
		page++
	}

	return allFolders, nil
}

// GetRoutines fetches all routines
func (c *Client) GetRoutines() ([]RoutineFull, error) {
	var allRoutines []RoutineFull
	page := 1

	for {
		url := fmt.Sprintf("%s/routines?page=%d&pageSize=10", baseURL, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("api-key", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch routines: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		var result RoutinesResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allRoutines = append(allRoutines, result.Routines...)

		if page >= result.PageCount {
			break
		}
		page++
	}

	return allRoutines, nil
}

// UpdateRoutine updates an existing routine
func (c *Client) UpdateRoutine(routineID string, routine CreateRoutineRequest) (*Routine, error) {
	url := fmt.Sprintf("%s/routines/%s", baseURL, routineID)

	// API expects the routine wrapped in a "routine" key
	wrapper := map[string]CreateRoutineRequest{"routine": routine}
	body, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update routine: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Routine Routine `json:"routine"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return &Routine{Title: routine.Title}, nil
	}

	return &result.Routine, nil
}

// exerciseAliases maps our exercise names to Hevy's expected names
var exerciseAliases = map[string][]string{
	// Main lifts
	"squat":          {"barbell squat", "squat (barbell)"},
	"bench press":    {"barbell bench press", "bench press (barbell)"},
	"deadlift":       {"barbell deadlift", "deadlift (barbell)"},
	"overhead press": {"overhead press (barbell)", "barbell overhead press", "shoulder press (barbell)"},

	// Accessories
	"barbell row":         {"bent over row (barbell)", "barbell bent over row", "bent over row"},
	"dumbbell press":      {"dumbbell bench press", "bench press (dumbbell)", "dumbbell chest press"},
	"dumbbell row":        {"dumbbell row", "bent over row (dumbbell)", "one arm dumbbell row"},
	"leg curl":            {"lying leg curl", "leg curl (machine)", "seated leg curl"},
	"leg press":           {"leg press (machine)", "leg press"},
	"tricep pushdown":     {"tricep pushdown", "triceps pushdown", "cable pushdown"},
	"cable fly":           {"cable fly", "cable chest fly", "cable crossover"},
	"good morning":        {"good morning", "good morning (barbell)"},
	"hanging leg raise":   {"hanging leg raise", "hanging knee raise"},
	"back extension":      {"back extension", "hyperextension", "back extension (machine)"},
	"lateral raise":       {"lateral raise (dumbbell)", "dumbbell lateral raise", "lateral raise"},
	"face pull":           {"face pull", "face pull (cable)"},
	"rear delt fly":       {"reverse fly (dumbbell)", "rear delt fly", "reverse fly"},
	"pull-up":             {"pull up", "pull-up", "pullup"},
	"dips":                {"dip", "tricep dip", "chest dip"},
	"lunges":              {"lunge (dumbbell)", "walking lunge", "lunge (barbell)"},
	"bulgarian split squat": {"bulgarian split squat", "split squat"},
}

// ExerciseMapper helps map exercise names to Hevy template IDs
type ExerciseMapper struct {
	templates map[string]ExerciseTemplate // lowercase title -> template
}

// NewExerciseMapper creates a mapper from a list of templates
func NewExerciseMapper(templates []ExerciseTemplate) *ExerciseMapper {
	m := &ExerciseMapper{
		templates: make(map[string]ExerciseTemplate),
	}
	for _, t := range templates {
		m.templates[strings.ToLower(t.Title)] = t
	}
	return m
}

// FindTemplate finds a template by name (case-insensitive, with aliases)
func (m *ExerciseMapper) FindTemplate(name string) (*ExerciseTemplate, error) {
	lower := strings.ToLower(name)

	// Try exact match first
	if t, ok := m.templates[lower]; ok {
		return &t, nil
	}

	// Try aliases
	if aliases, ok := exerciseAliases[lower]; ok {
		for _, alias := range aliases {
			if t, ok := m.templates[alias]; ok {
				return &t, nil
			}
		}
	}

	// Try partial match as fallback
	for title, t := range m.templates {
		if strings.Contains(title, lower) || strings.Contains(lower, title) {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("no template found for exercise: %s", name)
}

// LbsToKg converts pounds to kilograms
func LbsToKg(lbs float64) float64 {
	return lbs * 0.453592
}
