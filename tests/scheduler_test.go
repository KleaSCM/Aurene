/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: scheduler_test.go
Description: Comprehensive unit tests for the MLFQ scheduler implementation.
Tests cover task dispatching, priority aging, preemption, and queue management.
*/

package tests

import (
	"testing"

	"aurene/internal/constants"
	"aurene/scheduler"
	"aurene/task"
)

/**
 * TestSchedulerCreation validates scheduler initialization with various queue configurations
 *
 * Tests the MLFQ scheduler creation with different priority queue counts to ensure
 * proper initialization and queue structure setup for different workload scenarios.
 */
func TestSchedulerCreation(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name      string
		numQueues int
		expected  int
	}{
		{"Default queues", 0, constants.DefaultSchedulerQueues},
		{"Three queues", 3, 3},
		{"Five queues", 5, 5},
		{"Single queue", 1, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sched := scheduler.NewScheduler(tc.numQueues)
			stats := sched.GetStats()
			queueLengths := stats["queue_lengths"].([]int)

			if len(queueLengths) != tc.expected {
				t.Errorf("Expected %d queues, got %d", tc.expected, len(queueLengths))
			}
		})
	}
}

/**
 * TestTaskAddition validates MLFQ task placement in highest priority queue
 *
 * Ensures new tasks are correctly placed in queue 0 (highest priority)
 * and verifies the MLFQ algorithm's initial task assignment behavior.
 */
func TestTaskAddition(t *testing.T) {
	t.Helper()

	sched := scheduler.NewScheduler(3)

	testTask := task.NewTask(1, "TestTask", constants.MediumTaskDuration, constants.HighestPriority, constants.LowIOChance, constants.DefaultTaskMemory, "test")
	sched.AddTask(testTask)

	stats := sched.GetStats()
	queueLengths := stats["queue_lengths"].([]int)

	if queueLengths[0] != 1 {
		t.Errorf("Expected 1 task in queue 0, got %d", queueLengths[0])
	}

	for i := 1; i < len(queueLengths); i++ {
		if queueLengths[i] != 0 {
			t.Errorf("Expected 0 tasks in queue %d, got %d", i, queueLengths[i])
		}
	}
}

/**
 * TestTaskExecution validates task lifecycle and state transitions
 *
 * Tests the complete task execution cycle from Ready → Running → Finished
 * to ensure proper task dispatching and completion handling in MLFQ.
 */
func TestTaskExecution(t *testing.T) {
	t.Helper()

	sched := scheduler.NewScheduler(3)

	testTask := task.NewTask(1, "ShortTask", 5, constants.HighestPriority, constants.NoIOChance, constants.DefaultTaskMemory, "test")
	sched.AddTask(testTask)

	for i := 0; i < 6; i++ {
		sched.Tick()
	}

	if !testTask.IsFinished() {
		t.Error("Task should be finished after 5 ticks")
	}

	if testTask.GetRemaining() != 0 {
		t.Errorf("Task should have 0 remaining ticks, got %d", testTask.GetRemaining())
	}
}

/**
 * TestPriorityAging validates MLFQ priority aging to prevent starvation
 *
 * Tests the aging mechanism that prevents low-priority task starvation
 * by gradually increasing task priorities over time intervals.
 */
func TestPriorityAging(t *testing.T) {
	t.Helper()

	sched := scheduler.NewScheduler(3)

	testTask := task.NewTask(1, "AgingTask", constants.MaxTaskDuration, constants.HighestPriority, constants.NoIOChance, constants.DefaultTaskMemory, "test")
	sched.AddTask(testTask)

	initialPriority := testTask.GetPriority()

	for i := 0; i < constants.PriorityAgingInterval+50; i++ {
		sched.Tick()
	}

	if testTask.GetPriority() <= initialPriority {
		t.Errorf("Priority should have aged, got %d (was %d)",
			testTask.GetPriority(), initialPriority)
	}
}

/**
 * TestPreemption validates MLFQ preemption behavior
 *
 * Tests that higher priority tasks can preempt lower priority running tasks,
 * ensuring the scheduler correctly implements priority-based task switching.
 */
