/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: simulate.go
Description: Simulate command for Aurene scheduler including predefined workload
scenarios, algorithm testing, and performance simulation for production deployment.
*/

package cmd

import (
	"aurene/config"
	"aurene/scheduler"
	"aurene/task"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
)

var simulateCmd = &cobra.Command{
	Use:   "simulate [scenario]",
	Short: "Run predefined workload scenarios",
	Long: `Run predefined workload scenarios to test scheduler performance.

Available scenarios:
- burst: High burst of short tasks
- mixed: Mix of long and short tasks
- io-heavy: Tasks with high IO blocking
- memory-heavy: Memory-intensive tasks
- real-time: Real-time task simulation
- custom: Custom scenario from config file

Examples:
  aurene simulate burst
  aurene simulate mixed --duration 30s
  aurene simulate custom --config workload.toml`,
	Args: cobra.ExactArgs(1),
	Run:  runSimulate,
}

var (
	simulateDuration  time.Duration
	simulateAlgorithm string
	simulateConfig    string
	simulateOutput    string
)

func init() {
	simulateCmd.Flags().DurationVarP(&simulateDuration, "duration", "d", 30*time.Second, "Simulation duration")
	simulateCmd.Flags().StringVarP(&simulateAlgorithm, "algorithm", "a", "mlfq", "Scheduling algorithm (mlfq, fcfs, rr, sjf)")
	simulateCmd.Flags().StringVarP(&simulateConfig, "config", "c", "", "Configuration file for custom scenario")
	simulateCmd.Flags().StringVarP(&simulateOutput, "output", "o", "", "Output file for simulation results")

	rootCmd.AddCommand(simulateCmd)
}

/**
 * runSimulate executes the specified workload scenario
 *
 * Runs predefined or custom workload scenarios to test
 * scheduler performance under different conditions.
 */
