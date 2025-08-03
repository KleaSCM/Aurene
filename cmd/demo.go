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
  aurene demo --tasks 10000000 --batch 10000 --complexity 80`,
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
	logger.Info("🚀 Starting SPECTACULAR Aurene Demo!")
	logger.Info("📊 Generating %d math problems as 'apps'", demoTotalTasks)
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
			if tick%100 == 0 && demoShowStats {
				mu.RLock()
				elapsed := time.Since(startTime)
				throughput := float64(tasksCompleted) / elapsed.Seconds()
				progress := float64(tasksCreated) / float64(demoTotalTasks) * 100

				fmt.Printf("\r📊 Progress: %.1f%% | Tasks: %d/%d | Completed: %d | Running: %d | Blocked: %d | Throughput: %.1f tasks/sec",
					progress, tasksCreated, demoTotalTasks, tasksCompleted, tasksRunning, tasksBlocked, throughput)
				mu.RUnlock()
			}
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

	config := workloads.MathWorkloadConfig{
		TotalTasks:      demoTotalTasks,
		BatchSize:       demoBatchSize,
		TaskTypes:       []workloads.MathTaskType{workloads.Addition, workloads.Multiplication, workloads.Division, workloads.SquareRoot, workloads.Power, workloads.Factorial, workloads.PrimeCheck, workloads.MatrixMultiply, workloads.Integration, workloads.Differentiation},
		ComplexityRange: [2]int{demoComplexity, demoComplexity + 10},
		PriorityRange:   [2]int{demoPriority, demoPriority + 2},
		MemoryRange:     [2]int64{demoMemory, demoMemory * 2},
		IOChanceRange:   [2]float64{0.1, 0.3},
	}

	generator := workloads.NewMathWorkloadGenerator(config)

	taskChan := make(chan *task.Task, demoBatchSize)

	go func() {
		defer close(taskChan)

		batches := int(demoTotalTasks) / demoBatchSize
		for i := 0; i < batches; i++ {
			tasks := generator.GenerateTasks()

			for _, t := range tasks {
				select {
				case taskChan <- t:
					mu.Lock()
					tasksCreated++
					mu.Unlock()
				case <-time.After(100 * time.Millisecond):
				}
			}

			if demoShowProgress && i%10 == 0 {
				mu.RLock()
				progress := float64(tasksCreated) / float64(demoTotalTasks) * 100
				fmt.Printf("\r🎯 Generated: %.1f%% (%d/%d tasks)", progress, tasksCreated, demoTotalTasks)
				mu.RUnlock()
			}
		}
	}()

	go func() {
		for t := range taskChan {
			engine.AddTask(t)
			time.Sleep(time.Microsecond)
		}
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	lastStats := time.Now()

	for {
		select {
		case <-ticker.C:
			if demoShowStats {
				mu.RLock()
				elapsed := time.Since(startTime)
				throughput := float64(tasksCompleted) / elapsed.Seconds()
				progress := float64(tasksCreated) / float64(demoTotalTasks) * 100
				completionProgress := float64(tasksCompleted) / float64(demoTotalTasks) * 100

				fmt.Printf("\r")
				fmt.Printf("🌌 AURENE DEMO - ")
				fmt.Printf("📊 Progress: %.1f%% | ")
				fmt.Printf("🎯 Created: %d/%d | ")
				fmt.Printf("✅ Completed: %d (%.1f%%) | ")
				fmt.Printf("🔄 Running: %d | ")
				fmt.Printf("⏸️  Blocked: %d | ")
				fmt.Printf("🚀 Throughput: %.1f tasks/sec",
					progress, tasksCreated, demoTotalTasks, tasksCompleted, completionProgress,
					tasksRunning, tasksBlocked, throughput)
				mu.RUnlock()
			}

			if demoShowQueues && time.Since(lastStats) > 2*time.Second {
				showQueueStatus(sched)
				lastStats = time.Now()
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
 * showQueueStatus displays current queue status
 *
 * Shows beautiful real-time queue information with
 * task counts and priority levels for demonstration.
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
	fmt.Printf("  Current Task: %s\n", getCurrentTaskName(sched))
}

/**
 * showFinalStats displays comprehensive final statistics
 *
 * Provides spectacular final performance metrics
 * and demonstration results.
 */
func showFinalStats(engine *runtime.Engine, startTime time.Time, tasksCreated, tasksCompleted int64) {
	elapsed := time.Since(startTime)
	throughput := float64(tasksCompleted) / elapsed.Seconds()
	completionRate := float64(tasksCompleted) / float64(tasksCreated) * 100

	engineStats := engine.GetStats()

	fmt.Printf("\n🎊 FINAL DEMO STATISTICS:\n")
	fmt.Printf("  ⏱️  Total Duration: %v\n", elapsed)
	fmt.Printf("  🎯 Tasks Created: %d\n", tasksCreated)
	fmt.Printf("  ✅ Tasks Completed: %d (%.1f%%)\n", tasksCompleted, completionRate)
	fmt.Printf("  🚀 Average Throughput: %.1f tasks/sec\n", throughput)
	fmt.Printf("  🔄 Total Ticks: %v\n", engineStats["total_ticks"])
	fmt.Printf("  ⚡ Context Switches: %v\n", engineStats["context_switches"])
	fmt.Printf("  📊 CPU Utilization: %.2f%%\n", engineStats["cpu_utilization"])
	fmt.Printf("  🏆 Peak Performance: %.1f tasks/sec\n", throughput)

	fmt.Printf("\n🌟 AURENE SCHEDULER DEMONSTRATION COMPLETE!\n")
	fmt.Printf("   The scheduler successfully handled %d math problems\n", tasksCompleted)
	fmt.Printf("   as realistic applications with real-time scheduling!\n")
}

/**
 * getCurrentTaskName gets the name of the current running task
 *
 * Retrieves the current task name for display purposes.
 */
func getCurrentTaskName(sched *scheduler.Scheduler) string {
	currentTask := sched.GetCurrentTask()
	if currentTask != nil {
		return currentTask.Name
	}
	return "None"
}
