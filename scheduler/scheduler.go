/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: scheduler.go
Description: Core MLFQ (Multi-Level Feedback Queue) scheduler implementation.
Implements priority-based scheduling with multiple queues, preemption, and priority aging
to prevent starvation. Each queue has different time slices for task execution.
*/

package scheduler

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"aurene/internal/constants"
	"aurene/task"
)

// Queue represents a priority queue for tasks
type Queue struct {
	tasks []*task.Task
	mu    sync.RWMutex
}

// NewQueue creates a new priority queue
func NewQueue() *Queue {
	return &Queue{
		tasks: make([]*task.Task, 0),
	}
}

func (q *Queue) Push(t *task.Task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = append(q.tasks, t)
}

func (q *Queue) Pop() *task.Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tasks) == 0 {
		return nil
	}

	highestPriority := q.tasks[0].GetPriority()
	highestIndex := 0

	for i, t := range q.tasks {
		if t.GetPriority() < highestPriority {
			highestPriority = t.GetPriority()
			highestIndex = i
		}
	}

	t := q.tasks[highestIndex]
	q.tasks = append(q.tasks[:highestIndex], q.tasks[highestIndex+1:]...)
	return t
}

// Peek returns the highest priority task without removing it
func (q *Queue) Peek() *task.Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.tasks) == 0 {
		return nil
	}

	highestPriority := q.tasks[0].GetPriority()
	highestIndex := 0

	for i, t := range q.tasks {
		if t.GetPriority() < highestPriority {
			highestPriority = t.GetPriority()
			highestIndex = i
		}
	}

	return q.tasks[highestIndex]
}

// Len returns the number of tasks in the queue
func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tasks)
}

// IsEmpty returns true if the queue is empty
func (q *Queue) IsEmpty() bool {
	return q.Len() == 0
}

// Scheduler implements the MLFQ scheduling algorithm
type Scheduler struct {
	queues        []*Queue
	currentTask   *task.Task
	blockedTasks  []*task.Task
	finishedTasks []*task.Task

	numQueues  int
	timeSlice  []int
	agingTicks int

	totalTicks      int64
	contextSwitches int64
	preemptions     int64

	currentTimeSlice int // Current time slice counter
	currentQueue     int // Current queue being served

	mu       sync.RWMutex
	running  bool
	tickRate time.Duration // Default 250Hz = 4ms per tick

	onTick        func(int64)
	onTaskStart   func(*task.Task)
	onTaskStop    func(*task.Task)
	onTaskFinish  func(*task.Task)
	onTaskBlock   func(*task.Task)
	onTaskUnblock func(*task.Task)
}

// NewScheduler creates a new MLFQ scheduler
func NewScheduler(numQueues int) *Scheduler {
	if numQueues <= 0 {
		numQueues = constants.DefaultSchedulerQueues
	}

	queues := make([]*Queue, numQueues)
	for i := 0; i < numQueues; i++ {
		queues[i] = NewQueue()
	}

	// Exponential time slices: 2^(queue_index + 3)
	timeSlice := make([]int, numQueues)
	for i := 0; i < numQueues; i++ {
		timeSlice[i] = constants.Queue0TimeSlice << i
	}

	return &Scheduler{
		queues:        queues,
		numQueues:     numQueues,
		timeSlice:     timeSlice,
		agingTicks:    constants.PriorityAgingInterval,
		tickRate:      4 * time.Millisecond, // 250Hz
		blockedTasks:  make([]*task.Task, 0),
		finishedTasks: make([]*task.Task, 0),
	}
}

// AddTask adds a new task to the highest priority queue
func (s *Scheduler) AddTask(t *task.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queues[0].Push(t)
}