func runSimulate(cmd *cobra.Command, args []string) {
	scenario := args[0]

	fmt.Printf("🎭 Running simulation: %s\n", scenario)
	fmt.Printf("⏱️  Duration: %v\n", simulateDuration)
	fmt.Printf("🧮 Algorithm: %s\n", simulateAlgorithm)
	fmt.Printf("📁 Config: %s\n\n", simulateConfig)

	// Create scheduler with specified algorithm
	sched := createScheduler(simulateAlgorithm)
	if sched == nil {
		fmt.Printf("❌ Unsupported algorithm: %s\n", simulateAlgorithm)
		os.Exit(1)
	}

	// Load scenario
	workload, err := loadScenario(scenario, simulateConfig)
	if err != nil {
		fmt.Printf("❌ Failed to load scenario: %v\n", err)
		os.Exit(1)
	}

	// Run simulation
	results := runSimulation(sched, workload, simulateDuration)

	// Display results
	displaySimulationResults(results, scenario)

	// Save results if output file specified
	if simulateOutput != "" {
		if err := saveSimulationResults(results, simulateOutput); err != nil {
			fmt.Printf("❌ Failed to save results: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("💾 Results saved to: %s\n", simulateOutput)
	}
}

/**
 * createScheduler creates a scheduler with the specified algorithm
 *
 * Instantiates the appropriate scheduling algorithm
 * based on the provided algorithm name.
 */
func createScheduler(algorithm string) scheduler.SchedulingStrategy {
	factory := scheduler.NewStrategyFactory()

	config := make(map[string]interface{})
	switch algorithm {
	case "rr", "round-robin":
		config["quantum"] = 10 * time.Millisecond
	case "realtime", "rt":
		config["deadline_mode"] = true
	}

	strategy, err := factory.CreateStrategy(algorithm, config)
	if err != nil {
		return nil
	}

	return strategy
}

/**
 * loadScenario loads the specified workload scenario
 *
 * Loads predefined scenarios or custom scenarios from
 * configuration files for simulation.
 */
func loadScenario(scenario, configFile string) (*WorkloadDefinition, error) {
	switch scenario {
	case "burst":
		return &WorkloadDefinition{
			Name:          "burst",
			Description:   "High burst of short tasks",
			TaskCount:     1000,
			ArrivalRate:   100.0, // 100 tasks per second
			DurationRange: [2]int{5, 20},
			MemoryRange:   [2]int64{1024, 1024 * 1024},
			IORate:        0.1,
			PriorityRange: [2]int{0, 2},
			BurstEnabled:  true,
			BurstSize:     50,
			BurstInterval: 5,
		}, nil

	case "mixed":
		return &WorkloadDefinition{
			Name:          "mixed",
			Description:   "Mix of long and short tasks",
			TaskCount:     500,
			ArrivalRate:   20.0,
			DurationRange: [2]int{10, 200},
			MemoryRange:   [2]int64{512 * 1024, 2 * 1024 * 1024},
			IORate:        0.2,
			PriorityRange: [2]int{0, 3},
			BurstEnabled:  false,
		}, nil

	case "io-heavy":
		return &WorkloadDefinition{
			Name:          "io-heavy",
			Description:   "Tasks with high IO blocking",
			TaskCount:     300,
			ArrivalRate:   15.0,
			DurationRange: [2]int{20, 150},
			MemoryRange:   [2]int64{1024 * 1024, 4 * 1024 * 1024},
			IORate:        0.6,
			PriorityRange: [2]int{0, 2},
			BurstEnabled:  false,
		}, nil

	case "memory-heavy":
		return &WorkloadDefinition{
			Name:          "memory-heavy",
			Description:   "Memory-intensive tasks",
			TaskCount:     200,
			ArrivalRate:   10.0,
			DurationRange: [2]int{50, 300},
			MemoryRange:   [2]int64{5 * 1024 * 1024, 20 * 1024 * 1024},
			IORate:        0.05,
			PriorityRange: [2]int{0, 2},
			BurstEnabled:  false,
		}, nil

	case "real-time":
		return &WorkloadDefinition{
			Name:          "real-time",
			Description:   "Real-time task simulation",
			TaskCount:     100,
			ArrivalRate:   5.0,
			DurationRange: [2]int{5, 50},
			MemoryRange:   [2]int64{256 * 1024, 1024 * 1024},
			IORate:        0.02,
			PriorityRange: [2]int{0, 1},
			BurstEnabled:  false,
		}, nil

	case "custom":
		if configFile == "" {
			return nil, fmt.Errorf("custom scenario requires --config file")
		}
		return loadCustomScenario(configFile)

	default:
		return nil, fmt.Errorf("unknown scenario: %s", scenario)
	}
}

/**
 * WorkloadDefinition represents a workload scenario
 *
 * Defines task patterns, arrival rates, and resource
 * requirements for simulation scenarios.
 */
type WorkloadDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	// Task generation
	TaskCount     int     `json:"task_count"`
	ArrivalRate   float64 `json:"arrival_rate"`
	DurationRange [2]int  `json:"duration_range"`

	// Resource patterns
	MemoryRange   [2]int64 `json:"memory_range"`
	IORate        float64  `json:"io_rate"`
	PriorityRange [2]int   `json:"priority_range"`

	// Burst patterns
	BurstEnabled  bool `json:"burst_enabled"`
	BurstSize     int  `json:"burst_size"`
	BurstInterval int  `json:"burst_interval"`
}

/**
 * loadCustomScenario loads a custom scenario from config file
 *
 * Reads workload definition from TOML configuration file
 * for custom simulation scenarios.
 */
func loadCustomScenario(configFile string) (*WorkloadDefinition, error) {
	configManager := config.NewConfigManager(configFile)
	cfg, err := configManager.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Workloads) == 0 {
		return nil, fmt.Errorf("no workloads defined in config file")
	}

	// Use the first workload definition
	workload := cfg.Workloads[0]

	return &WorkloadDefinition{
		Name:          workload.Name,
		Description:   workload.Description,
		TaskCount:     workload.TaskCount,
		ArrivalRate:   workload.ArrivalRate,
		DurationRange: workload.DurationRange,
		MemoryRange:   workload.MemoryRange,
		IORate:        workload.IORate,
		PriorityRange: workload.PriorityRange,
		BurstEnabled:  workload.BurstEnabled,
		BurstSize:     workload.BurstSize,
		BurstInterval: workload.BurstInterval,
	}, nil
}

/**
 * SimulationResult represents simulation execution results
 *
 * Contains performance metrics and statistics from
 * workload simulation execution.
 */
type SimulationResult struct {
	Scenario       string        `json:"scenario"`
	Algorithm      string        `json:"algorithm"`
	Duration       time.Duration `json:"duration"`
	TasksCreated   int64         `json:"tasks_created"`
	TasksCompleted int64         `json:"tasks_completed"`
	TasksBlocked   int64         `json:"tasks_blocked"`
	Throughput     float64       `json:"throughput"`
	AverageLatency time.Duration `json:"average_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	CPUUtilization float64       `json:"cpu_utilization"`
	MemoryUsage    float64       `json:"memory_usage"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
}

