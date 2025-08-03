/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: config.go
Description: Configuration management system for Aurene scheduler including TOML parser,
workload definitions, system settings, and validation for production deployment.
*/

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

/**
 * SchedulerConfig represents the complete scheduler configuration
 *
 * Defines all system parameters including scheduling algorithms,
 * performance settings, memory limits, and workload definitions
 * for production deployment.
 */
type SchedulerConfig struct {
	// Core scheduler settings
	Scheduler SchedulerSettings `toml:"scheduler"`

	// Memory management
	Memory MemorySettings `toml:"memory"`

	// Performance monitoring
	Performance PerformanceSettings `toml:"performance"`

	// Workload definitions
	Workloads []WorkloadDefinition `toml:"workload"`

	// System integration
	System SystemSettings `toml:"system"`

	// Logging configuration
	Logging LoggingSettings `toml:"logging"`
}

/**
 * SchedulerSettings defines core scheduling parameters
 *
 * Configures the scheduling algorithm, queue settings,
 * time slices, and priority management for production use.
 */
type SchedulerSettings struct {
	// Algorithm selection
	Algorithm string `toml:"algorithm"` // "mlfq", "fcfs", "rr", "sjf"

	// Queue configuration
	NumQueues     int   `toml:"num_queues"`
	TickRate      int   `toml:"tick_rate"`
	TimeSlice     []int `toml:"time_slice"`
	AgingInterval int   `toml:"aging_interval"`

	// Round-Robin specific
	Quantum time.Duration `toml:"quantum"`

	// Real-time scheduling
	RealTimeEnabled bool `toml:"realtime_enabled"`
	DeadlineMode    bool `toml:"deadline_mode"`
}

/**
 * MemorySettings defines memory management parameters
 *
 * Configures memory limits, swapping behavior, and
 * memory-aware scheduling for production systems.
 */
type MemorySettings struct {
	// Memory limits
	MaxMemoryPerTask int64 `toml:"max_memory_per_task"`
	TotalMemoryLimit int64 `toml:"total_memory_limit"`

	// Swapping behavior
	SwappingEnabled bool `toml:"swapping_enabled"`
	SwapThreshold   int  `toml:"swap_threshold"`

	// Memory pressure simulation
	MemoryPressureEnabled bool `toml:"memory_pressure_enabled"`
	PressureThreshold     int  `toml:"pressure_threshold"`
}

/**
 * PerformanceSettings defines monitoring parameters
 *
 * Configures performance tracking, benchmarking,
 * and system health monitoring for production use.
 */
type PerformanceSettings struct {
	// Monitoring intervals
	StatsInterval       time.Duration `toml:"stats_interval"`
	HealthCheckInterval time.Duration `toml:"health_check_interval"`

	// Benchmarking
	BenchmarkEnabled  bool          `toml:"benchmark_enabled"`
	BenchmarkDuration time.Duration `toml:"benchmark_duration"`

	// Resource tracking
	TrackCPUUsage    bool `toml:"track_cpu_usage"`
	TrackMemoryUsage bool `toml:"track_memory_usage"`
	TrackIOUsage     bool `toml:"track_io_usage"`
}

/**
 * WorkloadDefinition defines a workload scenario
 *
 * Describes task patterns, arrival rates, and
 * resource requirements for realistic testing.
 */
