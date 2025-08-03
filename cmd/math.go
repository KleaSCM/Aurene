/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: math.go
Description: Math workload testing command for Aurene scheduler.
Executes millions of mathematical computation tasks to test
scheduler performance under extreme load conditions.
*/

package cmd

import (
	"aurene/scheduler"
	"aurene/task"
	"aurene/workloads"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var mathCmd = &cobra.Command{
	Use:   "math",
	Short: "Test scheduler with math workload",
	Long: `Test the Aurene scheduler with millions of mathematical computation tasks.
This command generates a massive workload of math tasks with varying complexity,
priority levels, and memory requirements to thoroughly test scheduler performance.`,
	Run: runMathTest,
}

var (
	mathTotalTasks    int64
	mathBatchSize     int
	mathDuration      time.Duration
	mathComplexityMin int
	mathComplexityMax int
	mathPriorityMin   int
	mathPriorityMax   int
	mathMemoryMin     int64
	mathMemoryMax     int64
)

func init() {
	rootCmd.AddCommand(mathCmd)

	mathCmd.Flags().Int64VarP(&mathTotalTasks, "tasks", "t", 1000000, "Total number of math tasks to generate")
	mathCmd.Flags().IntVarP(&mathBatchSize, "batch", "b", 10000, "Batch size for task generation")
	mathCmd.Flags().DurationVarP(&mathDuration, "duration", "d", 5*time.Minute, "Test duration")
	mathCmd.Flags().IntVarP(&mathComplexityMin, "complexity-min", "", 1, "Minimum task complexity")
	mathCmd.Flags().IntVarP(&mathComplexityMax, "complexity-max", "", 100, "Maximum task complexity")
	mathCmd.Flags().IntVarP(&mathPriorityMin, "priority-min", "", 0, "Minimum task priority")
	mathCmd.Flags().IntVarP(&mathPriorityMax, "priority-max", "", 10, "Maximum task priority")
	mathCmd.Flags().Int64VarP(&mathMemoryMin, "memory-min", "", 1024*1024, "Minimum memory per task (bytes)")
	mathCmd.Flags().Int64VarP(&mathMemoryMax, "memory-max", "", 100*1024*1024, "Maximum memory per task (bytes)")
}

func runMathTest(cmd *cobra.Command, args []string) {
	fmt.Printf("🚀 Starting Math Workload Test\n")
	fmt.Printf("📊 Configuration:\n")
	fmt.Printf("   Total Tasks: %d\n", mathTotalTasks)
	fmt.Printf("   Batch Size: %d\n", mathBatchSize)
	fmt.Printf("   Duration: %v\n", mathDuration)
	fmt.Printf("   Strategy: MLFQ (Multi-Level Feedback Queue)\n")
	fmt.Printf("   Complexity Range: %d-%d\n", mathComplexityMin, mathComplexityMax)
	fmt.Printf("   Priority Range: %d-%d\n", mathPriorityMin, mathPriorityMax)
	fmt.Printf("   Memory Range: %d-%d MB\n", mathMemoryMin/1024/1024, mathMemoryMax/1024/1024)

	config := workloads.MathWorkloadConfig{
		TotalTasks:      mathTotalTasks,
		TaskTypes:       []workloads.MathTaskType{workloads.Addition, workloads.Multiplication, workloads.Division, workloads.SquareRoot, workloads.Power, workloads.Factorial, workloads.PrimeCheck, workloads.MatrixMultiply, workloads.Integration, workloads.Differentiation},
		ComplexityRange: [2]int{mathComplexityMin, mathComplexityMax},
		PriorityRange:   [2]int{mathPriorityMin, mathPriorityMax},
		MemoryRange:     [2]int64{mathMemoryMin, mathMemoryMax},
		IOChanceRange:   [2]float64{0.0, 0.3},
		BatchSize:       mathBatchSize,
	}

	generator := workloads.NewMathWorkloadGenerator(config)
	sched := scheduler.NewScheduler(5)

	startTime := time.Now()
	taskChannel := make(chan *task.Task, 1000)

	go generator.GenerateStreamingTasks(taskChannel)

	tasksProcessed := int64(0)
	tasksCompleted := int64(0)
	tasksFailed := int64(0)

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(mathDuration)
		done <- true
	}()

	for {
		select {
		case task := <-taskChannel:
			if task != nil {
				sched.AddTask(task)
				tasksProcessed++
			}
		case <-ticker.C:
			sched.Tick()
			currentTask := sched.GetCurrentTask()
			if currentTask != nil && currentTask.IsFinished() {
				tasksCompleted++
			}
		case <-done:
			goto finish
		}
	}

finish:
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	stats := sched.GetStats()
	generatorStats := generator.GetStats()

	fmt.Printf("\n📈 Test Results:\n")
	fmt.Printf("   Duration: %v\n", duration)
	fmt.Printf("   Tasks Processed: %d\n", tasksProcessed)
	fmt.Printf("   Tasks Completed: %d\n", tasksCompleted)
	fmt.Printf("   Tasks Failed: %d\n", tasksFailed)
	fmt.Printf("   Throughput: %.2f tasks/sec\n", float64(tasksCompleted)/duration.Seconds())
	fmt.Printf("   Context Switches: %d\n", stats["context_switches"])
	fmt.Printf("   Average Latency: %v\n", stats["average_latency"])
	fmt.Printf("   Generator Stats: %+v\n", generatorStats)

	if tasksCompleted > 0 {
		fmt.Printf("✅ Math workload test completed successfully!\n")
	} else {
		fmt.Printf("❌ Math workload test failed!\n")
		os.Exit(1)
	}
}
