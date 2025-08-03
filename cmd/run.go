/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: run.go
Description: Run command that starts the Aurene scheduler with configurable parameters.
Provides real-time operation, task injection capabilities, and live monitoring
of scheduler performance and task execution.
*/

package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"aurene/internal/constants"
	"aurene/internal/logger"
	"aurene/runtime"
	"aurene/scheduler"
	"aurene/task"

	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Start the Aurene CPU scheduler",
		Long: `Start the Aurene CPU scheduler with real-time operation.

The scheduler will run at 250Hz by default and can be configured with various
options. Tasks can be injected dynamically and performance metrics are tracked.

Examples:
  aurene run                    # Start with default settings
  aurene run --tick-rate 500    # Start at 500Hz
  aurene run --queues 5         # Use 5 priority queues
  aurene run --demo             # Run with demo workload`,
		RunE: runScheduler,
	}

	tickRate   int
	numQueues  int
	demoMode   bool
	configFile string
)

func init() {
	runCmd.Flags().IntVarP(&tickRate, "tick-rate", "t", 250, "Scheduler tick rate in Hz")
	runCmd.Flags().IntVarP(&numQueues, "queues", "q", 3, "Number of priority queues")
	runCmd.Flags().BoolVarP(&demoMode, "demo", "d", false, "Run with demo workload")
	runCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path")

	rootCmd.AddCommand(runCmd)
}

/**
 * runScheduler initializes and starts the Aurene CPU scheduler
 *
 * Creates the MLFQ scheduler with configurable queues and tick rate,
 * sets up real-time monitoring callbacks, and manages graceful shutdown.
 * Implements the core CLI functionality for scheduler operation.
 */
func runScheduler(cmd *cobra.Command, args []string) error {
	logger := logger.New()
	logger.Info("Starting Aurene - The Crystalline Scheduler")

	/**
	 * Calculate tick duration from frequency using the formula:
	 * duration = 1 / frequency = 1,000,000,000ns / frequency_hz
	 * This converts Hz to nanoseconds for precise timing control.
	 */
	tickDuration := time.Duration(1000000000/tickRate) * time.Nanosecond
	logger.Info("Configuration: %d queues, %dHz tick rate (%v per tick)",
		numQueues, tickRate, tickDuration)

	sched := scheduler.NewScheduler(numQueues)

	engine := runtime.NewEngine(sched, tickDuration)

	if demoMode {
		setupDemoWorkload(engine, logger)
	}

	engine.SetCallbacks(
		func(tick int64) {
			if tick%1000 == 0 {
				currentTask := engine.GetCurrentTask()
				if currentTask != nil {
					logger.Info("Tick %d: Running %s (Remaining: %d)",
						tick, currentTask.Name, currentTask.GetRemaining())
				}
			}
		},
		func(event string, t *task.Task) {
			switch event {
			case "finish":
				logger.Info("🎉 Task completed: %s", t.Name)
			case "block":
				logger.Info("⏳ Task blocked: %s", t.Name)
			}
		},
	)

	if err := engine.Start(); err != nil {
		logger.Error("Failed to start engine: %v", err)
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Aurene scheduler is now running! Press Ctrl+C to stop.")
	logger.Info("Use 'aurene add' in another terminal to inject tasks")

	ticker := time.NewTicker(time.Duration(constants.StatusUpdateIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			logger.Info("Received shutdown signal, stopping scheduler...")
			engine.Stop()

			stats := engine.GetStats()
			logger.Info("Final Statistics:")
			logger.Info("  Total Ticks: %v", stats["total_ticks"])
			logger.Info("  Context Switches: %v", stats["context_switches"])
			logger.Info("  Preemptions: %v", stats["preemptions"])
			logger.Info("  Finished Tasks: %v", stats["finished_tasks"])
			logger.Info("  Uptime: %v", stats["engine_uptime"])

			return nil

		case <-ticker.C:
			if engine.IsRunning() {
				stats := engine.GetStats()
				currentTask := engine.GetCurrentTask()

				var taskName string
				if currentTask != nil {
					taskName = currentTask.Name
				} else {
					taskName = "None"
				}

				logger.Info("Status: Running %s | CPU: %.1f%% | Ticks: %v",
					taskName, stats["cpu_utilization"], stats["total_ticks"])
			}
		}
	}
}

/**
 * setupDemoWorkload creates a realistic test workload for scheduler validation
 *
 * Injects diverse task types with varying priorities, durations, and IO patterns
 * to test MLFQ behavior under realistic conditions. Tasks simulate real
 * applications like browsers, terminals, and system utilities.
 */
func setupDemoWorkload(engine *runtime.Engine, logger *logger.Logger) {
	logger.Info("Setting up demo workload...")

	tasks := []struct {
		name     string
		duration int64
		priority int
		ioChance float64
		memory   int64
		group    string
	}{
		{"Firefox", constants.LongTaskDuration, constants.HighestPriority, constants.MediumIOChance, constants.LargeTaskMemory, "browser"},
		{"Terminal", constants.ShortTaskDuration, constants.MediumPriority, constants.LowIOChance, constants.SmallTaskMemory, "system"},
		{"FileManager", constants.MediumTaskDuration, constants.MediumPriority, constants.HighIOChance, constants.DefaultTaskMemory, "system"},
		{"Calculator", constants.ShortTaskDuration, constants.LowestPriority, constants.NoIOChance, constants.SmallTaskMemory, "utility"},
		{"TextEditor", constants.LongTaskDuration, constants.MediumPriority, constants.LowIOChance, constants.LargeTaskMemory, "development"},
		{"MusicPlayer", constants.LongTaskDuration, constants.HighestPriority, constants.HighIOChance, constants.LargeTaskMemory, "media"},
		{"Backup", constants.LongTaskDuration, constants.LowestPriority, constants.HighIOChance, constants.LargeTaskMemory, "system"},
		{"Game", constants.LongTaskDuration, constants.HighestPriority, constants.LowIOChance, constants.LargeTaskMemory, "entertainment"},
	}

	for i, taskInfo := range tasks {
		t := task.NewTask(
			int64(i+1),
			taskInfo.name,
			taskInfo.duration,
			taskInfo.priority,
			taskInfo.ioChance,
			taskInfo.memory,
			taskInfo.group,
		)
		engine.AddTask(t)
	}

	logger.Info("Demo workload created with %d tasks", len(tasks))
}
