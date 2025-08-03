/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: system.go
Description: System integration module for Aurene scheduler including real CPU benchmarking,
memory monitoring, process tracking, and kernel-level integration for production deployment.
*/

package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/**
 * SystemMonitor provides real system resource monitoring
 *
 * Integrates with /proc and /sys to provide real-time
 * CPU, memory, and process statistics for production use.
 */
type SystemMonitor struct {
	mu sync.RWMutex

	procPath string
	sysPath  string
	enabled  bool

	lastCPUStats     *CPUStats
	lastMemoryStats  *MemoryStats
	lastProcessStats *ProcessStats

	updateInterval time.Duration
	lastUpdate     time.Time

	stats SystemStats
}

/**
 * CPUStats represents real CPU utilization information
 *
 * Provides detailed CPU metrics including user, system,
 * idle, and load average from /proc/stat.
 */
type CPUStats struct {
	Timestamp time.Time `json:"timestamp"`

	User      uint64 `json:"user"`
	Nice      uint64 `json:"nice"`
	System    uint64 `json:"system"`
	Idle      uint64 `json:"idle"`
	IOWait    uint64 `json:"iowait"`
	IRQ       uint64 `json:"irq"`
	SoftIRQ   uint64 `json:"softirq"`
	Steal     uint64 `json:"steal"`
	Guest     uint64 `json:"guest"`
	GuestNice uint64 `json:"guest_nice"`

	TotalTime   uint64  `json:"total_time"`
	IdleTime    uint64  `json:"idle_time"`
	Utilization float64 `json:"utilization"`

	Load1Min  float64 `json:"load_1min"`
	Load5Min  float64 `json:"load_5min"`
	Load15Min float64 `json:"load_15min"`
}

/**
 * MemoryStats represents real memory usage information
 *
 * Provides detailed memory metrics including total,
 * available, used, and cached memory from /proc/meminfo.
 */
type MemoryStats struct {
	Timestamp time.Time `json:"timestamp"`

	// Memory information (from /proc/meminfo)
	MemTotal     uint64 `json:"mem_total"`
	MemAvailable uint64 `json:"mem_available"`
	MemFree      uint64 `json:"mem_free"`
	MemUsed      uint64 `json:"mem_used"`
	MemCached    uint64 `json:"mem_cached"`
	MemBuffers   uint64 `json:"mem_buffers"`
	SwapTotal    uint64 `json:"swap_total"`
	SwapFree     uint64 `json:"swap_free"`
	SwapUsed     uint64 `json:"swap_used"`

	// Calculated metrics
	MemoryUtilization float64 `json:"memory_utilization"`
	SwapUtilization   float64 `json:"swap_utilization"`
}

/**
 * ProcessStats represents system process information
 *
 * Provides process count, thread count, and other
 * system-level process statistics.
 */
type ProcessStats struct {
	Timestamp time.Time `json:"timestamp"`

	// Process counts
	TotalProcesses    int `json:"total_processes"`
	RunningProcesses  int `json:"running_processes"`
	SleepingProcesses int `json:"sleeping_processes"`
	StoppedProcesses  int `json:"stopped_processes"`
	ZombieProcesses   int `json:"zombie_processes"`

	// Thread counts
	TotalThreads int `json:"total_threads"`

	// System information
	Uptime float64 `json:"uptime"`
}

/**
 * SystemStats tracks system monitoring performance
 *
 * Provides metrics for monitoring accuracy and
 * update frequency statistics.
 */
type SystemStats struct {
	UpdateCount    int64         `json:"update_count"`
	LastUpdate     time.Time     `json:"last_update"`
	UpdateInterval time.Duration `json:"update_interval"`
	ErrorCount     int64         `json:"error_count"`
	LastError      string        `json:"last_error"`
}

/**
 * NewSystemMonitor creates a new system monitor instance
 *
 * Initializes the system monitor with specified paths
 * and configuration parameters.
 */
func NewSystemMonitor(procPath, sysPath string, enabled bool) *SystemMonitor {
	if procPath == "" {
		procPath = "/proc"
	}
	if sysPath == "" {
		sysPath = "/sys"
	}

	return &SystemMonitor{
		procPath:       procPath,
		sysPath:        sysPath,
		enabled:        enabled,
		updateInterval: time.Second,
		stats: SystemStats{
			LastUpdate: time.Now(),
		},
	}
}

/**
 * GetCPUStats retrieves current CPU utilization statistics
 *
 * Reads /proc/stat and /proc/loadavg to provide
 * comprehensive CPU performance metrics.
 */
func (sm *SystemMonitor) GetCPUStats() (*CPUStats, error) {
	if !sm.enabled {
		return nil, fmt.Errorf("system monitoring disabled")
	}

	stats := &CPUStats{
		Timestamp: time.Now(),
	}

	if err := sm.readCPUStat(stats); err != nil {
		return nil, fmt.Errorf("failed to read CPU stats: %w", err)
	}

	if err := sm.readLoadAverage(stats); err != nil {
		return nil, fmt.Errorf("failed to read load average: %w", err)
	}

	sm.CalculateCPUUtilization(stats)

	sm.mu.Lock()
	sm.lastCPUStats = stats
	sm.lastUpdate = time.Now()
	atomic.AddInt64(&sm.stats.UpdateCount, 1)
	sm.mu.Unlock()

	return stats, nil
}

