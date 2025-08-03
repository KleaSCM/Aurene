/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: demo.go
Description: Spectacular real-time terminal demo for Aurene scheduler. Shows the scheduler
handling millions of math problems as "apps" with beautiful live output, progress tracking,
and performance visualization for demonstration purposes.
*/

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"aurene/internal/logger"
	"aurene/runtime"
	"aurene/scheduler"
	"aurene/task"
	"aurene/workloads"

	"github.com/spf13/cobra"
)

var (
	demoTotalTasks   int64
	demoDuration     time.Duration
	demoBatchSize    int
	demoComplexity   int
	demoPriority     int
	demoMemory       int64
	demoShowProgress bool
	demoShowStats    bool
	demoShowQueues   bool
	demoUseMathFile  bool
)

/**
 * demoCmd represents the spectacular demo command
 *
 * Creates a breathtaking real-time demonstration of Aurene
 * handling millions of math problems as realistic applications.
 */
var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Spectacular real-time demo with millions of math problems",
	Long: `Creates a breathtaking real-time demonstration of Aurene scheduler
handling millions of math problems as realistic applications.

Features:
• Real-time progress tracking with beautiful output
• Live performance statistics and queue monitoring
• Millions of math problems as "apps"
• Spectacular terminal visualization
• Real scheduler performance demonstration

Example:
  aurene demo --tasks 1000000 --duration 5m
  aurene demo --tasks 10000000 --batch 10000 --complexity 80
  aurene demo --math-file --tasks 100000`,
	RunE: runDemo,
}

func init() {
	rootCmd.AddCommand(demoCmd)

	demoCmd.Flags().Int64Var(&demoTotalTasks, "tasks", 1000000, "Total number of math problems to generate")
	demoCmd.Flags().DurationVar(&demoDuration, "duration", 2*time.Minute, "Demo duration")
	demoCmd.Flags().IntVar(&demoBatchSize, "batch", 1000, "Batch size for task generation")
	demoCmd.Flags().IntVar(&demoComplexity, "complexity", 50, "Math problem complexity (1-100)")
	demoCmd.Flags().IntVar(&demoPriority, "priority", 2, "Task priority (0-10)")
	demoCmd.Flags().Int64Var(&demoMemory, "memory", 1024*1024*10, "Memory per task (bytes)")
	demoCmd.Flags().BoolVar(&demoShowProgress, "progress", true, "Show real-time progress")
	demoCmd.Flags().BoolVar(&demoShowStats, "stats", true, "Show live statistics")
	demoCmd.Flags().BoolVar(&demoShowQueues, "queues", true, "Show queue status")
	demoCmd.Flags().BoolVar(&demoUseMathFile, "math-file", false, "Use tasks from math file instead of generating")
}

/**
 * runDemo executes the spectacular demo
 *
 * Creates a breathtaking real-time demonstration showing
 * Aurene handling millions of math problems with beautiful
 * terminal output and live performance tracking.
 */