type WorkloadDefinition struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`

	// Task generation
	TaskCount     int     `toml:"task_count"`
	ArrivalRate   float64 `toml:"arrival_rate"`
	DurationRange [2]int  `toml:"duration_range"`

	// Resource patterns
	MemoryRange   [2]int64 `toml:"memory_range"`
	IORate        float64  `toml:"io_rate"`
	PriorityRange [2]int   `toml:"priority_range"`

	// Burst patterns
	BurstEnabled  bool `toml:"burst_enabled"`
	BurstSize     int  `toml:"burst_size"`
	BurstInterval int  `toml:"burst_interval"`
}

/**
 * SystemSettings defines system integration parameters
 *
 * Configures integration with real system resources,
 * process monitoring, and kernel-level features.
 */
type SystemSettings struct {
	// System integration
	UseRealCPUStats    bool `toml:"use_real_cpu_stats"`
	UseRealMemoryStats bool `toml:"use_real_memory_stats"`
	UseRealIOStats     bool `toml:"use_real_io_stats"`

	// Process monitoring
	MonitorSystemProcesses bool          `toml:"monitor_system_processes"`
	ProcessPollInterval    time.Duration `toml:"process_poll_interval"`

	// Kernel integration
	KernelIntegrationEnabled bool   `toml:"kernel_integration_enabled"`
	ProcPath                 string `toml:"proc_path"`
	SysPath                  string `toml:"sys_path"`
}

/**
 * LoggingSettings defines logging configuration
 *
 * Configures log levels, output formats, and
 * log rotation for production deployment.
 */
type LoggingSettings struct {
	Level      string `toml:"level"`  // "debug", "info", "warn", "error"
	Format     string `toml:"format"` // "text", "json"
	OutputFile string `toml:"output_file"`

	// Log rotation
	MaxSize    int `toml:"max_size"`
	MaxBackups int `toml:"max_backups"`
	MaxAge     int `toml:"max_age"`
}

/**
 * ConfigManager handles configuration loading and validation
 *
 * Provides functionality to load, validate, and manage
 * scheduler configuration for production deployment.
 */
type ConfigManager struct {
	configPath string
	config     *SchedulerConfig
}

/**
 * NewConfigManager creates a new configuration manager
 *
 * Initializes the configuration manager with default
 * paths and validates the configuration structure.
 */
func NewConfigManager(configPath string) *ConfigManager {
	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		configPath = filepath.Join(homeDir, ".aurene", "config.toml")
	}

	return &ConfigManager{
		configPath: configPath,
	}
}

/**
 * LoadConfig loads and validates the scheduler configuration
 *
 * Reads the TOML configuration file, validates all settings,
 * and provides default values for missing parameters.
 */
func (cm *ConfigManager) LoadConfig() (*SchedulerConfig, error) {
	config := &SchedulerConfig{}

	// Load TOML file if it exists
	if _, err := os.Stat(cm.configPath); err == nil {
		if _, err := toml.DecodeFile(cm.configPath, config); err != nil {
			return nil, fmt.Errorf("failed to decode config file: %w", err)
		}
	}

	// Apply defaults and validate
	if err := cm.applyDefaults(config); err != nil {
		return nil, fmt.Errorf("failed to apply defaults: %w", err)
	}

	if err := cm.validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	cm.config = config
	return config, nil
}

/**
 * applyDefaults sets default values for missing configuration
 *
 * Ensures all configuration parameters have sensible defaults
 * for production deployment.
 */
func (cm *ConfigManager) applyDefaults(config *SchedulerConfig) error {
	// Scheduler defaults
	if config.Scheduler.Algorithm == "" {
		config.Scheduler.Algorithm = "mlfq"
	}
	if config.Scheduler.NumQueues == 0 {
		config.Scheduler.NumQueues = 3
	}
	if config.Scheduler.TickRate == 0 {
		config.Scheduler.TickRate = 250
	}
	if len(config.Scheduler.TimeSlice) == 0 {
		config.Scheduler.TimeSlice = []int{8, 16, 32}
	}
	if config.Scheduler.AgingInterval == 0 {
		config.Scheduler.AgingInterval = 100
	}
	if config.Scheduler.Quantum == 0 {
		config.Scheduler.Quantum = 10 * time.Millisecond
	}

	// Memory defaults
	if config.Memory.MaxMemoryPerTask == 0 {
		config.Memory.MaxMemoryPerTask = 1024 * 1024 * 1024 // 1GB
	}
	if config.Memory.TotalMemoryLimit == 0 {
		config.Memory.TotalMemoryLimit = 8 * 1024 * 1024 * 1024 // 8GB
	}
	if config.Memory.SwapThreshold == 0 {
		config.Memory.SwapThreshold = 80
	}
	if config.Memory.PressureThreshold == 0 {
		config.Memory.PressureThreshold = 90
	}

	// Performance defaults
	if config.Performance.StatsInterval == 0 {
		config.Performance.StatsInterval = 5 * time.Second
	}
	if config.Performance.HealthCheckInterval == 0 {
		config.Performance.HealthCheckInterval = 30 * time.Second
	}
	if config.Performance.BenchmarkDuration == 0 {
		config.Performance.BenchmarkDuration = 60 * time.Second
	}

	// System defaults
	if config.System.ProcPath == "" {
		config.System.ProcPath = "/proc"
	}
	if config.System.SysPath == "" {
		config.System.SysPath = "/sys"
	}
	if config.System.ProcessPollInterval == 0 {
		config.System.ProcessPollInterval = 1 * time.Second
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
	if config.Logging.MaxSize == 0 {
		config.Logging.MaxSize = 100 // MB
	}
	if config.Logging.MaxBackups == 0 {
		config.Logging.MaxBackups = 5
	}
	if config.Logging.MaxAge == 0 {
		config.Logging.MaxAge = 30 // days
	}

	return nil
}

/**
 * validateConfig validates the configuration parameters
 *
 * Ensures all configuration values are within acceptable
 * ranges for production deployment.
 */
func (cm *ConfigManager) validateConfig(config *SchedulerConfig) error {
	// Validate scheduler settings
	if config.Scheduler.NumQueues < 1 || config.Scheduler.NumQueues > 10 {
		return fmt.Errorf("num_queues must be between 1 and 10")
	}
	if config.Scheduler.TickRate < 1 || config.Scheduler.TickRate > 10000 {
		return fmt.Errorf("tick_rate must be between 1 and 10000")
	}
	if len(config.Scheduler.TimeSlice) != config.Scheduler.NumQueues {
		return fmt.Errorf("time_slice length must match num_queues")
	}

	// Validate memory settings
	if config.Memory.MaxMemoryPerTask < 1024*1024 {
		return fmt.Errorf("max_memory_per_task must be at least 1MB")
	}
	if config.Memory.TotalMemoryLimit < config.Memory.MaxMemoryPerTask {
		return fmt.Errorf("total_memory_limit must be greater than max_memory_per_task")
	}

	// Validate performance settings
	if config.Performance.StatsInterval < time.Second {
		return fmt.Errorf("stats_interval must be at least 1 second")
	}
	if config.Performance.HealthCheckInterval < time.Second {
		return fmt.Errorf("health_check_interval must be at least 1 second")
	}

	return nil
}

/**
 * SaveConfig saves the current configuration to file
 *
 * Writes the validated configuration back to the TOML file
 * for persistence across restarts.
 */
func (cm *ConfigManager) SaveConfig(config *SchedulerConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create file
	file, err := os.Create(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Encode TOML
	if err := toml.NewEncoder(file).Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
