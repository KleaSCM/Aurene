/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: memory_test.go
Description: Unit tests for the memory management module including allocation,
deallocation, swapping, and pressure detection tests for production validation.
*/

package tests

import (
	"aurene/memory"
	"testing"
	"time"
)

/**
 * TestMemoryManagerCreation validates memory manager initialization
 *
 * Tests that the memory manager is created with correct initial
 * state and configuration parameters.
 */
func TestMemoryManagerCreation(t *testing.T) {
	t.Helper()

	totalMemory := int64(8 * 1024 * 1024 * 1024)  // 8GB
	maxMemoryPerTask := int64(1024 * 1024 * 1024) // 1GB

	mm := memory.NewMemoryManager(totalMemory, maxMemoryPerTask, true, true)

	if mm == nil {
		t.Fatal("Memory manager should not be nil")
	}

	usage := mm.GetMemoryUsage()
	if usage["total_memory"] != totalMemory {
		t.Errorf("Expected total memory %d, got %v", totalMemory, usage["total_memory"])
	}

	if usage["available_memory"] != totalMemory {
		t.Errorf("Expected available memory %d, got %v", totalMemory, usage["available_memory"])
	}

	if usage["allocated_memory"] != int64(0) {
		t.Errorf("Expected allocated memory 0, got %v", usage["allocated_memory"])
	}
}

/**
 * TestMemoryAllocation validates basic memory allocation
 *
 * Tests that memory can be allocated successfully and
 * the available memory is reduced accordingly.
 */
func TestMemoryAllocation(t *testing.T) {
	t.Helper()

	mm := memory.NewMemoryManager(1024*1024*1024, 512*1024*1024, false, false)

	req := &memory.MemoryRequest{
		TaskID:   1,
		TaskName: "test_task",
		Size:     1024 * 1024, // 1MB
		Priority: 0,
		Group:    "test",
	}

	resp := mm.AllocateMemory(req)

	if !resp.Success {
		t.Errorf("Memory allocation should succeed, got error: %s", resp.Error)
	}

	if resp.Allocated != req.Size {
		t.Errorf("Expected allocated size %d, got %d", req.Size, resp.Allocated)
	}

	usage := mm.GetMemoryUsage()
	expectedAvailable := int64(1024*1024*1024) - req.Size
	if usage["available_memory"] != expectedAvailable {
		t.Errorf("Expected available memory %d, got %v", expectedAvailable, usage["available_memory"])
	}
}

/**
 * TestMemoryAllocationLimit validates memory allocation limits
 *
 * Tests that memory allocation fails when the request exceeds
 * the maximum memory per task limit.
 */
func TestMemoryAllocationLimit(t *testing.T) {
	t.Helper()

	maxMemoryPerTask := int64(1024 * 1024) // 1MB
	mm := memory.NewMemoryManager(1024*1024*1024, maxMemoryPerTask, false, false)

	req := &memory.MemoryRequest{
		TaskID:   1,
		TaskName: "test_task",
		Size:     2 * 1024 * 1024, // 2MB (exceeds limit)
		Priority: 0,
		Group:    "test",
	}

	resp := mm.AllocateMemory(req)

	if resp.Success {
		t.Error("Memory allocation should fail when exceeding limit")
	}

	if resp.Error == "" {
		t.Error("Should return error message when allocation fails")
	}
}

/**
 * TestMemoryDeallocation validates memory deallocation
 *
 * Tests that allocated memory can be freed and the
 * available memory is restored accordingly.
 */
func TestMemoryDeallocation(t *testing.T) {
	t.Helper()

	mm := memory.NewMemoryManager(1024*1024*1024, 512*1024*1024, false, false)

	req := &memory.MemoryRequest{
		TaskID:   1,
		TaskName: "test_task",
		Size:     1024 * 1024, // 1MB
		Priority: 0,
		Group:    "test",
	}

	allocResp := mm.AllocateMemory(req)
	if !allocResp.Success {
		t.Fatal("Memory allocation should succeed")
	}

	deallocResp := mm.DeallocateMemory(1)

	if !deallocResp.Success {
		t.Errorf("Memory deallocation should succeed, got error: %s", deallocResp.Error)
	}

	if deallocResp.Allocated != req.Size {
		t.Errorf("Expected deallocated size %d, got %d", req.Size, deallocResp.Allocated)
	}

	usage := mm.GetMemoryUsage()
	expectedAvailable := int64(1024 * 1024 * 1024)
	if usage["available_memory"] != expectedAvailable {
		t.Errorf("Expected available memory %d, got %v", expectedAvailable, usage["available_memory"])
	}
}

/**
 * TestSwapping validates memory swapping functionality
 *
 * Tests that memory can be swapped out when available
 * memory is insufficient for new allocations.
 */
func TestSwapping(t *testing.T) {
	t.Helper()

	mm := memory.NewMemoryManager(1024*1024*1024, 512*1024*1024, true, false)

	for i := int64(1); i <= 10; i++ {
		req := &memory.MemoryRequest{
			TaskID:   i,
			TaskName: "test_task",
			Size:     100 * 1024 * 1024, // 100MB
			Priority: 0,
			Group:    "test",
		}

		resp := mm.AllocateMemory(req)
		if !resp.Success {
			t.Fatalf("Memory allocation %d should succeed", i)
		}
	}

	req := &memory.MemoryRequest{
		TaskID:   11,
		TaskName: "test_task",
		Size:     200 * 1024 * 1024, // 200MB
		Priority: 0,
		Group:    "test",
	}

	resp := mm.AllocateMemory(req)

	if !resp.Success {
		t.Error("Memory allocation should succeed with swapping")
	}

	if !resp.Swapped {
		t.Error("Allocation should indicate swapping occurred")
	}

	usage := mm.GetMemoryUsage()
	if usage["swap_usage"] == int64(0) {
		t.Error("Swap usage should be greater than 0")
	}
}

