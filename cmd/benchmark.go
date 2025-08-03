/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: benchmark.go
Description: Benchmark command for Aurene scheduler including comprehensive performance
testing, stress testing, and regression detection for production deployment.
*/

package cmd

import (
	"aurene/tests"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run comprehensive performance benchmarks",
	Long: `Run comprehensive performance benchmarks for the Aurene scheduler.

This command executes a full suite of performance tests including:
- Stress testing with thousands of tasks
- Concurrency testing for thread safety
- Memory leak detection
- Performance regression testing
- Latency and throughput measurements

The benchmark results are saved to JSON format for analysis.`,
	Run: runBenchmark,
}

var (
	benchmarkOutput      string
	benchmarkDuration    time.Duration
	benchmarkConcurrency int
	benchmarkMaxTasks    int
)

func init() {
	benchmarkCmd.Flags().StringVarP(&benchmarkOutput, "output", "o", "benchmark_results.json", "Output file for benchmark results")
	benchmarkCmd.Flags().DurationVarP(&benchmarkDuration, "duration", "d", 10*time.Second, "Duration for each benchmark test")
	benchmarkCmd.Flags().IntVarP(&benchmarkConcurrency, "concurrency", "c", 100, "Concurrency level for stress tests")
	benchmarkCmd.Flags().IntVarP(&benchmarkMaxTasks, "max-tasks", "t", 10000, "Maximum number of tasks for stress tests")

	rootCmd.AddCommand(benchmarkCmd)
}

/**
 * runBenchmark executes the comprehensive benchmark suite
 *
 * Runs all benchmark tests and saves results to the specified
 * output file for performance analysis.
 */
func runBenchmark(cmd *cobra.Command, args []string) {
	fmt.Println("🚀 Starting Aurene Scheduler Benchmark Suite...")
	fmt.Printf("⏱️  Test Duration: %v\n", benchmarkDuration)
	fmt.Printf("🔄 Concurrency Level: %d\n", benchmarkConcurrency)
	fmt.Printf("📊 Max Tasks: %d\n", benchmarkMaxTasks)
	fmt.Printf("📁 Output File: %s\n\n", benchmarkOutput)

	// Create benchmark suite
	suite := tests.NewBenchmarkSuite()

	// Configure test parameters
	suite.GetConfig().TaskDuration = benchmarkDuration
	suite.GetConfig().ConcurrencyLevel = benchmarkConcurrency
	suite.GetConfig().MaxTasks = benchmarkMaxTasks

	// Run all tests
	fmt.Println("🧪 Running benchmark tests...")
	startTime := time.Now()

	results := suite.RunAllTests()

	totalDuration := time.Since(startTime)

	// Display results
	fmt.Println("\n📈 Benchmark Results:")
	fmt.Println("==================================================")

	passed := 0
	failed := 0

	for scenario, result := range results {
		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
			failed++
		} else {
			passed++
		}

		fmt.Printf("%s %s\n", status, scenario)
		fmt.Printf("   Duration: %v\n", result.Duration)
		fmt.Printf("   Tasks Created: %d\n", result.TasksCreated)
		fmt.Printf("   Tasks Completed: %d\n", result.TasksCompleted)
		fmt.Printf("   Throughput: %.2f tasks/sec\n", result.Throughput)
		fmt.Printf("   Avg Latency: %v\n", result.AverageLatency)
		fmt.Printf("   Max Latency: %v\n", result.MaxLatency)
		if result.MemoryUsage > 0 {
			fmt.Printf("   Memory Usage: %.2f%%\n", result.MemoryUsage)
		}
		if result.Error != "" {
			fmt.Printf("   Error: %s\n", result.Error)
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("📊 Summary:")
	fmt.Println("==================================================")
	fmt.Printf("Total Tests: %d\n", len(results))
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Total Duration: %v\n", totalDuration)

	stats := suite.GetStats()
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
	fmt.Printf("Max Throughput: %.2f tasks/sec\n", stats.MaxThroughput)

	// Save results to file
	if err := saveBenchmarkResults(results, benchmarkOutput); err != nil {
		fmt.Printf("❌ Failed to save results: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n💾 Results saved to: %s\n", benchmarkOutput)

	if failed > 0 {
		fmt.Printf("\n⚠️  %d test(s) failed. Please review the results.\n", failed)
		os.Exit(1)
	} else {
		fmt.Println("\n🎉 All tests passed! Scheduler is performing well.")
	}
}

/**
 * saveBenchmarkResults saves benchmark results to JSON file
 *
 * Serializes benchmark results to JSON format for
 * external analysis and reporting.
 */
func saveBenchmarkResults(results map[string]*tests.BenchmarkResult, filename string) error {
	// Create output structure
	output := struct {
		Timestamp time.Time                         `json:"timestamp"`
		Duration  time.Duration                     `json:"total_duration"`
		Results   map[string]*tests.BenchmarkResult `json:"results"`
		Summary   struct {
			TotalTests int `json:"total_tests"`
			Passed     int `json:"passed"`
			Failed     int `json:"failed"`
		} `json:"summary"`
	}{
		Timestamp: time.Now(),
		Results:   results,
	}

	// Calculate summary
	passed := 0
	failed := 0
	var totalDuration time.Duration

	for _, result := range results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
		totalDuration += result.Duration
	}

	output.Duration = totalDuration
	output.Summary.TotalTests = len(results)
	output.Summary.Passed = passed
	output.Summary.Failed = failed

	// Marshal to JSON
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write results file: %w", err)
	}

	return nil
}
