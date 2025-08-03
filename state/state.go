/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: state.go
Description: State management for the Aurene scheduler including statistics persistence,
loading, and export functionality. Provides comprehensive performance metrics
tracking and data export capabilities.
*/

package state

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/**
 * SchedulerStats represents comprehensive scheduler performance metrics
 *
 * Tracks all key performance indicators including task metrics, system utilization,
 * queue statistics, and timing information for analysis and monitoring.
 */
type SchedulerStats struct {
	// Basic metrics
	TotalTicks      int64   `json:"total_ticks"`
	ContextSwitches int64   `json:"context_switches"`
	Preemptions     int64   `json:"preemptions"`
	CPUUtilization  float64 `json:"cpu_utilization"`

	// Task statistics
	TotalTasks            int     `json:"total_tasks"`
	FinishedTasks         int     `json:"finished_tasks"`
	BlockedTasks          int     `json:"blocked_tasks"`
	AverageTurnaroundTime float64 `json:"average_turnaround_time"`
	AverageWaitTime       float64 `json:"average_wait_time"`

	// Queue information
	QueueLengths []int `json:"queue_lengths"`

	// Detailed metrics
	EngineUptime        time.Duration `json:"engine_uptime"`
	TasksPerSecond      float64       `json:"tasks_per_second"`
	IOBlockRate         float64       `json:"io_block_rate"`
	PriorityAgingEvents int64         `json:"priority_aging_events"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

/**
 * StateManager handles scheduler state persistence and loading
 *
 * Provides functionality to save and load scheduler statistics,
 * supporting multiple export formats and data persistence.
 */
type StateManager struct {
	statsFile string
}

/**
 * NewStateManager creates a new state manager instance
 *
 * Initializes the state manager with default file paths
 * for statistics persistence and loading.
 */
func NewStateManager() *StateManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	statsFile := filepath.Join(homeDir, ".aurene", "stats.json")
	return &StateManager{
		statsFile: statsFile,
	}
}

/**
 * LoadStats retrieves saved scheduler statistics
 *
 * Loads statistics from persistent storage, returning nil
 * if no statistics are available or on error.
 */
func (sm *StateManager) LoadStats() (*SchedulerStats, error) {
	data, err := os.ReadFile(sm.statsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats SchedulerStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	return &stats, nil
}

/**
 * SaveStats persists scheduler statistics to storage
 *
 * Saves current statistics to persistent storage for later
 * retrieval and analysis.
 */
func (sm *StateManager) SaveStats(stats *SchedulerStats) error {
	stats.UpdatedAt = time.Now()

	// Ensure directory exists
	dir := filepath.Dir(sm.statsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create stats directory: %w", err)
	}

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	if err := os.WriteFile(sm.statsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write stats file: %w", err)
	}

	return nil
}

/**
 * ClearStats removes all saved statistics
 *
 * Deletes the statistics file and clears all persisted data.
 */
func (sm *StateManager) ClearStats() error {
	if err := os.Remove(sm.statsFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stats file: %w", err)
	}
	return nil
}

/**
 * ToCSV exports statistics to CSV format
 *
 * Converts scheduler statistics to CSV format for external
 * analysis and spreadsheet integration.
 */
func (sm *SchedulerStats) ToCSV() ([]byte, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// Write headers
	headers := []string{
		"Metric", "Value", "Unit",
	}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Write data rows
	rows := [][]string{
		{"Total Ticks", fmt.Sprintf("%d", sm.TotalTicks), "ticks"},
		{"Context Switches", fmt.Sprintf("%d", sm.ContextSwitches), "switches"},
		{"Preemptions", fmt.Sprintf("%d", sm.Preemptions), "preemptions"},
		{"CPU Utilization", fmt.Sprintf("%.2f", sm.CPUUtilization), "%"},
		{"Total Tasks", fmt.Sprintf("%d", sm.TotalTasks), "tasks"},
		{"Finished Tasks", fmt.Sprintf("%d", sm.FinishedTasks), "tasks"},
		{"Blocked Tasks", fmt.Sprintf("%d", sm.BlockedTasks), "tasks"},
		{"Average Turnaround Time", fmt.Sprintf("%.2f", sm.AverageTurnaroundTime), "ticks"},
		{"Average Wait Time", fmt.Sprintf("%.2f", sm.AverageWaitTime), "ticks"},
		{"Tasks per Second", fmt.Sprintf("%.2f", sm.TasksPerSecond), "tasks/sec"},
		{"IO Block Rate", fmt.Sprintf("%.2f", sm.IOBlockRate), "%"},
		{"Priority Aging Events", fmt.Sprintf("%d", sm.PriorityAgingEvents), "events"},
		{"Engine Uptime", sm.EngineUptime.String(), "duration"},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(builder.String()), writer.Error()
}

/**
 * ToJSON exports statistics to JSON format
 *
 * Converts scheduler statistics to JSON format for API
 * integration and programmatic analysis.
 */
func (sm *SchedulerStats) ToJSON() ([]byte, error) {
	return json.MarshalIndent(sm, "", "  ")
}