// Tick advances the scheduler by one tick
func (s *Scheduler) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalTicks++

	if s.totalTicks%int64(s.agingTicks) == 0 {
		s.agePriorities()
	}

	// Unblock some tasks (simulate IO completion)
	s.unblockTasks()

	if s.currentTask != nil && s.currentTask.GetState() == task.Running {
		finished := s.currentTask.Execute()
		if finished || s.currentTask.IsFinished() {
			s.handleTaskFinish(s.currentTask)
			s.currentTask = nil
		} else if s.currentTask.GetState() == task.Blocked {
			s.handleTaskBlock(s.currentTask)
			s.currentTask = nil
		}
	}

	// Always check for preemption and dispatch if needed
	s.dispatchNextTask()

	// Call tick callback
	if s.onTick != nil {
		s.onTick(s.totalTicks)
	}
}

// dispatchNextTask selects and starts the next task to run
func (s *Scheduler) dispatchNextTask() {
	// First, check for preemption by higher priority tasks
	if s.currentTask != nil && s.currentTask.GetState() == task.Running {
		for i := 0; i < s.currentTask.GetPriority(); i++ {
			if !s.queues[i].IsEmpty() {
				// Higher priority task available - preempt current task
				higherTask := s.queues[i].Pop()
				if higherTask != nil {
					s.currentTask.Stop()
					priority := s.currentTask.GetPriority()
					if priority >= s.numQueues {
						priority = s.numQueues - 1
					}
					s.queues[priority].Push(s.currentTask)
					s.preemptions++

					stoppedTask := s.currentTask
					if s.onTaskStop != nil {
						s.onTaskStop(stoppedTask)
					}

					s.currentTask = nil

					s.currentTask = higherTask
					higherTask.Start()
					s.contextSwitches++
					s.currentTimeSlice = 0

					if s.onTaskStart != nil {
						s.onTaskStart(higherTask)
					}
					return
				}
			}
		}
	}

	// Check if current task has exhausted its time slice
	if s.currentTask != nil && s.currentTask.GetState() == task.Running {
		s.currentTimeSlice++
		queueIndex := s.currentTask.GetPriority()
		if queueIndex >= s.numQueues {
			queueIndex = s.numQueues - 1
		}

		// If time slice exhausted, demote task to lower priority queue
		if s.currentTimeSlice >= s.timeSlice[queueIndex] {
			// Check if task completed during this time slice
			if s.currentTask.IsFinished() {
				s.handleTaskFinish(s.currentTask)
				s.currentTask = nil
				s.currentTimeSlice = 0
				return
			}

			s.currentTask.Stop()
			newPriority := queueIndex + 1
			if newPriority >= s.numQueues {
				newPriority = s.numQueues - 1
			}
			stoppedTask := s.currentTask
			if s.onTaskStop != nil {
				s.onTaskStop(stoppedTask)
			}

			s.currentTask.SetPriority(newPriority)
			s.queues[newPriority].Push(s.currentTask)
			s.currentTask = nil
			s.currentTimeSlice = 0
		}
	}

	// If no current task, find highest priority non-empty queue
	if s.currentTask == nil {
		for i := 0; i < s.numQueues; i++ {
			if !s.queues[i].IsEmpty() {
				t := s.queues[i].Pop()
				if t != nil {
					s.currentTask = t
					t.Start()
					s.contextSwitches++
					s.currentTimeSlice = 0

					if s.onTaskStart != nil {
						s.onTaskStart(t)
					}
					return
				}
			}
		}
	}
}

// handleTaskFinish processes a completed task
func (s *Scheduler) handleTaskFinish(t *task.Task) {
	s.finishedTasks = append(s.finishedTasks, t)

	if s.onTaskFinish != nil {
		s.onTaskFinish(t)
	}
}

// handleTaskBlock processes a task that has blocked on IO
func (s *Scheduler) handleTaskBlock(t *task.Task) {
	s.blockedTasks = append(s.blockedTasks, t)

	if s.onTaskBlock != nil {
		s.onTaskBlock(t)
	}
}

/**
 * unblockTasks simulates realistic IO completion for blocked tasks
 *
 * Uses proper random number generation for IO completion simulation.
 * Each blocked task has a probability-based chance to complete IO
 * and return to the ready queue for scheduling.
 */
