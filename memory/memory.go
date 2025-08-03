/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: memory.go
Description: Memory management system for Aurene scheduler including memory limits,
swapping simulation, memory pressure handling, and memory-aware scheduling.
*/

package memory

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/**
 * MemoryManager handles system memory allocation and management
 *
 * Provides memory limits, swapping simulation, and memory pressure
 * detection for production deployment with real memory constraints.
 */
type MemoryManager struct {
	mu sync.RWMutex

	// Memory limits
	totalMemory      int64
	availableMemory  int64
	maxMemoryPerTask int64

	// Memory allocation tracking
	allocatedMemory int64
	taskMemory      map[int64]int64 // taskID -> memory usage

	// Swapping simulation
	swapEnabled   bool
	swapThreshold int
	swapFile      map[int64]int64 // taskID -> swapped memory
	swapUsage     int64

	// Memory pressure
	pressureEnabled   bool
	pressureThreshold int
	pressureLevel     int

	// Statistics
	stats MemoryStats
}

/**
 * MemoryStats tracks memory management performance
 *
 * Provides comprehensive metrics for memory usage,
 * swapping behavior, and pressure events.
 */
type MemoryStats struct {
	TotalAllocations   int64     `json:"total_allocations"`
	TotalDeallocations int64     `json:"total_deallocations"`
	SwapEvents         int64     `json:"swap_events"`
	PressureEvents     int64     `json:"pressure_events"`
	MemoryUtilization  float64   `json:"memory_utilization"`
	SwapUtilization    float64   `json:"swap_utilization"`
	LastUpdated        time.Time `json:"last_updated"`
}

/**
 * MemoryRequest represents a memory allocation request
 *
 * Contains task information and memory requirements
 * for allocation decisions.
 */
type MemoryRequest struct {
	TaskID   int64
	TaskName string
	Size     int64
	Priority int
	Group    string
}

/**
 * MemoryResponse represents the result of a memory operation
 *
 * Contains allocation status, memory address, and
 * any error information.
 */
type MemoryResponse struct {
	Success   bool
	Allocated int64
	Address   int64
	Error     string
	Swapped   bool
	Pressure  bool
}

/**
 * NewMemoryManager creates a new memory manager instance
 *
 * Initializes the memory manager with specified limits
 * and configuration parameters.
 */
func NewMemoryManager(totalMemory, maxMemoryPerTask int64, swapEnabled, pressureEnabled bool) *MemoryManager {
	mm := &MemoryManager{
		totalMemory:       totalMemory,
		availableMemory:   totalMemory,
		maxMemoryPerTask:  maxMemoryPerTask,
		taskMemory:        make(map[int64]int64),
		swapFile:          make(map[int64]int64),
		swapEnabled:       swapEnabled,
		swapThreshold:     80, // 80% memory usage triggers swapping
		pressureEnabled:   pressureEnabled,
		pressureThreshold: 90, // 90% memory usage triggers pressure
		stats: MemoryStats{
			LastUpdated: time.Now(),
		},
	}

	return mm
}

/**
 * AllocateMemory attempts to allocate memory for a task
 *
 * Checks memory limits, handles swapping if enabled,
 * and returns allocation status with detailed information.
 */
func (mm *MemoryManager) AllocateMemory(req *MemoryRequest) *MemoryResponse {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if req.Size <= 0 {
		return &MemoryResponse{
			Success: false,
			Error:   "invalid memory size",
		}
	}

	if req.Size > mm.maxMemoryPerTask {
		return &MemoryResponse{
			Success: false,
			Error:   fmt.Sprintf("memory request exceeds per-task limit: %d > %d", req.Size, mm.maxMemoryPerTask),
		}
	}

	if mm.availableMemory >= req.Size {
		// Direct allocation
		mm.availableMemory -= req.Size
		mm.allocatedMemory += req.Size
		mm.taskMemory[req.TaskID] = req.Size

		atomic.AddInt64(&mm.stats.TotalAllocations, 1)
		mm.updateStats()

		return &MemoryResponse{
			Success:   true,
			Allocated: req.Size,
			Address:   mm.generateAddress(),
		}
	}

	if mm.swapEnabled {
		swapped := mm.trySwapping(req.Size)
		if swapped {
			mm.availableMemory -= req.Size
			mm.allocatedMemory += req.Size
			mm.taskMemory[req.TaskID] = req.Size

			atomic.AddInt64(&mm.stats.TotalAllocations, 1)
			atomic.AddInt64(&mm.stats.SwapEvents, 1)
			mm.updateStats()

			return &MemoryResponse{
				Success:   true,
				Allocated: req.Size,
				Address:   mm.generateAddress(),
				Swapped:   true,
			}
		}
	}

	pressure := mm.checkMemoryPressure()

	return &MemoryResponse{
		Success:  false,
		Error:    "insufficient memory",
		Pressure: pressure,
	}
}

/**
 * DeallocateMemory frees memory allocated to a task
 *
 * Releases memory back to the available pool and
 * updates statistics accordingly.
 */
func (mm *MemoryManager) DeallocateMemory(taskID int64) *MemoryResponse {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	memory, exists := mm.taskMemory[taskID]
	if !exists {
		return &MemoryResponse{
			Success: false,
			Error:   "task not found",
		}
	}

	// Free memory
	mm.availableMemory += memory
	mm.allocatedMemory -= memory
	delete(mm.taskMemory, taskID)

	if mm.swapEnabled && mm.swapUsage > 0 {
		mm.tryUnswapping()
	}

	atomic.AddInt64(&mm.stats.TotalDeallocations, 1)
	mm.updateStats()

	return &MemoryResponse{
		Success:   true,
		Allocated: memory,
	}
}