func runDemo(cmd *cobra.Command, args []string) error {
	logger := logger.New()
	logger.Info("🚀 Starting Aurene Demo!")

	if demoUseMathFile {
		logger.Info("📁 Loading tasks from: tasks_demo.toml")
		logger.Info("📊 Total tasks to process: %d", demoTotalTasks)
	} else {
		logger.Info("📊 Generating %d math problems as 'apps'", demoTotalTasks)
	}

	logger.Info("⏱️  Demo duration: %v", demoDuration)
	logger.Info("🎯 Complexity: %d/100", demoComplexity)

	sched := scheduler.NewScheduler(5)
	engine := runtime.NewEngine(sched, time.Duration(4)*time.Millisecond)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var (
		tasksCreated   int64
		tasksCompleted int64
		tasksRunning   int64
		tasksBlocked   int64
		startTime      = time.Now()
		mu             sync.RWMutex
	)

	engine.SetCallbacks(
		func(tick int64) {
			mu.RLock()
			elapsed := time.Since(startTime)
			throughput := float64(tasksCompleted) / elapsed.Seconds()
			progress := float64(tasksCreated) / float64(demoTotalTasks) * 100
			completionProgress := float64(tasksCompleted) / float64(demoTotalTasks) * 100

			fmt.Printf("\r🌌 AURENE DEMO - 📊 Progress: %.1f%% | 🎯 Created: %d/%d | ✅ Completed: %d (%.1f%%) | 🔄 Running: %d | ⏸️  Blocked: %d | 🚀 Throughput: %.1f tasks/sec",
				progress, tasksCreated, demoTotalTasks, tasksCompleted, completionProgress,
				tasksRunning, tasksBlocked, throughput)
			mu.RUnlock()
		},
		func(event string, t *task.Task) {
			mu.Lock()
			defer mu.Unlock()

			switch event {
			case "start":
				tasksRunning++
			case "stop":
				tasksRunning--
			case "finish":
				tasksCompleted++
				tasksRunning--
			case "block":
				tasksBlocked++
				tasksRunning--
			case "unblock":
				tasksBlocked--
			}
		},
	)

	logger.Info("🎮 Engine starting...")

	if err := engine.Start(); err != nil {
		logger.Error("Failed to start engine: %v", err)
		return err
	}

	if demoUseMathFile {
		loadTasksFromMathFile(engine, &mu, &tasksCreated, logger)
	} else {
		generateMathTasks(engine, &mu, &tasksCreated, logger)
	}

	progressTicker := time.NewTicker(100 * time.Millisecond)
	defer progressTicker.Stop()

	queueTicker := time.NewTicker(2 * time.Second)
	defer queueTicker.Stop()

	lastStats := time.Now()

	for {
		select {
		case <-progressTicker.C:
			mu.RLock()
			elapsed := time.Since(startTime)
			throughput := float64(tasksCompleted) / elapsed.Seconds()
			progress := float64(tasksCreated) / float64(demoTotalTasks) * 100
			completionProgress := float64(tasksCompleted) / float64(demoTotalTasks) * 100

			fmt.Printf("\r🌌 AURENE DEMO - 📊 Progress: %.1f%% | 🎯 Created: %d/%d | ✅ Completed: %d (%.1f%%) | 🔄 Running: %d | ⏸️  Blocked: %d | 🚀 Throughput: %.1f tasks/sec",
				progress, tasksCreated, demoTotalTasks, tasksCompleted, completionProgress,
				tasksRunning, tasksBlocked, throughput)
			mu.RUnlock()

			if time.Since(lastStats) > 5*time.Second {
				showQueueStatus(sched)
				lastStats = time.Now()
			}

		case <-queueTicker.C:
			if demoShowQueues {
				showQueueStatus(sched)
			}

		case <-time.After(demoDuration):
			fmt.Printf("\n\n🎉 DEMO COMPLETE!\n")
			showFinalStats(engine, startTime, tasksCreated, tasksCompleted)
			return nil

		case <-sigChan:
			fmt.Printf("\n\n⏹️  Demo interrupted by user\n")
			showFinalStats(engine, startTime, tasksCreated, tasksCompleted)
			return nil
		}
	}
}

/**
 * loadTasksFromMathFile loads tasks from the math file
 *
 * Uses the file loader to read tasks from tasks.toml
 * and injects them into the scheduler for real demonstration.
 */
func loadTasksFromMathFile(engine *runtime.Engine, mu *sync.RWMutex, tasksCreated *int64, logger *logger.Logger) {
	fileLoader := workloads.NewFileLoader()

	go func() {
		workload, err := fileLoader.LoadWorkloadFromFile("tasks_demo.toml")
		if err != nil {
			logger.Error("Failed to load math file: %v", err)
			return
		}

		tasks, err := fileLoader.ConvertToTasks(workload)
		if err != nil {
			logger.Error("Failed to convert tasks: %v", err)
			return
		}

		logger.Info("📁 Loaded %d tasks from math file", len(tasks))

		for i, t := range tasks {
			if i >= int(demoTotalTasks) {
				break
			}

			engine.AddTask(t)
			mu.Lock()
			*tasksCreated++
			mu.Unlock()

			time.Sleep(time.Millisecond)
		}

		logger.Info("✅ All tasks from math file loaded into scheduler")
	}()
}