/**
 * runSimulation executes the workload simulation
 *
 * Runs the specified workload scenario and collects
 * performance metrics and statistics.
 */
func runSimulation(sched scheduler.SchedulingStrategy, workload *WorkloadDefinition, duration time.Duration) *SimulationResult {
	result := &SimulationResult{
		Scenario:  workload.Name,
		Algorithm: sched.GetName(),
		StartTime: time.Now(),
	}

	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
	}()

	// Generate tasks based on workload definition
	tasksPerSecond := workload.ArrivalRate
	taskInterval := time.Duration(float64(time.Second) / tasksPerSecond)

	ticker := time.NewTicker(taskInterval)
	defer ticker.Stop()

	done := make(chan bool)
	var tasksCreated int64
	var maxLatency time.Duration
	var totalLatency time.Duration
	var latencyCount int64

	// Task generation goroutine
	go func() {
		taskID := int64(0)
		burstCount := 0

		for {
			select {
			case <-ticker.C:
				// Handle burst mode
				if workload.BurstEnabled && burstCount < workload.BurstSize {
					burstCount++
				} else if workload.BurstEnabled {
					time.Sleep(time.Duration(workload.BurstInterval) * time.Second)
					burstCount = 0
					continue
				}

				start := time.Now()

				// Create task with random parameters
				duration := rand.IntN(workload.DurationRange[1]-workload.DurationRange[0]) + workload.DurationRange[0]
				memory := rand.Int64N(workload.MemoryRange[1]-workload.MemoryRange[0]) + workload.MemoryRange[0]
				priority := rand.IntN(workload.PriorityRange[1]-workload.PriorityRange[0]) + workload.PriorityRange[0]

				task := task.NewTask(
					taskID,
					fmt.Sprintf("%s_task_%d", workload.Name, taskID),
					int64(duration),
					priority,
					workload.IORate,
					memory,
					workload.Name,
				)

				sched.AddTask(task)
				taskID++
				atomic.AddInt64(&tasksCreated, 1)

				latency := time.Since(start)
				totalLatency += latency
				latencyCount++

				if latency > maxLatency {
					maxLatency = latency
				}

			case <-done:
				return
			}
		}
	}()

	// Scheduler execution goroutine
	go func() {
		schedTicker := time.NewTicker(time.Millisecond * 4) // 250Hz
		defer schedTicker.Stop()

		startTime := time.Now()
		for time.Since(startTime) < duration {
			select {
			case <-schedTicker.C:
				sched.Tick()
			case <-done:
				return
			}
		}
		done <- true
	}()

	<-done

	// Calculate results
	if latencyCount > 0 {
		result.AverageLatency = totalLatency / time.Duration(latencyCount)
	}
	result.MaxLatency = maxLatency
	result.TasksCreated = atomic.LoadInt64(&tasksCreated)
	result.Throughput = float64(result.TasksCreated) / result.Duration.Seconds()

	// Get scheduler stats
	stats := sched.GetStats()
	if cpuUtil, ok := stats["cpu_utilization"].(float64); ok {
		result.CPUUtilization = cpuUtil
	}
	if memUsage, ok := stats["memory_usage"].(float64); ok {
		result.MemoryUsage = memUsage
	}

	return result
}

/**
 * displaySimulationResults displays simulation execution results
 *
 * Formats and displays comprehensive simulation results
 * including performance metrics and statistics.
 */
func displaySimulationResults(result *SimulationResult, scenario string) {
	fmt.Println("📊 Simulation Results:")
	fmt.Println("==================================================")
	fmt.Printf("Scenario: %s\n", result.Scenario)
	fmt.Printf("Algorithm: %s\n", result.Algorithm)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Tasks Created: %d\n", result.TasksCreated)
	fmt.Printf("Tasks Completed: %d\n", result.TasksCompleted)
	fmt.Printf("Tasks Blocked: %d\n", result.TasksBlocked)
	fmt.Printf("Throughput: %.2f tasks/sec\n", result.Throughput)
	fmt.Printf("Average Latency: %v\n", result.AverageLatency)
	fmt.Printf("Max Latency: %v\n", result.MaxLatency)
	fmt.Printf("CPU Utilization: %.2f%%\n", result.CPUUtilization)
	fmt.Printf("Memory Usage: %.2f%%\n", result.MemoryUsage)
	fmt.Println()
}

/**
 * saveSimulationResults saves simulation results to file
 *
 * Serializes simulation results to JSON format for
 * external analysis and reporting.
 */
func saveSimulationResults(result *SimulationResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write results file: %w", err)
	}

	return nil
}
