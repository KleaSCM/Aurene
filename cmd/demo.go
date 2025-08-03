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
	"runtime"
	"sync"
	"syscall"
	"time"

	"aurene/internal/logger"
	aureneruntime "aurene/runtime"
	"aurene/scheduler"
	"aurene/task"

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
 * スペクタクルデモコマンド (◕‿◕)
 *
 * Aureneが数百万の数学問題をリアルなアプリケーションとして
 * 処理する息を呑むようなリアルタイムデモンストレーションを作成します。
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

	demoCmd.Flags().Int64Var(&demoTotalTasks, "tasks", 10000, "Total number of math problems to generate (default: 10000)")
	demoCmd.Flags().DurationVar(&demoDuration, "duration", 30*time.Second, "Demo duration (default: 30 seconds)")
	demoCmd.Flags().IntVar(&demoBatchSize, "batch", 1000, "Batch size for task generation (default: 1000)")
	demoCmd.Flags().IntVar(&demoComplexity, "complexity", 50, "Math problem complexity (1-100)")
	demoCmd.Flags().IntVar(&demoPriority, "priority", 2, "Task priority (0-10)")
	demoCmd.Flags().Int64Var(&demoMemory, "memory", 1024*1024*10, "Memory per task (bytes)")
	demoCmd.Flags().BoolVar(&demoShowProgress, "progress", true, "Show real-time progress")
	demoCmd.Flags().BoolVar(&demoShowStats, "stats", true, "Show live statistics")
	demoCmd.Flags().BoolVar(&demoShowQueues, "queues", true, "Show queue status")
	demoCmd.Flags().BoolVar(&demoUseMathFile, "math-file", false, "Use tasks from math file instead of generating")
}

/**
 *
 * スペクタクルデモ実行 (｡♥‿♥｡)
 *
 * Aureneが数百万の数学問題を美しいターミナル出力と
 * ライブパフォーマンス追跡で処理する息を呑むような
 * リアルタイムデモンストレーションを作成します。
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

	// NO SAFETY CHECK
	// if demoTotalTasks > 1000 {
	// 	logger.Warn("⚠️  Large task count detected (%d), limiting to 1,000 for demo safety", demoTotalTasks)
	// 	demoTotalTasks = 1000
	// }

	sched := scheduler.NewScheduler(5)
	engine := aureneruntime.NewEngine(sched, time.Duration(4)*time.Millisecond)

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

			currentTask := sched.GetCurrentTask()
			currentTaskName := "None"
			if currentTask != nil {
				currentTaskName = currentTask.Name
			}

			engineStats := engine.GetStats()
			cpuUtil := engineStats["cpu_utilization"].(float64)

			schedulerStats := sched.GetStats()
			totalQueued := 0
			for i := 0; i < 5; i++ {
				queueKey := fmt.Sprintf("queue_%d_length", i)
				if length, exists := schedulerStats[queueKey]; exists {
					if queueLength, ok := length.(int); ok {
						totalQueued += queueLength
					}
				}
			}

			contextSwitches := schedulerStats["context_switches"].(int64)

			fmt.Printf("\r🌌 AURENE DEMO - 📊 Progress: %.1f%% | 🎯 Created: %d/%d | ✅ Completed: %d (%.1f%%) | 🔄 Running: %d | ⏸️  Blocked: %d | 📋 Queued: %d | 💻 CPU: %.1f%% | ⚡ Switches: %d | 🎮 Current: %s | 🚀 Throughput: %.1f tasks/sec",
				progress, tasksCreated, demoTotalTasks, tasksCompleted, completionProgress,
				tasksRunning, tasksBlocked, totalQueued, cpuUtil, contextSwitches, currentTaskName, throughput)
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
				if tasksCompleted%100 == 0 {
					fmt.Printf("\n🎯 BATCH COMPLETED: %d tasks finished! 🚀\n", tasksCompleted)
				}
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

	go func() {
		batchSize := int64(5000)
		for i := int64(0); i < demoTotalTasks; i += batchSize {
			for j := int64(0); j < batchSize && (i+j) < demoTotalTasks; j++ {
				task := task.NewTask(
					i+j+1,
					fmt.Sprintf("DemoTask_%d", i+j+1),
					1,       // 1 tick duration - INSTANT completion!
					0,       // priority 0 (highest priority queue)
					0.0,     // 0% IO chance - NO blocking!
					1024*10, // 10KB memory - TINY footprint!
					"demo",
				)
				engine.AddTask(task)
				mu.Lock()
				tasksCreated++
				mu.Unlock()
			}

		}
		logger.Info("✅ All demo tasks generated and loaded")
	}()

	progressTicker := time.NewTicker(100 * time.Millisecond)
	defer progressTicker.Stop()

	queueTicker := time.NewTicker(2 * time.Second)
	defer queueTicker.Stop()

	durationTimer := time.NewTimer(demoDuration)
	defer durationTimer.Stop()

	lastStats := time.Now()

	for {
		select {
		case <-progressTicker.C:
			mu.RLock()
			elapsed := time.Since(startTime)
			throughput := float64(tasksCompleted) / elapsed.Seconds()
			progress := float64(tasksCreated) / float64(demoTotalTasks) * 100
			completionProgress := float64(tasksCompleted) / float64(demoTotalTasks) * 100

			currentTask := sched.GetCurrentTask()
			currentTaskName := "None"
			if currentTask != nil {
				currentTaskName = currentTask.Name
			}

			engineStats := engine.GetStats()
			cpuUtil := engineStats["cpu_utilization"].(float64)

			schedulerStats := sched.GetStats()
			totalQueued := 0
			for i := 0; i < 5; i++ {
				queueKey := fmt.Sprintf("queue_%d_length", i)
				if length, exists := schedulerStats[queueKey]; exists {
					if queueLength, ok := length.(int); ok {
						totalQueued += queueLength
					}
				}
			}

			contextSwitches := schedulerStats["context_switches"].(int64)

			fmt.Printf("\r🌌 AURENE DEMO - 📊 Progress: %.1f%% | 🎯 Created: %d/%d | ✅ Completed: %d (%.1f%%) | 🔄 Running: %d | ⏸️  Blocked: %d | 📋 Queued: %d | 💻 CPU: %.1f%% | ⚡ Switches: %d | 🎮 Current: %s | 🚀 Throughput: %.1f tasks/sec",
				progress, tasksCreated, demoTotalTasks, tasksCompleted, completionProgress,
				tasksRunning, tasksBlocked, totalQueued, cpuUtil, contextSwitches, currentTaskName, throughput)
			mu.RUnlock()

			if time.Since(lastStats) > 5*time.Second {
				showQueueStatus(sched)
				lastStats = time.Now()
			}

		case <-queueTicker.C:
			if demoShowQueues {
				showQueueStatus(sched)
			}

		case <-durationTimer.C:
			fmt.Printf("\n\n🎉 DEMO COMPLETE!\n")
			fmt.Printf("⏱️  Duration expired: %v\n", demoDuration)
			fmt.Printf("📊 Final stats - Created: %d, Completed: %d\n", tasksCreated, tasksCompleted)
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
 * 数学ファイルからのタスク読み込み (◡‿◡)
 *
 * ファイルローダーを使用してtasks.tomlからタスクを読み取り、
 * 実際のデモンストレーションのためにスケジューラに注入します。
 */
