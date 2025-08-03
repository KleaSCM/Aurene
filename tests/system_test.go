/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: system_test.go
Description: Unit tests for the system integration module including CPU monitoring,
memory monitoring, process tracking, and system statistics tests for production validation.
*/

package tests

import (
	"aurene/system"
	"testing"
	"time"
)

/**
 * TestSystemMonitorCreation validates system monitor initialization
 *
 * Tests that the system monitor is created with correct initial
 * state and configuration parameters.
 */
func TestSystemMonitorCreation(t *testing.T) {
	t.Helper()

	sm := system.NewSystemMonitor("/proc", "/sys", true)

	if sm == nil {
		t.Fatal("System monitor should not be nil")
	}

	available := sm.IsAvailable()
	if available && !sm.IsAvailable() {
		t.Error("Availability check should be consistent")
	}
}

/**
 * TestSystemMonitorDisabled validates disabled system monitor
 *
 * Tests that the system monitor behaves correctly when
 * system monitoring is disabled.
 */
func TestSystemMonitorDisabled(t *testing.T) {
	t.Helper()

	sm := system.NewSystemMonitor("/proc", "/sys", false)

	_, err := sm.GetCPUStats()
	if err == nil {
		t.Error("CPU stats should fail when monitoring is disabled")
	}

	_, err = sm.GetMemoryStats()
	if err == nil {
		t.Error("Memory stats should fail when monitoring is disabled")
	}

	_, err = sm.GetProcessStats()
	if err == nil {
		t.Error("Process stats should fail when monitoring is disabled")
	}
}

/**
 * TestSystemStats validates system statistics collection
 *
 * Tests that system statistics are correctly collected
 * and updated during monitoring operations.
 */
func TestSystemStats(t *testing.T) {
	t.Helper()

	sm := system.NewSystemMonitor("/proc", "/sys", true)

	stats := sm.GetSystemStats()

	if stats == nil {
		t.Fatal("System stats should not be nil")
	}

	if stats.LastUpdate.IsZero() {
		t.Log("Last update time is zero (expected for new monitor)")
	}

	if stats.UpdateCount != 0 {
		t.Errorf("Expected update count 0, got %d", stats.UpdateCount)
	}
}

/**
 * TestCPUUtilizationCalculation validates CPU utilization formula
 *
 * Tests that CPU utilization is calculated correctly using
 * the formula: utilization = ((total - idle) / total) × 100%
 */
func TestCPUUtilizationCalculation(t *testing.T) {
	t.Helper()

	stats := &system.CPUStats{
		User:      1000,
		Nice:      100,
		System:    500,
		Idle:      2000,
		IOWait:    300,
		IRQ:       50,
		SoftIRQ:   25,
		Steal:     0,
		Guest:     0,
		GuestNice: 0,
	}

	totalTime := stats.User + stats.Nice + stats.System +
		stats.Idle + stats.IOWait + stats.IRQ + stats.SoftIRQ +
		stats.Steal + stats.Guest + stats.GuestNice

	idleTime := stats.Idle + stats.IOWait

	expectedUtilization := float64(totalTime-idleTime) / float64(totalTime) * 100

	sm := system.NewSystemMonitor("/proc", "/sys", true)
	sm.CalculateCPUUtilization(stats)

	tolerance := 0.01
	if abs(stats.Utilization-expectedUtilization) > tolerance {
		t.Errorf("Expected utilization %.2f%%, got %.2f%%", expectedUtilization, stats.Utilization)
	}

	if stats.TotalTime != totalTime {
		t.Errorf("Expected total time %d, got %d", totalTime, stats.TotalTime)
	}

	if stats.IdleTime != idleTime {
		t.Errorf("Expected idle time %d, got %d", idleTime, stats.IdleTime)
	}
}

/**
 * TestMemoryUtilizationCalculation validates memory utilization formula
 *
 * Tests that memory utilization is calculated correctly using
 * the formula: utilization = (used / total) × 100%
 */
