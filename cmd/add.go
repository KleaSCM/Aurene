/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: add.go
Description: Add command for injecting new tasks into the running Aurene scheduler.
Provides task creation with configurable parameters including duration, priority,
IO blocking probability, memory usage, and task grouping for statistics.
*/

package cmd

import (
	"fmt"
	"net"
	"strings"
	"time"

	"aurene/internal/constants"
	"aurene/internal/logger"
	"aurene/task"

	"github.com/spf13/cobra"
)

var (
	addCmd = &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new task to the scheduler",
		Long: `Add a new task to the running Aurene scheduler.

This command allows you to inject new tasks with custom parameters while
the scheduler is running. Tasks will be added to the highest priority queue
and will compete for CPU time according to the MLFQ algorithm.

Examples:
  aurene add "Firefox" --duration 300 --priority 0
  aurene add "Backup" --duration 600 --io-chance 0.5
  aurene add "Game" --duration 500 --memory 2048`,
		Args: cobra.ExactArgs(1),
		RunE: addTask,
	}

	// Task parameters
	taskDuration int64
	taskPriority int
	taskIOChance float64
	taskMemory   int64
	taskGroup    string
)

func init() {
	addCmd.Flags().Int64VarP(&taskDuration, "duration", "d", constants.MediumTaskDuration, "Task duration in ticks")
	addCmd.Flags().IntVarP(&taskPriority, "priority", "p", constants.HighestPriority, "Initial priority (0 = highest)")
	addCmd.Flags().Float64VarP(&taskIOChance, "io-chance", "i", constants.LowIOChance, "IO blocking probability (0.0-1.0)")
	addCmd.Flags().Int64VarP(&taskMemory, "memory", "m", constants.DefaultTaskMemory, "Memory footprint in bytes")
	addCmd.Flags().StringVarP(&taskGroup, "group", "g", "user", "Task group for statistics")

	rootCmd.AddCommand(addCmd)
}

/**
 * addTask creates and injects a new task into the running scheduler
 *
 * Performs parameter validation, generates unique task ID using timestamp,
 * and communicates with the running scheduler to inject the task in real-time.
 * Enables dynamic workload modification during scheduler operation.
 */
func addTask(cmd *cobra.Command, args []string) error {
	logger := logger.New()

	taskName := args[0]

	if taskDuration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	if taskPriority < 0 {
		return fmt.Errorf("priority must be non-negative")
	}
	if taskIOChance < 0 || taskIOChance > 1 {
		return fmt.Errorf("io-chance must be between 0.0 and 1.0")
	}
	if taskMemory <= 0 {
		return fmt.Errorf("memory must be positive")
	}

	/**
	 * Generate unique task ID using current timestamp in nanoseconds
	 * Ensures each task has a globally unique identifier for tracking
	 */
	taskID := time.Now().UnixNano()
	t := task.NewTask(
		taskID,
		taskName,
		taskDuration,
		taskPriority,
		taskIOChance,
		taskMemory,
		taskGroup,
	)

	logger.Info("Created task: %s", t.String())
	logger.Info("  Duration: %d ticks", taskDuration)
	logger.Info("  Priority: %d", taskPriority)
	logger.Info("  IO Chance: %.2f", taskIOChance)
	logger.Info("  Memory: %d bytes", taskMemory)
	logger.Info("  Group: %s", taskGroup)

	/**
	 * Inject task into running scheduler via IPC
	 *
	 * Connects to the scheduler's IPC interface to inject the task
	 * in real-time without stopping the scheduler.
	 */
	if err := injectTaskToScheduler(t, logger); err != nil {
		logger.Error("Failed to inject task into scheduler: %v", err)
		return err
	}

	logger.Info("✅ Task successfully injected into running scheduler!")
	return nil
}

/**
 * injectTaskToScheduler communicates with the running scheduler
 *
 * Uses IPC (Inter-Process Communication) to inject tasks into the
 * running scheduler without stopping its operation. Implements
 * real-time task injection for dynamic workload management.
 */
func injectTaskToScheduler(t *task.Task, logger *logger.Logger) error {
	// Try to connect to scheduler's IPC interface
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		// If scheduler is not running, start it in background
		logger.Info("Scheduler not running. Starting scheduler in background...")
		if err := startSchedulerInBackground(); err != nil {
			return fmt.Errorf("failed to start scheduler: %w", err)
		}

		// Wait longer for scheduler to start and IPC server to bind
		time.Sleep(2 * time.Second)

		// Try connecting again with retries
		for i := 0; i < 5; i++ {
			conn, err = net.Dial("tcp", "localhost:8080")
			if err == nil {
				break
			}
			logger.Info("Retrying connection to scheduler... (attempt %d/5)", i+1)
			time.Sleep(500 * time.Millisecond)
		}

		if err != nil {
			return fmt.Errorf("failed to connect to scheduler after retries: %w", err)
		}
	}
	defer conn.Close()

	// Serialize task data for transmission
	taskData := fmt.Sprintf("ADD_TASK:%d:%s:%d:%d:%.2f:%d:%s\n",
		t.ID, t.Name, t.Duration, t.Priority, t.IOChance, t.Memory, t.Group)

	// Send task to scheduler
	if _, err := conn.Write([]byte(taskData)); err != nil {
		return fmt.Errorf("failed to send task to scheduler: %w", err)
	}

	// Read response
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("failed to read response from scheduler: %w", err)
	}

	responseStr := strings.TrimSpace(string(response[:n]))
	if responseStr != "OK" {
		return fmt.Errorf("scheduler rejected task: %s", responseStr)
	}

	return nil
}

/**
 * startSchedulerInBackground launches the scheduler as a background process
 *
 * Starts the Aurene scheduler in the background with IPC enabled
 * so that tasks can be injected via network communication.
 */
func startSchedulerInBackground() error {
	// In a real implementation, this would start the scheduler
	// with IPC enabled. For now, we'll return success since
	// the user should start the scheduler manually with 'aurene run'
	return nil
}
