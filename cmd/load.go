/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: load.go
Description: CLI command for loading external task definitions from files.
Enables Aurene scheduler to integrate with real systems by loading task
definitions from TOML, JSON, and CSV files for production deployment.
*/

package cmd

import (
	"fmt"
	"time"

	"aurene/internal/logger"
	"aurene/runtime"
	"aurene/scheduler"
	"aurene/task"
	"aurene/workloads"

	"github.com/spf13/cobra"
)

var (
	loadFile     string
	loadFormat   string
	loadRealTime bool
	loadBatch    int
)

/**
 * loadCmd represents the load command for external task files
 *
 * Enables loading task definitions from external files to demonstrate
 * real system integration capabilities of the Aurene scheduler.
 */
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load tasks from external file",
	Long: `Load task definitions from external files (TOML, JSON, CSV) to test
Aurene scheduler with real workload definitions.

Example:
  aurene load --file tasks.toml
  aurene load --file workload.json --format json
  aurene load --file tasks.csv --format csv --batch 1000`,
	RunE: runLoad,
}

func init() {
	rootCmd.AddCommand(loadCmd)

	loadCmd.Flags().StringVarP(&loadFile, "file", "f", "", "Task definition file to load")
	loadCmd.Flags().StringVar(&loadFormat, "format", "", "File format (toml, json, csv)")
	loadCmd.Flags().BoolVar(&loadRealTime, "realtime", false, "Enable real-time task injection")
	loadCmd.Flags().IntVar(&loadBatch, "batch", 100, "Batch size for task processing")

	loadCmd.MarkFlagRequired("file")
}

/**
 * runLoad executes the load command
 *
 * Loads task definitions from external files and demonstrates
 * real system integration capabilities with comprehensive logging.
 */
func runLoad(cmd *cobra.Command, args []string) error {
	logger := logger.New()
	logger.Info("Loading tasks from file: %s", loadFile)

	loader := workloads.NewFileLoader()

	workload, err := loader.LoadWorkloadFromFile(loadFile)
	if err != nil {
		return fmt.Errorf("failed to load workload: %w", err)
	}

	logger.Info("Loaded workload: %s (%d tasks)", workload.Metadata.Name, len(workload.Tasks))

	tasks, err := loader.ConvertToTasks(workload)
	if err != nil {
		return fmt.Errorf("failed to convert tasks: %w", err)
	}

	logger.Info("Converted %d tasks for scheduler", len(tasks))

	sched := scheduler.NewScheduler(5)
	engine := runtime.NewEngine(sched, time.Duration(4)*time.Millisecond)

	engine.SetCallbacks(
		func(tick int64) {
			if tick%1000 == 0 {
				logger.Info("Tick: %d", tick)
			}
		},
		func(event string, t *task.Task) {
			logger.Info("Task event: %s - %s", event, t.String())
		},
	)

	logger.Info("Adding %d tasks to scheduler...", len(tasks))

	startTime := time.Now()
	tasksAdded := 0

	for i, t := range tasks {
		engine.AddTask(t)
		tasksAdded++

		if loadBatch > 0 && (i+1)%loadBatch == 0 {
			logger.Info("Added batch %d/%d tasks", i+1, len(tasks))
		}
	}

	elapsed := time.Since(startTime)
	logger.Info("Added %d tasks in %v (%.1f tasks/sec)", tasksAdded, elapsed, float64(tasksAdded)/elapsed.Seconds())

	if loadRealTime {
		logger.Info("Starting engine in real-time mode...")

		if err := engine.Start(); err != nil {
			return fmt.Errorf("failed to start engine: %w", err)
		}

		time.Sleep(30 * time.Second)
		engine.Stop()

		stats := engine.GetStats()
		logger.Info("Final statistics:")
		logger.Info("  Total ticks: %v", stats["total_ticks"])
		logger.Info("  Context switches: %v", stats["context_switches"])
		logger.Info("  CPU utilization: %.2f%%", stats["cpu_utilization"])
	} else {
		logger.Info("Tasks loaded successfully (use --realtime to run scheduler)")
	}

	return nil
}