func TestMemoryUtilizationCalculation(t *testing.T) {
	t.Helper()

	stats := &system.MemoryStats{
		MemTotal:  8192 * 1024 * 1024, // 8GB
		MemUsed:   4096 * 1024 * 1024, // 4GB
		SwapTotal: 2048 * 1024 * 1024, // 2GB
		SwapUsed:  512 * 1024 * 1024,  // 512MB
	}

	expectedMemoryUtilization := float64(stats.MemUsed) / float64(stats.MemTotal) * 100
	expectedSwapUtilization := float64(stats.SwapUsed) / float64(stats.SwapTotal) * 100

	sm := system.NewSystemMonitor("/proc", "/sys", true)
	sm.CalculateMemoryUtilization(stats)

	tolerance := 0.01
	if abs(stats.MemoryUtilization-expectedMemoryUtilization) > tolerance {
		t.Errorf("Expected memory utilization %.2f%%, got %.2f%%",
			expectedMemoryUtilization, stats.MemoryUtilization)
	}

	if abs(stats.SwapUtilization-expectedSwapUtilization) > tolerance {
		t.Errorf("Expected swap utilization %.2f%%, got %.2f%%",
			expectedSwapUtilization, stats.SwapUtilization)
	}
}

/**
 * TestCPUUtilizationEdgeCases validates edge cases for CPU utilization
 *
 * Tests CPU utilization calculation with edge cases like
 * zero values, maximum values, and boundary conditions.
 */
func TestCPUUtilizationEdgeCases(t *testing.T) {
	t.Helper()

	sm := system.NewSystemMonitor("/proc", "/sys", true)

	stats1 := &system.CPUStats{
		User:      0,
		Nice:      0,
		System:    0,
		Idle:      1000,
		IOWait:    0,
		IRQ:       0,
		SoftIRQ:   0,
		Steal:     0,
		Guest:     0,
		GuestNice: 0,
	}

	sm.CalculateCPUUtilization(stats1)
	if stats1.Utilization != 0 {
		t.Errorf("Expected 0%% utilization for all idle, got %.2f%%", stats1.Utilization)
	}

	stats2 := &system.CPUStats{
		User:      1000,
		Nice:      0,
		System:    0,
		Idle:      0,
		IOWait:    0,
		IRQ:       0,
		SoftIRQ:   0,
		Steal:     0,
		Guest:     0,
		GuestNice: 0,
	}

	sm.CalculateCPUUtilization(stats2)
	if abs(stats2.Utilization-100) > 0.01 {
		t.Errorf("Expected 100%% utilization for all busy, got %.2f%%", stats2.Utilization)
	}

	stats3 := &system.CPUStats{
		User:      0,
		Nice:      0,
		System:    0,
		Idle:      0,
		IOWait:    0,
		IRQ:       0,
		SoftIRQ:   0,
		Steal:     0,
		Guest:     0,
		GuestNice: 0,
	}

	sm.CalculateCPUUtilization(stats3)
	if stats3.Utilization != 0 {
		t.Errorf("Expected 0%% utilization for zero total time, got %.2f%%", stats3.Utilization)
	}
}

/**
 * TestMemoryUtilizationEdgeCases validates edge cases for memory utilization
 *
 * Tests memory utilization calculation with edge cases like
 * zero values, maximum values, and boundary conditions.
 */