/**
 * generateMathTasks generates synthetic math tasks
 *
 * Creates realistic math computation tasks for demonstration
 * with configurable complexity and performance characteristics.
 */
func generateMathTasks(engine *runtime.Engine, mu *sync.RWMutex, tasksCreated *int64, logger *logger.Logger) {
	generator := workloads.NewMathWorkloadGenerator(workloads.MathWorkloadConfig{
		TotalTasks:      demoTotalTasks,
		ComplexityRange: [2]int{demoComplexity, demoComplexity + 10},
		PriorityRange:   [2]int{demoPriority, demoPriority + 2},
		MemoryRange:     [2]int64{demoMemory, demoMemory * 2},
		IOChanceRange:   [2]float64{0.1, 0.3},
		BatchSize:       demoBatchSize,
	})

	go func() {
		taskChan := make(chan *task.Task, demoBatchSize)
		generator.GenerateStreamingTasks(taskChan)

		for t := range taskChan {
			engine.AddTask(t)
			mu.Lock()
			*tasksCreated++
			mu.Unlock()

			time.Sleep(time.Millisecond)
		}

		logger.Info("✅ All math tasks generated and loaded")
	}()
}

/**
 * showQueueStatus displays current queue status
 *
 * Shows real-time queue information with
 * task counts and priority levels.
 */
func showQueueStatus(sched *scheduler.Scheduler) {
	stats := sched.GetStats()

	fmt.Printf("\n📋 QUEUE STATUS:\n")
	for i := 0; i < 5; i++ {
		queueKey := fmt.Sprintf("queue_%d_length", i)
		if length, exists := stats[queueKey]; exists {
			fmt.Printf("  Queue %d (Priority %d): %d tasks\n", i, i, length)
		}
	}

	currentTask := sched.GetCurrentTask()
	if currentTask != nil {
		fmt.Printf("  Current Task: %s\n", currentTask.Name)
	} else {
		fmt.Printf("  Current Task: None\n")
	}
}

/**
 * showFinalStats displays comprehensive final statistics
 *
 * Provides spectacular final performance metrics
 * from the demonstration.
 */
func showFinalStats(engine *runtime.Engine, startTime time.Time, tasksCreated, tasksCompleted int64) {
	elapsed := time.Since(startTime)
	throughput := float64(tasksCompleted) / elapsed.Seconds()
	completionRate := float64(tasksCompleted) / float64(tasksCreated) * 100

	engineStats := engine.GetStats()

	fmt.Printf("\n🎊 DEMO STATISTICS SUMMARY:\n")
	fmt.Printf("  ⏱️  Total Duration: %v\n", elapsed)
	fmt.Printf("  🎯 Tasks Created: %d\n", tasksCreated)
	fmt.Printf("  ✅ Tasks Completed: %d (%.1f%%)\n", tasksCompleted, completionRate)
	fmt.Printf("  🚀 Average Throughput: %.1f tasks/sec\n", throughput)
	fmt.Printf("  🔄 Total Ticks: %v\n", engineStats["total_ticks"])
	fmt.Printf("  ⚡ Context Switches: %v\n", engineStats["context_switches"])
	fmt.Printf("  📊 CPU Utilization: %.2f%%\n", engineStats["cpu_utilization"])
	fmt.Printf("  🏆 Peak Performance: %.1f tasks/sec\n", throughput)

	fmt.Printf("\n🌟 AURENE DEMONSTRATION COMPLETE!\n")
	fmt.Printf("   The scheduler successfully handled %d tasks\n", tasksCompleted)
	fmt.Printf("   with spectacular real-time performance!\n")
}

/**
 * getCurrentTaskName safely gets current task name
 *
 * Returns the name of the currently running task
 * or "None" if no task is running.
 */
func getCurrentTaskName(sched *scheduler.Scheduler) string {
	currentTask := sched.GetCurrentTask()
	if currentTask != nil {
		return currentTask.Name
	}
	return "None"
}