/**
 * TestMemoryPressure validates memory pressure detection
 *
 * Tests that memory pressure is detected when utilization
 * exceeds the configured threshold.
 */
func TestMemoryPressure(t *testing.T) {
	t.Helper()

	mm := memory.NewMemoryManager(1024*1024*1024, 1024*1024*1024, false, true)

	req := &memory.MemoryRequest{
		TaskID:   1,
		TaskName: "test_task",
		Size:     950 * 1024 * 1024, // 950MB (95% of 1GB)
		Priority: 0,
		Group:    "test",
	}

	resp := mm.AllocateMemory(req)
	if !resp.Success {
		t.Logf("Memory allocation failed: %s", resp.Error)
		req.Size = 920 * 1024 * 1024 // 920MB
		resp = mm.AllocateMemory(req)
		if !resp.Success {
			t.Fatal("Memory allocation should succeed with smaller size")
		}
	}

	usage := mm.GetMemoryUsage()
	pressureLevel := usage["pressure_level"].(int)

	if pressureLevel == 0 {
		t.Error("Memory pressure should be detected")
	}

	if pressureLevel < 90 {
		t.Errorf("Expected pressure level >= 90, got %d", pressureLevel)
	}
}

/**
 * TestMemoryStats validates memory statistics collection
 *
 * Tests that memory statistics are correctly calculated
 * and updated during memory operations.
 */
func TestMemoryStats(t *testing.T) {
	t.Helper()

	mm := memory.NewMemoryManager(1024*1024*1024, 512*1024*1024, true, true)

	for i := int64(1); i <= 5; i++ {
		req := &memory.MemoryRequest{
			TaskID:   i,
			TaskName: "test_task",
			Size:     100 * 1024 * 1024, // 100MB
			Priority: 0,
			Group:    "test",
		}

		mm.AllocateMemory(req)
	}

	mm.DeallocateMemory(1)
	mm.DeallocateMemory(2)

	stats := mm.GetStats()

	if stats.TotalAllocations != 5 {
		t.Errorf("Expected 5 allocations, got %d", stats.TotalAllocations)
	}

	if stats.TotalDeallocations != 2 {
		t.Errorf("Expected 2 deallocations, got %d", stats.TotalDeallocations)
	}

	if stats.MemoryUtilization <= 0 {
		t.Error("Memory utilization should be greater than 0")
	}

	if stats.LastUpdated.IsZero() {
		t.Error("Last updated time should be set")
	}
}

/**
 * TestMemoryReset validates memory manager reset functionality
 *
 * Tests that the memory manager can be reset to its
 * initial state for testing and debugging.
 */
func TestMemoryReset(t *testing.T) {
	t.Helper()

	mm := memory.NewMemoryManager(1024*1024*1024, 512*1024*1024, true, true)

	req := &memory.MemoryRequest{
		TaskID:   1,
		TaskName: "test_task",
		Size:     100 * 1024 * 1024, // 100MB
		Priority: 0,
		Group:    "test",
	}

	mm.AllocateMemory(req)

	mm.Reset()

	usage := mm.GetMemoryUsage()
	if usage["allocated_memory"] != int64(0) {
		t.Error("Allocated memory should be 0 after reset")
	}

	if usage["swap_usage"] != int64(0) {
		t.Error("Swap usage should be 0 after reset")
	}

	stats := mm.GetStats()
	if stats.TotalAllocations != 0 {
		t.Error("Total allocations should be 0 after reset")
	}

	if stats.TotalDeallocations != 0 {
		t.Error("Total deallocations should be 0 after reset")
	}
}

/**
 * TestConcurrentMemoryOperations validates thread safety
 *
 * Tests that memory operations are thread-safe when
 * performed concurrently by multiple goroutines.
 */
func TestConcurrentMemoryOperations(t *testing.T) {
	t.Helper()

	mm := memory.NewMemoryManager(1024*1024*1024, 100*1024*1024, true, true)

	done := make(chan bool)
	numGoroutines := 10
	allocationsPerGoroutine := 5

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < allocationsPerGoroutine; j++ {
				taskID := int64(id*allocationsPerGoroutine + j + 1)
				req := &memory.MemoryRequest{
					TaskID:   taskID,
					TaskName: "concurrent_task",
					Size:     10 * 1024 * 1024, // 10MB
					Priority: 0,
					Group:    "concurrent",
				}

				resp := mm.AllocateMemory(req)
				if !resp.Success {
					t.Errorf("Concurrent allocation should succeed for task %d", taskID)
				}

				time.Sleep(time.Millisecond)
				deallocResp := mm.DeallocateMemory(taskID)
				if !deallocResp.Success {
					t.Errorf("Concurrent deallocation should succeed for task %d", taskID)
				}
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	usage := mm.GetMemoryUsage()
	if usage["allocated_memory"] != int64(0) {
		t.Error("All memory should be deallocated after concurrent operations")
	}

	stats := mm.GetStats()
	expectedAllocations := int64(numGoroutines * allocationsPerGoroutine)
	if stats.TotalAllocations != expectedAllocations {
		t.Errorf("Expected %d allocations, got %d", expectedAllocations, stats.TotalAllocations)
	}

	if stats.TotalDeallocations != expectedAllocations {
		t.Errorf("Expected %d deallocations, got %d", expectedAllocations, stats.TotalDeallocations)
	}
}