/**
 * GetMemoryUsage returns current memory usage statistics
 *
 * Provides detailed information about memory allocation,
 * swapping, and pressure levels.
 */
func (mm *MemoryManager) GetMemoryUsage() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	utilization := float64(mm.allocatedMemory) / float64(mm.totalMemory) * 100
	swapUtilization := float64(mm.swapUsage) / float64(mm.totalMemory) * 100

	mm.mu.RUnlock()
	mm.mu.Lock()
	_ = mm.checkMemoryPressure()
	mm.mu.Unlock()
	mm.mu.RLock()

	return map[string]interface{}{
		"total_memory":     mm.totalMemory,
		"available_memory": mm.availableMemory,
		"allocated_memory": mm.allocatedMemory,
		"swap_usage":       mm.swapUsage,
		"utilization":      utilization,
		"swap_utilization": swapUtilization,
		"pressure_level":   mm.pressureLevel,
		"active_tasks":     len(mm.taskMemory),
		"swap_enabled":     mm.swapEnabled,
		"pressure_enabled": mm.pressureEnabled,
	}
}

/**
 * GetStats returns memory management statistics
 *
 * Provides comprehensive metrics for monitoring
 * and performance analysis.
 */
func (mm *MemoryManager) GetStats() *MemoryStats {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	stats := mm.stats
	stats.MemoryUtilization = float64(mm.allocatedMemory) / float64(mm.totalMemory) * 100
	stats.SwapUtilization = float64(mm.swapUsage) / float64(mm.totalMemory) * 100
	stats.LastUpdated = time.Now()

	return &stats
}

/**
 * trySwapping attempts to free memory through swapping
 *
 * Simulates swapping out low-priority tasks to make
 * room for new allocations.
 */
func (mm *MemoryManager) trySwapping(requiredSize int64) bool {
	var tasksToSwap []int64
	var totalSwapped int64

	for taskID, memory := range mm.taskMemory {
		if totalSwapped >= requiredSize {
			break
		}
		tasksToSwap = append(tasksToSwap, taskID)
		totalSwapped += memory
	}

	if totalSwapped < requiredSize {
		return false
	}

	for _, taskID := range tasksToSwap {
		memory := mm.taskMemory[taskID]
		mm.swapFile[taskID] = memory
		mm.swapUsage += memory
		mm.availableMemory += memory
		mm.allocatedMemory -= memory
		delete(mm.taskMemory, taskID)
	}

	return true
}

/**
 * tryUnswapping attempts to bring swapped memory back
 *
 * Simulates bringing swapped memory back into RAM
 * when sufficient memory becomes available.
 */
func (mm *MemoryManager) tryUnswapping() {
	// Simple unswapping: bring back one task at a time
	for taskID, memory := range mm.swapFile {
		if mm.availableMemory >= memory {
			mm.swapFile[taskID] = memory
			mm.swapUsage -= memory
			mm.availableMemory -= memory
			mm.allocatedMemory += memory
			mm.taskMemory[taskID] = memory
			delete(mm.swapFile, taskID)
			break
		}
	}
}

/**
 * checkMemoryPressure determines if system is under memory pressure
 *
 * メモリ圧迫検出システム (◕‿◕)
 *
 * 計算式: pressure_level = (allocated_memory / total_memory) × 100%
 *
 * メモリ使用率が閾値を超えた時に圧迫イベントをトリガーして、
 * プロアクティブなメモリ管理とスワッピング決定を可能にします (◡‿◡)
 */
func (mm *MemoryManager) checkMemoryPressure() bool {
	if !mm.pressureEnabled {
		return false
	}

	// Formula: pressure_level = (allocated / total) × 100%
	utilization := float64(mm.allocatedMemory) / float64(mm.totalMemory) * 100

	if utilization >= float64(mm.pressureThreshold) {
		mm.pressureLevel = int(utilization)
		atomic.AddInt64(&mm.stats.PressureEvents, 1)
		return true
	}

	mm.pressureLevel = 0
	return false
}

/**
 * generateAddress creates a simulated memory address
 *
 * Generates a unique memory address for allocation
 * tracking and debugging purposes.
 */
func (mm *MemoryManager) generateAddress() int64 {
	// Simple address generation for simulation
	return time.Now().UnixNano() % 0x7FFFFFFF
}

/**
 * updateStats updates memory management statistics
 *
 * メモリ統計更新システム (◕‿◕)✨
 *
 * 使用率計算式:
 *   memory_utilization = (allocated_memory / total_memory) × 100%
 *   swap_utilization = (swap_usage / total_memory) × 100%
 *
 * これらのメトリクスはシステム健全性の監視と
 * メモリ管理決定に不可欠です (◡‿◡)
 */
func (mm *MemoryManager) updateStats() {
	// Formula: memory_utilization = (allocated / total) × 100%
	mm.stats.MemoryUtilization = float64(mm.allocatedMemory) / float64(mm.totalMemory) * 100

	// Formula: swap_utilization = (swap_usage / total) × 100%
	mm.stats.SwapUtilization = float64(mm.swapUsage) / float64(mm.totalMemory) * 100
	mm.stats.LastUpdated = time.Now()
}

/**
 * Reset clears all memory allocations and statistics
 *
 * Resets the memory manager to initial state for
 * testing and debugging purposes.
 */
func (mm *MemoryManager) Reset() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.availableMemory = mm.totalMemory
	mm.allocatedMemory = 0
	mm.swapUsage = 0
	mm.pressureLevel = 0

	mm.taskMemory = make(map[int64]int64)
	mm.swapFile = make(map[int64]int64)

	mm.stats = MemoryStats{
		LastUpdated: time.Now(),
	}
}