func (s *Scheduler) unblockTasks() {
	remaining := make([]*task.Task, 0)

	for _, t := range s.blockedTasks {
		// Use proper random number generation for IO completion
		// Formula: unblock_if(random_float < IOUnblockProbability / 100.0)
		if rand.Float64() < float64(constants.IOUnblockProbability)/100.0 {
			t.Unblock()
			// Add back to appropriate queue based on current priority
			priority := t.GetPriority()
			if priority >= s.numQueues {
				priority = s.numQueues - 1
			}
			s.queues[priority].Push(t)

			if s.onTaskUnblock != nil {
				s.onTaskUnblock(t)
			}
		} else {
			remaining = append(remaining, t)
		}
	}

	s.blockedTasks = remaining
}

/**
 * agePriorities prevents starvation by aging task priorities
 *
 * 優先度エイジングシステム (◕‿◕)✨
 *
 * 低優先度タスクの飢餓状態を防ぐための適切な優先度エイジングを実装します。
 * タスクは時間とともに徐々に優先度が上がり、
 * 公平なスケジューリングを確保します (◡‿◡)💕
 */
func (s *Scheduler) agePriorities() {
	if s.currentTask != nil {
		newPriority := s.currentTask.GetPriority() + constants.PriorityAgingFactor
		if newPriority >= s.numQueues {
			newPriority = s.numQueues - 1
		}
		s.currentTask.SetPriority(newPriority)
	}

	for i := 0; i < s.numQueues; i++ {
		queue := s.queues[i]
		tasks := make([]*task.Task, 0)

		for !queue.IsEmpty() {
			t := queue.Pop()
			if t != nil {
				tasks = append(tasks, t)
			}
		}

		for _, t := range tasks {
			newPriority := t.GetPriority() + constants.PriorityAgingFactor
			if newPriority >= s.numQueues {
				newPriority = s.numQueues - 1
			}
			t.SetPriority(newPriority)
			s.queues[newPriority].Push(t)
		}
	}
}

// Preempt forces a context switch
func (s *Scheduler) Preempt() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentTask != nil {
		s.currentTask.Stop()
		newPriority := s.currentTask.GetPriority() + 1
		if newPriority >= s.numQueues {
			newPriority = s.numQueues - 1
		}
		s.currentTask.SetPriority(newPriority)
		s.queues[newPriority].Push(s.currentTask)

		if s.onTaskStop != nil {
			s.onTaskStop(s.currentTask)
		}

		s.currentTask = nil
		s.preemptions++
	}
}

// GetCurrentTask returns the currently running task
func (s *Scheduler) GetCurrentTask() *task.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentTask
}

// GetStats returns scheduler statistics
func (s *Scheduler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_ticks"] = s.totalTicks
	stats["context_switches"] = s.contextSwitches
	stats["preemptions"] = s.preemptions
	stats["finished_tasks"] = len(s.finishedTasks)
	stats["blocked_tasks"] = len(s.blockedTasks)

	queueStats := make([]int, s.numQueues)
	for i := 0; i < s.numQueues; i++ {
		queueStats[i] = s.queues[i].Len()
	}
	stats["queue_lengths"] = queueStats

	return stats
}

// SetCallbacks sets the scheduler callbacks
func (s *Scheduler) SetCallbacks(onTick func(int64), onTaskStart, onTaskStop, onTaskFinish, onTaskBlock, onTaskUnblock func(*task.Task)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.onTick = onTick
	s.onTaskStart = onTaskStart
	s.onTaskStop = onTaskStop
	s.onTaskFinish = onTaskFinish
	s.onTaskBlock = onTaskBlock
	s.onTaskUnblock = onTaskUnblock
}

// String returns a string representation of the scheduler state
func (s *Scheduler) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var currentTaskStr string
	if s.currentTask != nil {
		currentTaskStr = s.currentTask.String()
	} else {
		currentTaskStr = "None"
	}

	return fmt.Sprintf("Scheduler{Current: %s, TotalTicks: %d, ContextSwitches: %d}",
		currentTaskStr, s.totalTicks, s.contextSwitches)
}