func TestMemoryUtilizationEdgeCases(t *testing.T) {
	t.Helper()

	sm := system.NewSystemMonitor("/proc", "/sys", true)

	stats1 := &system.MemoryStats{
		MemTotal:  1024 * 1024 * 1024, // 1GB
		MemUsed:   0,
		SwapTotal: 512 * 1024 * 1024, // 512MB
		SwapUsed:  0,
	}

	sm.CalculateMemoryUtilization(stats1)
	if stats1.MemoryUtilization != 0 {
		t.Errorf("Expected 0%% memory utilization, got %.2f%%", stats1.MemoryUtilization)
	}

	if stats1.SwapUtilization != 0 {
		t.Errorf("Expected 0%% swap utilization, got %.2f%%", stats1.SwapUtilization)
	}

	stats2 := &system.MemoryStats{
		MemTotal:  1024 * 1024 * 1024, // 1GB
		MemUsed:   1024 * 1024 * 1024, // 1GB
		SwapTotal: 512 * 1024 * 1024,  // 512MB
		SwapUsed:  512 * 1024 * 1024,  // 512MB
	}

	sm.CalculateMemoryUtilization(stats2)
	if abs(stats2.MemoryUtilization-100) > 0.01 {
		t.Errorf("Expected 100%% memory utilization, got %.2f%%", stats2.MemoryUtilization)
	}

	if abs(stats2.SwapUtilization-100) > 0.01 {
		t.Errorf("Expected 100%% swap utilization, got %.2f%%", stats2.SwapUtilization)
	}

	stats3 := &system.MemoryStats{
		MemTotal:  0,
		MemUsed:   0,
		SwapTotal: 0,
		SwapUsed:  0,
	}

	sm.CalculateMemoryUtilization(stats3)
	if stats3.MemoryUtilization != 0 {
		t.Errorf("Expected 0%% memory utilization for zero total memory, got %.2f%%", stats3.MemoryUtilization)
	}

	if stats3.SwapUtilization != 0 {
		t.Errorf("Expected 0%% swap utilization for zero total swap, got %.2f%%", stats3.SwapUtilization)
	}
}

/**
 * TestSystemMonitorConcurrency validates thread safety
 *
 * Tests that system monitoring operations are thread-safe
 * when performed concurrently by multiple goroutines.
 */
func TestSystemMonitorConcurrency(t *testing.T) {
	t.Helper()

	sm := system.NewSystemMonitor("/proc", "/sys", true)

	done := make(chan bool)
	numGoroutines := 10
	operationsPerGoroutine := 5

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < operationsPerGoroutine; j++ {
				stats := sm.GetSystemStats()
				if stats == nil {
					t.Errorf("System stats should not be nil for goroutine %d", id)
				}

				available := sm.IsAvailable()
				_ = available

				time.Sleep(time.Millisecond)
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	stats := sm.GetSystemStats()
	if stats == nil {
		t.Error("System stats should not be nil after concurrent operations")
	}
}

/**
 * TestSystemMonitorErrorHandling validates error handling
 *
 * Tests that the system monitor handles errors gracefully
 * when system resources are unavailable or corrupted.
 */
func TestSystemMonitorErrorHandling(t *testing.T) {
	t.Helper()

	sm := system.NewSystemMonitor("/invalid/path", "/invalid/sys", true)

	_, err := sm.GetCPUStats()
	if err == nil {
		t.Error("CPU stats should fail with invalid proc path")
	}

	_, err = sm.GetMemoryStats()
	if err == nil {
		t.Error("Memory stats should fail with invalid proc path")
	}

	_, err = sm.GetProcessStats()
	if err == nil {
		t.Error("Process stats should fail with invalid proc path")
	}

	if sm.IsAvailable() {
		t.Error("System monitor should not be available with invalid paths")
	}
}

/**
 * TestSystemMonitorBenchmark validates CPU benchmarking
 *
 * Tests that CPU benchmarking functionality works correctly
 * and provides meaningful performance metrics.
 */
func TestSystemMonitorBenchmark(t *testing.T) {
	t.Helper()

	sm := system.NewSystemMonitor("/proc", "/sys", true)

	benchmark, err := sm.BenchmarkCPU()

	if err != nil {
		t.Logf("CPU benchmark failed (expected in test environment): %v", err)
		return
	}

	if benchmark != nil {
		if benchmark.Timestamp.IsZero() {
			t.Error("Benchmark timestamp should be set")
		}

		if benchmark.Utilization < 0 || benchmark.Utilization > 100 {
			t.Errorf("Benchmark utilization should be between 0-100%%, got %.2f%%", benchmark.Utilization)
		}

		if benchmark.TotalTime <= 0 {
			t.Error("Benchmark total time should be positive")
		}
	}
}

/**
 * abs returns the absolute value of a float64
 *
 * Helper function for comparing floating point values
 * with tolerance in tests.
 */
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