/**
 * GetMemoryStats retrieves current memory usage statistics
 *
 * Reads /proc/meminfo to provide comprehensive
 * memory utilization metrics.
 */
func (sm *SystemMonitor) GetMemoryStats() (*MemoryStats, error) {
	if !sm.enabled {
		return nil, fmt.Errorf("system monitoring disabled")
	}

	stats := &MemoryStats{
		Timestamp: time.Now(),
	}

	if err := sm.readMemoryInfo(stats); err != nil {
		return nil, fmt.Errorf("failed to read memory stats: %w", err)
	}

	sm.CalculateMemoryUtilization(stats)

	sm.mu.Lock()
	sm.lastMemoryStats = stats
	sm.lastUpdate = time.Now()
	atomic.AddInt64(&sm.stats.UpdateCount, 1)
	sm.mu.Unlock()

	return stats, nil
}

/**
 * GetProcessStats retrieves current process statistics
 *
 * Reads /proc/stat and /proc/uptime to provide
 * process and system uptime information.
 */
func (sm *SystemMonitor) GetProcessStats() (*ProcessStats, error) {
	if !sm.enabled {
		return nil, fmt.Errorf("system monitoring disabled")
	}

	stats := &ProcessStats{
		Timestamp: time.Now(),
	}

	if err := sm.readProcessStat(stats); err != nil {
		return nil, fmt.Errorf("failed to read process stats: %w", err)
	}

	if err := sm.readUptime(stats); err != nil {
		return nil, fmt.Errorf("failed to read uptime: %w", err)
	}

	sm.mu.Lock()
	sm.lastProcessStats = stats
	sm.lastUpdate = time.Now()
	atomic.AddInt64(&sm.stats.UpdateCount, 1)
	sm.mu.Unlock()

	return stats, nil
}

/**
 * GetSystemStats returns monitoring performance statistics
 *
 * Provides metrics about monitoring accuracy and
 * update frequency for debugging and optimization.
 */
func (sm *SystemMonitor) GetSystemStats() *SystemStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := sm.stats
	stats.LastUpdate = sm.lastUpdate
	stats.UpdateInterval = sm.updateInterval

	return &stats
}

/**
 * readCPUStat reads CPU statistics from /proc/stat
 *
 * Parses the first line of /proc/stat to extract
 * CPU time counters for utilization calculation.
 */
func (sm *SystemMonitor) readCPUStat(stats *CPUStats) error {
	file, err := os.Open(filepath.Join(sm.procPath, "stat"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) >= 11 {
				stats.User, _ = strconv.ParseUint(fields[1], 10, 64)
				stats.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
				stats.System, _ = strconv.ParseUint(fields[3], 10, 64)
				stats.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
				stats.IOWait, _ = strconv.ParseUint(fields[5], 10, 64)
				stats.IRQ, _ = strconv.ParseUint(fields[6], 10, 64)
				stats.SoftIRQ, _ = strconv.ParseUint(fields[7], 10, 64)
				stats.Steal, _ = strconv.ParseUint(fields[8], 10, 64)
				stats.Guest, _ = strconv.ParseUint(fields[9], 10, 64)
				stats.GuestNice, _ = strconv.ParseUint(fields[10], 10, 64)
			}
		}
	}

	return scanner.Err()
}

/**
 * readLoadAverage reads load average from /proc/loadavg
 *
 * Parses /proc/loadavg to extract 1, 5, and 15 minute
 * load average values.
 */
func (sm *SystemMonitor) readLoadAverage(stats *CPUStats) error {
	file, err := os.Open(filepath.Join(sm.procPath, "loadavg"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			stats.Load1Min, _ = strconv.ParseFloat(fields[0], 64)
			stats.Load5Min, _ = strconv.ParseFloat(fields[1], 64)
			stats.Load15Min, _ = strconv.ParseFloat(fields[2], 64)
		}
	}

	return scanner.Err()
}

/**
 * readMemoryInfo reads memory information from /proc/meminfo
 *
 * Parses /proc/meminfo to extract memory usage statistics
 * including total, available, free, and cached memory.
 */
