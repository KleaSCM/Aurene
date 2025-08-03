/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: stats.go
Description: Stats command for displaying comprehensive scheduler performance metrics.
Provides detailed statistics including turnaround time, wait time, context switches,
CPU utilization, and queue information for performance analysis.
*/

package cmd

import (
	"fmt"
	"os"

	"aurene/internal/logger"
	"aurene/state"

	"github.com/spf13/cobra"
)

var (
	statsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Display scheduler performance statistics",
		Long: `Display comprehensive performance statistics for the Aurene scheduler.

Shows detailed metrics including:
• Task turnaround and wait times
• Context switches and preemptions
• CPU utilization and queue lengths
• Task completion rates and IO patterns

Examples:
  aurene stats                    # Show current statistics
  aurene stats --export csv      # Export to CSV format
  aurene stats --detailed        # Show detailed breakdown`,
		RunE: showStats,
	}

	// Flags
	exportFormat string
	detailed     bool
	outputFile   string
)

func init() {
	statsCmd.Flags().StringVarP(&exportFormat, "export", "e", "", "Export format (csv, json)")
	statsCmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed breakdown")
	statsCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")

	rootCmd.AddCommand(statsCmd)
}

/**
 * showStats displays comprehensive scheduler performance metrics
 *
 * Retrieves and formats statistics from the scheduler state,
 * including task performance, system utilization, and queue metrics.
 * Supports export to various formats for analysis.
 */
func showStats(cmd *cobra.Command, args []string) error {
	logger := logger.New()
	logger.Info("Retrieving Aurene scheduler statistics...")

	// Load scheduler state
	stateManager := state.NewStateManager()
	stats, err := stateManager.LoadStats()
	if err != nil {
		logger.Error("Failed to load statistics: %v", err)
		return err
	}

	if stats == nil {
		logger.Info("No statistics available. Start the scheduler with 'aurene run' to collect data.")
		return nil
	}

	// Display statistics
	displayStats(stats, detailed, logger)

	// Export if requested
	if exportFormat != "" {
		if err := exportStats(stats, exportFormat, outputFile, logger); err != nil {
			return err
		}
	}

	return nil
}

/**
 * displayStats formats and displays scheduler statistics
 *
 * Shows comprehensive metrics including task performance,
 * system utilization, and queue information in a readable format.
 */
func displayStats(stats *state.SchedulerStats, detailed bool, logger *logger.Logger) {
	logger.Info("🌌 Aurene Scheduler Statistics")
	logger.Info("==================================================")

	// Basic metrics
	logger.Info("📊 Performance Metrics:")
	logger.Info("  Total Ticks: %d", stats.TotalTicks)
	logger.Info("  Context Switches: %d", stats.ContextSwitches)
	logger.Info("  Preemptions: %d", stats.Preemptions)
	logger.Info("  CPU Utilization: %.2f%%", stats.CPUUtilization)

	// Task statistics
	logger.Info("📋 Task Statistics:")
	logger.Info("  Total Tasks: %d", stats.TotalTasks)
	logger.Info("  Finished Tasks: %d", stats.FinishedTasks)
	logger.Info("  Blocked Tasks: %d", stats.BlockedTasks)
	logger.Info("  Average Turnaround Time: %.2f ticks", stats.AverageTurnaroundTime)
	logger.Info("  Average Wait Time: %.2f ticks", stats.AverageWaitTime)

	// Queue information
	logger.Info("🎯 Queue Information:")
	for i, length := range stats.QueueLengths {
		logger.Info("  Queue %d: %d tasks", i, length)
	}

	if detailed {
		logger.Info("🔍 Detailed Breakdown:")
		logger.Info("  Engine Uptime: %v", stats.EngineUptime)
		logger.Info("  Tasks per Second: %.2f", stats.TasksPerSecond)
		logger.Info("  IO Block Rate: %.2f%%", stats.IOBlockRate)
		logger.Info("  Priority Aging Events: %d", stats.PriorityAgingEvents)
	}

	logger.Info("==================================================")
}

/**
 * exportStats exports statistics to various formats
 *
 * Supports CSV and JSON export formats for external analysis
 * and integration with monitoring systems.
 */
func exportStats(stats *state.SchedulerStats, format, outputFile string, logger *logger.Logger) error {
	var data []byte
	var err error

	switch format {
	case "csv":
		data, err = stats.ToCSV()
	case "json":
		data, err = stats.ToJSON()
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		logger.Error("Failed to export statistics: %v", err)
		return err
	}

	// Write to file or stdout
	if outputFile != "" {
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			logger.Error("Failed to write output file: %v", err)
			return err
		}
		logger.Info("Statistics exported to: %s", outputFile)
	} else {
		fmt.Print(string(data))
	}

	return nil
}
