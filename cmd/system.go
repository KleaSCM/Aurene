/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: system.go
Description: Real system monitoring command for Aurene scheduler. Shows actual
CPU usage from the Ryzen 7 processor and real system statistics.
*/

package cmd

import (
	"fmt"
	"time"

	"aurene/internal/logger"
	"aurene/system"

	"github.com/spf13/cobra"
)

var (
	systemDuration time.Duration
	systemInterval time.Duration
)

/**
 * systemCmd represents the real system monitoring command
 *
 * Provides real-time monitoring of actual CPU and memory usage
 * from the system's /proc filesystem for performance comparison.
 */
var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Monitor real system CPU usage",
	Long: `Monitor real system CPU usage and statistics from /proc filesystem.

Shows actual CPU utilization, memory usage, and system performance
for comparison with the simulated scheduler performance.

Example:
  aurene system --duration 30s --interval 1s`,
	RunE: runSystem,
}

func init() {
	rootCmd.AddCommand(systemCmd)

	systemCmd.Flags().DurationVar(&systemDuration, "duration", 30*time.Second, "Monitoring duration")
	systemCmd.Flags().DurationVar(&systemInterval, "interval", 1*time.Second, "Update interval")
}

/**
 * runSystem executes the real system monitoring
 *
 * Monitors actual CPU and memory utilization from /proc/stat
 * and /proc/meminfo to provide real system performance metrics.
 */
func runSystem(cmd *cobra.Command, args []string) error {
	logger := logger.New()
	logger.Info("🖥️  Starting system monitoring")
	logger.Info("⏱️  Duration: %v", systemDuration)
	logger.Info("🔄 Update interval: %v", systemInterval)

	monitor := system.NewSystemMonitor("/proc", "/sys", true)
	if !monitor.IsAvailable() {
		logger.Error("System monitoring not available on this system")
		return fmt.Errorf("system monitoring requires Linux /proc filesystem")
	}

	ticker := time.NewTicker(systemInterval)
	defer ticker.Stop()

	startTime := time.Now()
	var (
		totalCPU    float64
		totalMemory float64
		samples     int
	)

	logger.Info("📊 Starting real system monitoring...")

	for {
		select {
		case <-ticker.C:
			cpuStats, err := monitor.GetCPUStats()
			if err != nil {
				logger.Error("Failed to get CPU stats: %v", err)
				continue
			}

			memoryStats, err := monitor.GetMemoryStats()
			if err != nil {
				logger.Error("Failed to get memory stats: %v", err)
				continue
			}

			monitor.CalculateCPUUtilization(cpuStats)
			monitor.CalculateMemoryUtilization(memoryStats)
			cpuUtil := cpuStats.Utilization
			memoryUtil := memoryStats.MemoryUtilization

			totalCPU += cpuUtil
			totalMemory += memoryUtil
			samples++

			elapsed := time.Since(startTime)
			avgCPU := totalCPU / float64(samples)
			avgMemory := totalMemory / float64(samples)

			fmt.Printf("\r🖥️  SYSTEM - CPU: %.1f%% | Memory: %.1f%% | Uptime: %v | Avg CPU: %.1f%% | Avg Memory: %.1f%%",
				cpuUtil, memoryUtil, elapsed, avgCPU, avgMemory)

		case <-time.After(systemDuration):
			fmt.Printf("\n\n🎊 SYSTEM STATISTICS SUMMARY:\n")
			fmt.Printf("  ⏱️  Total Duration: %v\n", time.Since(startTime))
			fmt.Printf("  📊 Samples Collected: %d\n", samples)
			fmt.Printf("  💻 Average CPU Usage: %.1f%%\n", totalCPU/float64(samples))
			fmt.Printf("  🧠 Average Memory Usage: %.1f%%\n", totalMemory/float64(samples))
			fmt.Printf("  🚀 Peak CPU Usage: %.1f%%\n", totalCPU/float64(samples))

			fmt.Printf("\n🌟 SYSTEM MONITORING COMPLETE!\n")
			fmt.Printf("   System performance monitoring completed successfully.\n")
			return nil
		}
	}
}
