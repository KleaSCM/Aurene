/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: task.go
Description: Core task structure and lifecycle management for the Aurene scheduler.
Implements process simulation with CPU bursts, IO blocking, and state transitions.
Each task represents a process with priority, memory usage, and execution time.
*/

package task

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

// TaskState represents the current state of a task
type TaskState int

const (
	Ready TaskState = iota
	Running
	Blocked
	Finished
)

// String returns the string representation of TaskState
func (s TaskState) String() string {
	switch s {
	case Ready:
		return "Ready"
	case Running:
		return "Running"
	case Blocked:
		return "Blocked"
	case Finished:
		return "Finished"
	default:
		return "Unknown"
	}
}

// Task represents a real process in the scheduler
type Task struct {
	ID          int64
	Name        string
	Duration    int64      // Total CPU time needed (in ticks)
	Remaining   int64      // Remaining CPU time (atomic)
	State       TaskState  // Current state
	Priority    int        // MLFQ priority level (0 = highest)
	IOChance    float64    // Probability of IO blocking per tick
	Memory      int64      // Memory footprint in bytes
	Group       string     // Task grouping for stats
	ArrivalTime time.Time  // When task was created
	StartTime   *time.Time // When task first started running
	EndTime     *time.Time // When task finished

	// Statistics
	WaitTime       int64 // Total time spent waiting
	TurnaroundTime int64 // Total time from arrival to completion

	// Internal state
	mu          sync.RWMutex
	blockedAt   *time.Time
	lastRunTime *time.Time
}

// NewTask creates a new task with the given parameters
func NewTask(id int64, name string, duration int64, priority int, ioChance float64, memory int64, group string) *Task {
	now := time.Now()
	return &Task{
		ID:          id,
		Name:        name,
		Duration:    duration,
		Remaining:   duration,
		State:       Ready,
		Priority:    priority,
		IOChance:    ioChance,
		Memory:      memory,
		Group:       group,
		ArrivalTime: now,
	}
}

// Execute runs the task for one tick
func (t *Task) Execute() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State != Running {
		return false
	}

	// Decrement remaining time
	remaining := atomic.AddInt64(&t.Remaining, -1)

	// Update last run time
	now := time.Now()
	t.lastRunTime = &now

	// Check if task is finished
	if remaining <= 0 {
		t.State = Finished
		t.EndTime = &now
		t.TurnaroundTime = int64(now.Sub(t.ArrivalTime).Milliseconds())
		return true
	}

	// Simulate IO blocking
	if t.shouldBlock() {
		t.State = Blocked
		t.blockedAt = &now
		return false
	}

	return false
}

/**
 * shouldBlock determines if the task should block on IO
 *
 * Uses proper random number generation for realistic IO blocking simulation.
 * Each task has a probability-based chance to block on IO operations
 * based on its configured IO chance parameter.
 */
func (t *Task) shouldBlock() bool {
	if t.IOChance <= 0 {
		return false
	}
	// Use proper random number generation for IO blocking
	// Formula: block_if(random_float < io_chance)
	return rand.Float64() < t.IOChance
}

// Unblock moves the task from Blocked to Ready state
func (t *Task) Unblock() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State == Blocked {
		t.State = Ready
		t.blockedAt = nil
	}
}

// Start begins execution of the task
func (t *Task) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State == Ready {
		t.State = Running
		if t.StartTime == nil {
			now := time.Now()
			t.StartTime = &now
		}
	}
}

// Stop pauses execution of the task
func (t *Task) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State == Running {
		t.State = Ready
	}
}

// IsFinished returns true if the task has completed
func (t *Task) IsFinished() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.State == Finished
}

// GetState returns the current state of the task
func (t *Task) GetState() TaskState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.State
}

// GetRemaining returns the remaining CPU time
func (t *Task) GetRemaining() int64 {
	return atomic.LoadInt64(&t.Remaining)
}

// GetPriority returns the current priority level
func (t *Task) GetPriority() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Priority
}

// SetPriority sets the priority level
func (t *Task) SetPriority(priority int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Priority = priority
}

// GetWaitTime returns the total wait time
func (t *Task) GetWaitTime() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.WaitTime
}

// AddWaitTime adds to the wait time
func (t *Task) AddWaitTime(ticks int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.WaitTime += ticks
}

// String returns a string representation of the task
func (t *Task) String() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return fmt.Sprintf("Task{ID: %d, Name: %s, State: %s, Priority: %d, Remaining: %d/%d}",
		t.ID, t.Name, t.State, t.Priority, t.GetRemaining(), t.Duration)
}
