/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: realtime.go
Description: Real-time scheduling test command for Aurene scheduler.
Tests deadline-based scheduling, EDF, and rate monotonic algorithms
with time-critical tasks and deadline miss detection.
*/

package cmd

import (
	"aurene/scheduler"
	"aurene/task"
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var realtimeCmd = &cobra.Command{
	Use:   "realtime",
	Short: "Test real-time scheduling capabilities",
	Long: `Test the Aurene scheduler's real-time capabilities including
deadline-based scheduling, EDF (Earliest Deadline First), and rate monotonic
algorithms with time-critical tasks and deadline miss detection.`,
	Run: runRealtimeTest,
}

var (
	realtimeTotalTasks    int64
	realtimeDuration      time.Duration
	realtimeAlgorithm     string
	realtimeDeadlineRange [2]time.Duration
	realtimePeriodRange   [2]time.Duration
	realtimePriorityRange [2]int
)

func init() {
	rootCmd.AddCommand(realtimeCmd)

	realtimeCmd.Flags().Int64VarP(&realtimeTotalTasks, "tasks", "t", 1000, "Total number of real-time tasks to generate")
	realtimeCmd.Flags().DurationVarP(&realtimeDuration, "duration", "d", 2*time.Minute, "Test duration")
	realtimeCmd.Flags().StringVarP(&realtimeAlgorithm, "algorithm", "a", "edf", "Real-time algorithm (edf, rate-monotonic, priority)")
	realtimeCmd.Flags().DurationVarP(&realtimeDeadlineRange[0], "deadline-min", "", 100*time.Millisecond, "Minimum deadline")
	realtimeCmd.Flags().DurationVarP(&realtimeDeadlineRange[1], "deadline-max", "", 5*time.Second, "Maximum deadline")
	realtimeCmd.Flags().DurationVarP(&realtimePeriodRange[0], "period-min", "", 50*time.Millisecond, "Minimum period")
	realtimeCmd.Flags().DurationVarP(&realtimePeriodRange[1], "period-max", "", 2*time.Second, "Maximum period")
	realtimeCmd.Flags().IntVarP(&realtimePriorityRange[0], "priority-min", "", 0, "Minimum priority")
	realtimeCmd.Flags().IntVarP(&realtimePriorityRange[1], "priority-max", "", 10, "Maximum priority")
}

func runRealtimeTest(cmd *cobra.Command, args []string) {
	fmt.Printf("⚡ Starting Real-Time Scheduling Test\n")
	fmt.Printf("📊 Configuration:\n")
	fmt.Printf("   Total Tasks: %d\n", realtimeTotalTasks)
	fmt.Printf("   Duration: %v\n", realtimeDuration)
	fmt.Printf("   Algorithm: %s\n", realtimeAlgorithm)
	fmt.Printf("   Deadline Range: %v - %v\n", realtimeDeadlineRange[0], realtimeDeadlineRange[1])
	fmt.Printf("   Period Range: %v - %v\n", realtimePeriodRange[0], realtimePeriodRange[1])
	fmt.Printf("   Priority Range: %d-%d\n", realtimePriorityRange[0], realtimePriorityRange[1])

	config := map[string]interface{}{
		"deadline_mode":  true,
		"edf_mode":       realtimeAlgorithm == "edf",
		"rate_monotonic": realtimeAlgorithm == "rate-monotonic",
	}

	rtStrategy := scheduler.NewRealTimeStrategy(true)
	rtStrategy.SetConfig(config)

	startTime := time.Now()
	tasksCreated := int64(0)
	tasksCompleted := int64(0)
	deadlineMisses := 0

	for i := int64(0); i < realtimeTotalTasks; i++ {
		deadline := time.Duration(rand.Int64N(int64(realtimeDeadlineRange[1]-realtimeDeadlineRange[0]))) + realtimeDeadlineRange[0]
		priority := rand.IntN(realtimePriorityRange[1]-realtimePriorityRange[0]) + realtimePriorityRange[0]
		duration := int64(deadline.Milliseconds() / 2)

		task := task.NewTask(
			i,
			fmt.Sprintf("realtime_task_%d", i),
			duration,
			priority,
			rand.Float64()*0.1,
			50*1024*1024,
			"realtime",
		)

		rtStrategy.AddTask(task)
		tasksCreated++

		time.Sleep(time.Microsecond * time.Duration(rand.IntN(100)))
	}

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(realtimeDuration)
		done <- true
	}()

	for {
		select {
		case <-ticker.C:
			rtStrategy.Tick()
			currentTask := rtStrategy.GetNextTask()
			if currentTask != nil && currentTask.IsFinished() {
				tasksCompleted++
			}
		case <-done:
			goto finish
		}
	}

finish:
	endTime := time.Now()
	testDuration := endTime.Sub(startTime)

	stats := rtStrategy.GetStats()
	if misses, exists := stats["deadline_misses"]; exists {
		deadlineMisses = misses.(int)
	}

	fmt.Printf("\n📈 Real-Time Test Results:\n")
	fmt.Printf("   Test Duration: %v\n", testDuration)
	fmt.Printf("   Tasks Created: %d\n", tasksCreated)
	fmt.Printf("   Tasks Completed: %d\n", tasksCompleted)
	fmt.Printf("   Deadline Misses: %d\n", deadlineMisses)
	fmt.Printf("   Success Rate: %.2f%%\n", float64(tasksCompleted)/float64(tasksCreated)*100)
	fmt.Printf("   Miss Rate: %.2f%%\n", float64(deadlineMisses)/float64(tasksCreated)*100)
	fmt.Printf("   Algorithm: %s\n", realtimeAlgorithm)
	fmt.Printf("   EDF Mode: %v\n", stats["edf_mode"])
	fmt.Printf("   Rate Monotonic: %v\n", stats["rate_monotonic"])

	if deadlineMisses == 0 {
		fmt.Printf("✅ Real-time test completed with no deadline misses!\n")
	} else if float64(deadlineMisses)/float64(tasksCreated) < 0.05 {
		fmt.Printf("⚠️  Real-time test completed with acceptable deadline miss rate\n")
	} else {
		fmt.Printf("❌ Real-time test failed with too many deadline misses!\n")
		os.Exit(1)
	}
}