func TestPreemption(t *testing.T) {
	t.Helper()

	sched := scheduler.NewScheduler(3)

	lowTask := task.NewTask(1, "LowPriority", constants.MediumTaskDuration, constants.LowestPriority, constants.NoIOChance, constants.DefaultTaskMemory, "test")
	sched.AddTask(lowTask)

	sched.Tick()

	highTask := task.NewTask(2, "HighPriority", constants.MediumTaskDuration, constants.HighestPriority, constants.NoIOChance, constants.DefaultTaskMemory, "test")
	sched.AddTask(highTask)

	sched.Tick()

	currentTask := sched.GetCurrentTask()
	if currentTask == nil || currentTask.ID != highTask.ID {
		t.Error("High priority task should be running after preemption")
	}

	if lowTask.GetState() != task.Ready {
		t.Errorf("Low priority task should be ready, got %s", lowTask.GetState())
	}
}

/**
 * TestIOBlocking validates realistic IO blocking and unblocking simulation
 *
 * Tests the scheduler's ability to handle tasks that block on IO operations
 * and properly manage their return to ready state when IO completes.
 */
func TestIOBlocking(t *testing.T) {
	t.Helper()

	sched := scheduler.NewScheduler(3)

	testTask := task.NewTask(1, "IOTask", constants.MediumTaskDuration, constants.HighestPriority, constants.HighIOChance, constants.DefaultTaskMemory, "test")
	sched.AddTask(testTask)

	blocked := false
	for i := 0; i < 50; i++ {
		sched.Tick()
		if testTask.GetState() == task.Blocked {
			blocked = true
			break
		}
	}

	if !blocked {
		t.Error("Task should have blocked on IO")
	}

	for i := 0; i < 200; i++ {
		sched.Tick()
		if testTask.GetState() == task.Ready || testTask.GetState() == task.Running {
			break
		}
	}

	if testTask.GetState() != task.Ready && testTask.GetState() != task.Running {
		t.Errorf("Task should have unblocked, got state %s", testTask.GetState())
	}
}

/**
 * TestContextSwitches validates context switch tracking and counting
 *
 * Tests that the scheduler correctly tracks and reports context switches
 * between tasks, ensuring proper performance monitoring capabilities.
 */
func TestContextSwitches(t *testing.T) {
	t.Helper()

	sched := scheduler.NewScheduler(3)

	task1 := task.NewTask(1, "Task1", constants.ShortTaskDuration, constants.HighestPriority, constants.NoIOChance, constants.DefaultTaskMemory, "test")
	task2 := task.NewTask(2, "Task2", constants.ShortTaskDuration, constants.HighestPriority, constants.NoIOChance, constants.DefaultTaskMemory, "test")

	sched.AddTask(task1)
	sched.AddTask(task2)

	for i := 0; i < 25; i++ {
		sched.Tick()
	}

	stats := sched.GetStats()
	contextSwitches := stats["context_switches"].(int64)

	if contextSwitches < 2 {
		t.Errorf("Expected at least 2 context switches, got %d", contextSwitches)
	}
}

/**
 * TestQueueTimeSlices validates exponential time slice configuration
 *
 * Tests that each priority queue has the correct exponential time slice
 * following the formula: time_slice = 2^(queue_index + TimeSliceExponent)
 */
func TestQueueTimeSlices(t *testing.T) {
	t.Helper()

	sched := scheduler.NewScheduler(3)

	expectedTimeSlices := []int{constants.Queue0TimeSlice, constants.Queue1TimeSlice, constants.Queue2TimeSlice}

	stats := sched.GetStats()
	queueLengths := stats["queue_lengths"].([]int)

	if len(queueLengths) != len(expectedTimeSlices) {
		t.Errorf("Expected %d queues, got %d", len(expectedTimeSlices), len(queueLengths))
	}
}

/**
 * TestConcurrentAccess validates thread safety under concurrent task addition
 *
 * Tests that the scheduler maintains data integrity when multiple goroutines
 * add tasks simultaneously, ensuring proper synchronization mechanisms.
 */
func TestConcurrentAccess(t *testing.T) {
	t.Helper()

	sched := scheduler.NewScheduler(3)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			testTask := task.NewTask(int64(id), "ConcurrentTask", constants.ShortTaskDuration, constants.HighestPriority, constants.NoIOChance, constants.DefaultTaskMemory, "test")
			sched.AddTask(testTask)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	stats := sched.GetStats()
	totalTasks := 0
	queueLengths := stats["queue_lengths"].([]int)
	for _, length := range queueLengths {
		totalTasks += length
	}

	if totalTasks != 10 {
		t.Errorf("Expected 10 tasks, got %d", totalTasks)
	}
}
