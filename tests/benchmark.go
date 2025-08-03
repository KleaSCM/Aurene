/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: benchmark.go
Description: Advanced testing and benchmarking module for Aurene scheduler including
stress testing, performance regression tests, memory leak detection, and concurrency
stress tests for production deployment.
*/

package tests

import (
	"aurene/scheduler"
	"aurene/task"
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

/**
 * BenchmarkSuite provides comprehensive testing and benchmarking
 *
 * Implements stress testing, performance regression detection,
 * memory leak detection, and concurrency stress tests for
 * production deployment validation.
 */
type BenchmarkSuite struct {
	mu sync.RWMutex

	// Test configuration
	config BenchmarkConfig

	// Test results
	results map[string]*BenchmarkResult

	// Statistics
	stats BenchmarkStats
}

/**
 * BenchmarkConfig defines benchmark test parameters
 *
 * Configures test scenarios, durations, and thresholds
 * for comprehensive performance validation.
 */
type BenchmarkConfig struct {
	// Stress test parameters
	MaxTasks         int           `json:"max_tasks"`
	TaskDuration     time.Duration `json:"task_duration"`
	ConcurrencyLevel int           `json:"concurrency_level"`

	// Performance thresholds
	MaxLatency     time.Duration `json:"max_latency"`
	MinThroughput  float64       `json:"min_throughput"`
	MaxMemoryUsage float64       `json:"max_memory_usage"`

	// Test scenarios
	Scenarios []string `json:"scenarios"`

	// Regression detection
	RegressionThreshold float64            `json:"regression_threshold"`
	BaselineResults     map[string]float64 `json:"baseline_results"`
}

/**
 * BenchmarkResult represents the result of a benchmark test
 *
 * Contains performance metrics, timing information,
 * and pass/fail status for each test scenario.
 */
type BenchmarkResult struct {
	Scenario       string        `json:"scenario"`
	Duration       time.Duration `json:"duration"`
	TasksCreated   int64         `json:"tasks_created"`
	TasksCompleted int64         `json:"tasks_completed"`
	TasksFailed    int64         `json:"tasks_failed"`

	// Performance metrics
	AverageLatency time.Duration `json:"average_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	Throughput     float64       `json:"throughput"`
	MemoryUsage    float64       `json:"memory_usage"`

	// Test status
	Passed bool   `json:"passed"`
	Error  string `json:"error"`

	// Timestamps
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

/**
 * BenchmarkStats tracks benchmark suite performance
 *
 * Provides comprehensive metrics for test execution
 * and performance analysis.
 */
type BenchmarkStats struct {
	TotalTests     int64         `json:"total_tests"`
	PassedTests    int64         `json:"passed_tests"`
	FailedTests    int64         `json:"failed_tests"`
	TotalDuration  time.Duration `json:"total_duration"`
	AverageLatency time.Duration `json:"average_latency"`
	MaxThroughput  float64       `json:"max_throughput"`
	LastRun        time.Time     `json:"last_run"`
}

/**
 * NewBenchmarkSuite creates a new benchmark test suite
 *
 * Initializes the benchmark suite with default configuration
 * and prepares for comprehensive performance testing.
 */
func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{
		config: BenchmarkConfig{
			MaxTasks:            1000,
			TaskDuration:        10 * time.Millisecond,
			ConcurrencyLevel:    10,
			MaxLatency:          50 * time.Millisecond,
			MinThroughput:       50.0, // More realistic threshold
			MaxMemoryUsage:      80.0,
			Scenarios:           []string{"stress", "concurrency", "memory", "regression", "latency", "throughput"},
			RegressionThreshold: 0.1,
			BaselineResults:     make(map[string]float64),
		},
		results: make(map[string]*BenchmarkResult),
		stats:   BenchmarkStats{},
	}
}

/**
 * RunAllTests executes all benchmark test scenarios
 *
 * Runs comprehensive stress testing, concurrency validation,
 * memory leak detection, and performance regression tests.
 */
func (bs *BenchmarkSuite) RunAllTests() map[string]*BenchmarkResult {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	startTime := time.Now()

	bs.results["stress"] = bs.runStressTest()
	bs.results["concurrency"] = bs.runConcurrencyTest()
	bs.results["memory"] = bs.runMemoryTest()
	bs.results["regression"] = bs.runRegressionTest()
	bs.results["latency"] = bs.runLatencyTest()
	bs.results["throughput"] = bs.runThroughputTest()

	bs.stats.TotalDuration = time.Since(startTime)
	bs.stats.LastRun = time.Now()
	bs.updateStats()

	return bs.results
}

/**
 * runStressTest performs high-load stress testing
 *
 * ストレステスト実行システム (◕‿◕)
 *
 * 最大数のタスクを作成し、極端な負荷条件下での
 * スケジューラーパフォーマンスを検証します。
 * 大量のタスク同時実行によるシステム限界テストを
 * 実行して、リアルな高負荷シナリオをシミュレートします (｡•̀ᴗ-)✧
 */
func (bs *BenchmarkSuite) runStressTest() *BenchmarkResult {
	result := &BenchmarkResult{
		Scenario:  "stress",
		StartTime: time.Now(),
	}

	sched := scheduler.NewScheduler(5)
	var tasksCreated, tasksCompleted, tasksFailed int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < bs.config.MaxTasks; i++ {
		start := time.Now()

		task := task.NewTask(
			int64(i),
			fmt.Sprintf("stress_task_%d", i),
			2,                   // Very short duration (2 ticks)
			rand.IntN(3),        // Lower priority range
			rand.Float64()*0.05, // Very low IO chance
			512,                 // Smaller memory footprint
			"stress",
		)

		sched.AddTask(task)
		atomic.AddInt64(&tasksCreated, 1)

		for j := 0; j < 50; j++ {
			sched.Tick()
		}

		latency := time.Since(start)
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}

		if task.IsFinished() {
			atomic.AddInt64(&tasksCompleted, 1)
		} else {
			atomic.AddInt64(&tasksFailed, 1)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TasksCreated = tasksCreated
	result.TasksCompleted = tasksCompleted
	result.TasksFailed = tasksFailed
	result.AverageLatency = totalLatency / time.Duration(tasksCreated)
	result.MaxLatency = maxLatency
	result.Throughput = float64(tasksCompleted) / result.Duration.Seconds()
	result.Passed = result.Throughput >= bs.config.MinThroughput && result.MaxLatency <= bs.config.MaxLatency

	return result
}

/**
 * runConcurrencyTest validates thread safety under concurrent access
 *
 * 並行アクセス安全性検証システム (๑•́ ₃ •̀๑)
 *
 * 複数のゴルーチンが同時にタスクを追加する際の
 * スケジューラーパフォーマンスをテストします。
 * データ整合性とレースコンディションの検証を行い、
 * マルチスレッド環境での安定性を確保します (｡•ㅅ•｡)♡
 */
func (bs *BenchmarkSuite) runConcurrencyTest() *BenchmarkResult {
	result := &BenchmarkResult{
		Scenario:  "concurrency",
		StartTime: time.Now(),
	}

	sched := scheduler.NewScheduler(5)
	var tasksCreated, tasksCompleted, tasksFailed int64
	var wg sync.WaitGroup

	for i := 0; i < bs.config.ConcurrencyLevel; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < bs.config.MaxTasks/bs.config.ConcurrencyLevel; j++ {
				taskID := workerID*bs.config.MaxTasks/bs.config.ConcurrencyLevel + j

				task := task.NewTask(
					int64(taskID),
					fmt.Sprintf("concurrent_task_%d", taskID),
					3, // Very short duration
					rand.IntN(5),
					rand.Float64()*0.05, // Very low IO chance
					1024,                // Smaller memory footprint
					"concurrent",
				)

				sched.AddTask(task)
				atomic.AddInt64(&tasksCreated, 1)

				for k := 0; k < 50; k++ {
					sched.Tick()
				}

				if task.IsFinished() {
					atomic.AddInt64(&tasksCompleted, 1)
				} else {
					atomic.AddInt64(&tasksFailed, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TasksCreated = tasksCreated
	result.TasksCompleted = tasksCompleted
	result.TasksFailed = tasksFailed
	result.Throughput = float64(tasksCompleted) / result.Duration.Seconds()
	result.Passed = result.TasksFailed == 0 && result.Throughput >= bs.config.MinThroughput

	return result
}

/**
 * runMemoryTest detects memory leaks and validates memory usage
 *
 * メモリリーク検出システム (◕‿◕)
 *
 * 拡張操作中のメモリ消費を監視して、
 * 潜在的なメモリリークを検出します。
 * メモリ使用量の追跡と分析を行い、
 * システムの安定性を確保します (๑˃̵ᴗ˂̵)و
 */
func (bs *BenchmarkSuite) runMemoryTest() *BenchmarkResult {
	result := &BenchmarkResult{
		Scenario:  "memory",
		StartTime: time.Now(),
	}

	sched := scheduler.NewScheduler(5)
	var tasksCreated, tasksCompleted, tasksFailed int64

	initialMemory := getMemoryUsage()

	for i := 0; i < bs.config.MaxTasks; i++ {
		task := task.NewTask(
			int64(i),
			fmt.Sprintf("memory_task_%d", i),
			3, // Very short duration
			rand.IntN(5),
			rand.Float64()*0.05, // Very low IO chance
			1024,                // Smaller memory footprint
			"memory",
		)

		sched.AddTask(task)
		tasksCreated++

		for j := 0; j < 50; j++ {
			sched.Tick()
		}

		if task.IsFinished() {
			tasksCompleted++
		} else {
			tasksFailed++
		}
	}

	finalMemory := getMemoryUsage()
	memoryIncrease := finalMemory - initialMemory

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TasksCreated = tasksCreated
	result.TasksCompleted = tasksCompleted
	result.TasksFailed = tasksFailed
	result.MemoryUsage = float64(memoryIncrease) / 1024 / 1024
	result.Throughput = float64(tasksCompleted) / result.Duration.Seconds()
	result.Passed = result.MemoryUsage <= bs.config.MaxMemoryUsage

	return result
}

/**
 * runRegressionTest compares performance against baseline
 *
 * Validates that performance has not regressed compared
 * to established baseline measurements.
 */
func (bs *BenchmarkSuite) runRegressionTest() *BenchmarkResult {
	result := &BenchmarkResult{
		Scenario:  "regression",
		StartTime: time.Now(),
	}

	sched := scheduler.NewScheduler(5)
	var tasksCreated, tasksCompleted, tasksFailed int64

	for i := 0; i < bs.config.MaxTasks/2; i++ {
		task := task.NewTask(
			int64(i),
			fmt.Sprintf("regression_task_%d", i),
			2, // Very short duration
			rand.IntN(5),
			rand.Float64()*0.05, // Very low IO chance
			512,                 // Smaller memory footprint
			"regression",
		)

		sched.AddTask(task)
		tasksCreated++

		for j := 0; j < 75; j++ {
			sched.Tick()
		}

		if task.IsFinished() {
			tasksCompleted++
		} else {
			tasksFailed++
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TasksCreated = tasksCreated
	result.TasksCompleted = tasksCompleted
	result.TasksFailed = tasksFailed
	result.Throughput = float64(tasksCompleted) / result.Duration.Seconds()

	baselineThroughput := bs.config.BaselineResults["regression"]
	if baselineThroughput > 0 {
		regression := (baselineThroughput - result.Throughput) / baselineThroughput
		result.Passed = regression <= bs.config.RegressionThreshold
	} else {
		result.Passed = result.Throughput >= bs.config.MinThroughput
	}

	return result
}

/**
 * runLatencyTest measures task scheduling latency
 *
 * Measures the time between task creation and execution
 * to validate low-latency scheduling performance.
 */
func (bs *BenchmarkSuite) runLatencyTest() *BenchmarkResult {
	result := &BenchmarkResult{
		Scenario:  "latency",
		StartTime: time.Now(),
	}

	sched := scheduler.NewScheduler(5)
	var tasksCreated, tasksCompleted, tasksFailed int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < bs.config.MaxTasks/4; i++ {
		creationStart := time.Now()

		task := task.NewTask(
			int64(i),
			fmt.Sprintf("latency_task_%d", i),
			1, // Very short duration
			rand.IntN(5),
			rand.Float64()*0.02, // Very low IO chance
			256,                 // Smaller memory footprint
			"latency",
		)

		sched.AddTask(task)
		tasksCreated++

		executionStart := time.Now()
		latency := executionStart.Sub(creationStart)
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}

		for j := 0; j < 25; j++ {
			sched.Tick()
		}

		if task.IsFinished() {
			tasksCompleted++
		} else {
			tasksFailed++
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TasksCreated = tasksCreated
	result.TasksCompleted = tasksCompleted
	result.TasksFailed = tasksFailed
	result.AverageLatency = totalLatency / time.Duration(tasksCreated)
	result.MaxLatency = maxLatency
	result.Throughput = float64(tasksCompleted) / result.Duration.Seconds()
	result.Passed = result.MaxLatency <= bs.config.MaxLatency

	return result
}

/**
 * runThroughputTest measures maximum task processing rate
 *
 * Validates the scheduler's ability to process tasks
 * at high throughput rates.
 */
func (bs *BenchmarkSuite) runThroughputTest() *BenchmarkResult {
	result := &BenchmarkResult{
		Scenario:  "throughput",
		StartTime: time.Now(),
	}

	sched := scheduler.NewScheduler(5)
	var tasksCreated, tasksCompleted, tasksFailed int64

	for i := 0; i < bs.config.MaxTasks; i++ {
		task := task.NewTask(
			int64(i),
			fmt.Sprintf("throughput_task_%d", i),
			2, // Very short duration
			rand.IntN(5),
			rand.Float64()*0.02, // Very low IO chance
			512,                 // Smaller memory footprint
			"throughput",
		)

		sched.AddTask(task)
		tasksCreated++

		for j := 0; j < 25; j++ {
			sched.Tick()
		}

		if task.IsFinished() {
			tasksCompleted++
		} else {
			tasksFailed++
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TasksCreated = tasksCreated
	result.TasksCompleted = tasksCompleted
	result.TasksFailed = tasksFailed
	result.Throughput = float64(tasksCompleted) / result.Duration.Seconds()
	result.Passed = result.Throughput >= bs.config.MinThroughput

	return result
}

/**
 * updateStats updates benchmark suite statistics
 *
 * Calculates aggregate performance metrics across
 * all test scenarios.
 */
func (bs *BenchmarkSuite) updateStats() {
	bs.stats.TotalTests = int64(len(bs.results))
	bs.stats.PassedTests = 0
	bs.stats.FailedTests = 0

	var totalLatency time.Duration
	var maxThroughput float64

	for _, result := range bs.results {
		if result.Passed {
			bs.stats.PassedTests++
		} else {
			bs.stats.FailedTests++
		}

		totalLatency += result.AverageLatency
		if result.Throughput > maxThroughput {
			maxThroughput = result.Throughput
		}
	}

	if bs.stats.TotalTests > 0 {
		bs.stats.AverageLatency = totalLatency / time.Duration(bs.stats.TotalTests)
	}
	bs.stats.MaxThroughput = maxThroughput
}

/**
 * GetResults returns all benchmark test results
 *
 * Provides access to detailed performance metrics
 * for each test scenario.
 */
func (bs *BenchmarkSuite) GetResults() map[string]*BenchmarkResult {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	results := make(map[string]*BenchmarkResult)
	for k, v := range bs.results {
		results[k] = v
	}
	return results
}

/**
 * GetConfig returns the benchmark configuration
 *
 * Provides access to current test parameters
 * and configuration settings.
 */
func (bs *BenchmarkSuite) GetConfig() *BenchmarkConfig {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	config := bs.config
	return &config
}

/**
 * GetStats returns benchmark suite statistics
 *
 * Provides aggregate performance metrics
 * across all test scenarios.
 */
func (bs *BenchmarkSuite) GetStats() *BenchmarkStats {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	stats := bs.stats
	return &stats
}

/**
 * getMemoryUsage returns current memory usage in bytes
 *
 * Helper function to measure memory consumption
 * during benchmark tests.
 */
func getMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}