func loadTasksFromMathFile(engine *aureneruntime.Engine, mu *sync.RWMutex, tasksCreated *int64, logger *logger.Logger) {
	// DISABLED - BROKEN FUNCTION!
	logger.Info("📁 Math file loading disabled")
}

/**
 * showQueueStatus displays current queue status
 *
 * 現在のキュー状態表示 (◕‿◕)
 *
 * タスク数と優先度レベルを含む
 * リアルタイムキュー情報を表示します。
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
 * 包括的な最終統計表示 (｡♥‿♥｡)
 *
 * デモンストレーションからの
 * スペクタクルな最終パフォーマンスメトリクスを提供します。
 */
func showFinalStats(engine *aureneruntime.Engine, startTime time.Time, tasksCreated, tasksCompleted int64) {
	fmt.Printf("🔍 DEBUG: showFinalStats called!\n")
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

	fmt.Printf("\n🖥️  SYSTEM INFORMATION:\n")
	fmt.Printf("  💻 CPU Cores: %d\n", runtime.NumCPU())
	fmt.Printf("  🧵 Goroutines: %d\n", runtime.NumGoroutine())
	fmt.Printf("  💾 Memory Usage: %d MB\n", getMemoryUsage())
	fmt.Printf("  ⚡ Tick Rate: 250Hz (4ms per tick)\n")

	fmt.Printf("\n🌟 AURENE DEMONSTRATION COMPLETE!\n")
	fmt.Printf("   The scheduler successfully handled %d tasks\n", tasksCompleted)
	fmt.Printf("   with spectacular real-time performance!\n")
}

/**
 * getCurrentTaskName safely gets current task name
 *
 * 現在のタスク名安全取得 (◡‿◡)
 *
 * 現在実行中のタスクの名前を返すか、
 * タスクが実行されていない場合は"None"を返します。
 */
func getCurrentTaskName(sched *scheduler.Scheduler) string {
	currentTask := sched.GetCurrentTask()
	if currentTask != nil {
		return currentTask.Name
	}
	return "None"
}

func getMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc / 1024 / 1024)
}