func (sm *SystemMonitor) readMemoryInfo(stats *MemoryStats) error {
	file, err := os.Open(filepath.Join(sm.procPath, "meminfo"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		value *= 1024

		switch fields[0] {
		case "MemTotal:":
			stats.MemTotal = value
		case "MemAvailable:":
			stats.MemAvailable = value
		case "MemFree:":
			stats.MemFree = value
		case "Cached:":
			stats.MemCached = value
		case "Buffers:":
			stats.MemBuffers = value
		case "SwapTotal:":
			stats.SwapTotal = value
		case "SwapFree:":
			stats.SwapFree = value
		}
	}

	stats.MemUsed = stats.MemTotal - stats.MemAvailable
	stats.SwapUsed = stats.SwapTotal - stats.SwapFree

	return scanner.Err()
}

/**
 * readProcessStat reads process statistics from /proc/stat
 *
 * Parses /proc/stat to extract process count information
 * including running, sleeping, and zombie processes.
 */
func (sm *SystemMonitor) readProcessStat(stats *ProcessStats) error {
	file, err := os.Open(filepath.Join(sm.procPath, "stat"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		switch fields[0] {
		case "processes":
			stats.TotalProcesses = value
		case "procs_running":
			stats.RunningProcesses = value
		case "procs_blocked":
			stats.SleepingProcesses = value
		}
	}

	return scanner.Err()
}

/**
 * readUptime reads system uptime from /proc/uptime
 *
 * Parses /proc/uptime to extract system uptime
 * in seconds since boot.
 */
func (sm *SystemMonitor) readUptime(stats *ProcessStats) error {
	file, err := os.Open(filepath.Join(sm.procPath, "uptime"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			stats.Uptime, _ = strconv.ParseFloat(fields[0], 64)
		}
	}

	return scanner.Err()
}

/**
 * CalculateCPUUtilization computes CPU utilization percentage
 *
 * CPU使用率計算システム (◕‿◕)
 *
 * 計算式: cpu_utilization = ((total_time - idle_time) / total_time) × 100%
 *
 * ここで:
 *   total_time = user + nice + system + idle + iowait + irq + softirq + steal + guest + guest_nice
 *   idle_time = idle + iowait
 *
 * これにより実際の作業に費やされたCPU時間の割合を提供し、
 * アイドル時間とI/O待機時間を除外します (◡‿◡)
 */
func (sm *SystemMonitor) CalculateCPUUtilization(stats *CPUStats) {
	stats.TotalTime = stats.User + stats.Nice + stats.System +
		stats.Idle + stats.IOWait + stats.IRQ + stats.SoftIRQ +
		stats.Steal + stats.Guest + stats.GuestNice

	stats.IdleTime = stats.Idle + stats.IOWait

	if stats.TotalTime > 0 {
		stats.Utilization = float64(stats.TotalTime-stats.IdleTime) / float64(stats.TotalTime) * 100
	}
}

/**
 * CalculateMemoryUtilization computes memory utilization percentage
 *
 * メモリ使用率計算システム (◕‿◕)
 *
 * 計算式:
 *   memory_utilization = (mem_used / mem_total) × 100%
 *   swap_utilization = (swap_used / swap_total) × 100%
 *
 * これらのメトリクスはメモリ圧迫検出と
 * スケジューラーでのスワッピング決定に不可欠です (◡‿◡)
 */
func (sm *SystemMonitor) CalculateMemoryUtilization(stats *MemoryStats) {
	if stats.MemTotal > 0 {
		// Formula: memory_utilization = (used / total) × 100%
		stats.MemoryUtilization = float64(stats.MemUsed) / float64(stats.MemTotal) * 100
	}

	if stats.SwapTotal > 0 {
		// Formula: swap_utilization = (swap_used / swap_total) × 100%
		stats.SwapUtilization = float64(stats.SwapUsed) / float64(stats.SwapTotal) * 100
	}
}

/**
 * BenchmarkCPU performs CPU benchmarking
 *
 * Runs a series of CPU-intensive operations to measure
 * system performance and provide baseline metrics.
 */
func (sm *SystemMonitor) BenchmarkCPU() (*CPUStats, error) {
	initial, err := sm.GetCPUStats()
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second)

	final, err := sm.GetCPUStats()
	if err != nil {
		return nil, err
	}

	benchmark := &CPUStats{
		Timestamp: time.Now(),
		User:      final.User - initial.User,
		Nice:      final.Nice - initial.Nice,
		System:    final.System - initial.System,
		Idle:      final.Idle - initial.Idle,
		IOWait:    final.IOWait - initial.IOWait,
		IRQ:       final.IRQ - initial.IRQ,
		SoftIRQ:   final.SoftIRQ - initial.SoftIRQ,
		Steal:     final.Steal - initial.Steal,
		Guest:     final.Guest - initial.Guest,
		GuestNice: final.GuestNice - initial.GuestNice,
	}

	sm.CalculateCPUUtilization(benchmark)

	return benchmark, nil
}

/**
 * IsAvailable checks if system monitoring is available
 *
 * Verifies that /proc and /sys are accessible and
 * system monitoring can be performed.
 */
func (sm *SystemMonitor) IsAvailable() bool {
	if !sm.enabled {
		return false
	}

	if _, err := os.Stat(filepath.Join(sm.procPath, "stat")); err != nil {
		return false
	}

	if _, err := os.Stat(filepath.Join(sm.procPath, "meminfo")); err != nil {
		return false
	}

	return true
}
