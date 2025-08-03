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

type Queue struct {
	tasks []*task.Task
	mu    sync.RWMutex
}

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

func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tasks)
}

func (q *Queue) IsEmpty() bool {
	return q.Len() == 0
}

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

func (s *Scheduler) AddTask(t *task.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queues[0].Push(t)
}

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
		if finished {
			if s.currentTask.ID <= 20 {
				fmt.Printf("[DEBUG] Task finished: %s\n", s.currentTask.Name)
			}
			s.handleTaskFinish(s.currentTask)
			s.currentTask = nil
		} else if s.currentTask.IsFinished() {
			if s.currentTask.ID <= 20 {
				fmt.Printf("[DEBUG] Task finished (IsFinished): %s\n", s.currentTask.Name)
			}
			s.handleTaskFinish(s.currentTask)
			s.currentTask = nil
		} else if s.currentTask.GetState() == task.Blocked {
			if s.currentTask.ID <= 20 {
				fmt.Printf("[DEBUG] Task blocked: %s\n", s.currentTask.Name)
			}
			s.handleTaskBlock(s.currentTask)
			s.currentTask = nil
		}
	}

	// BATCH PROCESSING: Process multiple tasks per tick for high throughput
	s.processBatchTasks()

	// Always check for preemption and dispatch if needed
	s.dispatchNextTask()

	// Call tick callback
	if s.onTick != nil {
		s.onTick(s.totalTicks)
	}
}

func (s *Scheduler) dispatchNextTask() {
	// First, check for preemption by higher priority tasks
	if s.currentTask != nil && s.currentTask.GetState() == task.Running {
		currentPriority := s.currentTask.GetPriority()
		if currentPriority >= s.numQueues {
			currentPriority = s.numQueues - 1
		}
		for i := 0; i < currentPriority; i++ {
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
					if higherTask.ID <= 20 {
						fmt.Printf("[DEBUG] Preempted and started: %s\n", higherTask.Name)
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

		// Check if task completed during this time slice FIRST
		if s.currentTask.IsFinished() {
			if s.currentTask.ID <= 20 {
				fmt.Printf("[DEBUG] Finished in time slice: %s\n", s.currentTask.Name)
			}
			s.handleTaskFinish(s.currentTask)
			s.currentTask = nil
			s.currentTimeSlice = 0
			return
		}

		// If time slice exhausted, demote task to lower priority queue
		if s.currentTimeSlice >= s.timeSlice[queueIndex] {
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
			if s.currentTask.ID <= 20 {
				fmt.Printf("[DEBUG] Demoted: %s to queue %d\n", s.currentTask.Name, newPriority)
			}
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
					if t.ID <= 20 {
						fmt.Printf("[DEBUG] Dispatched: %s from queue %d\n", t.Name, i)
					}
					return
				}
			}
		}
		fmt.Printf("[DEBUG] No tasks to dispatch. Queues: ")
		for i := 0; i < s.numQueues; i++ {
			fmt.Printf("Q%d=%d ", i, s.queues[i].Len())
		}
		fmt.Printf("\n")
	}
}

func (s *Scheduler) handleTaskFinish(t *task.Task) {
	s.finishedTasks = append(s.finishedTasks, t)

	if s.onTaskFinish != nil {
		s.onTaskFinish(t)
	}
}

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
 * 優先度エイジングシステム (◕‿◕)
 *
 * 低優先度タスクの飢餓状態を防ぐための適切な優先度エイジングを実装します。
 * タスクは時間とともに徐々に優先度が上がり、
 * 公平なスケジューリングを確保します (◡‿◡)
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

func (s *Scheduler) GetCurrentTask() *task.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentTask
}

func (s *Scheduler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_ticks"] = s.totalTicks
	stats["context_switches"] = s.contextSwitches
	stats["preemptions"] = s.preemptions
	stats["finished_tasks"] = len(s.finishedTasks)
	stats["blocked_tasks"] = len(s.blockedTasks)

	// Add individual queue lengths for demo compatibility
	for i := 0; i < s.numQueues; i++ {
		queueKey := fmt.Sprintf("queue_%d_length", i)
		stats[queueKey] = s.queues[i].Len()
	}

	return stats
}

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

/**
 * processBatchTasks processes multiple tasks per tick for high throughput
 *
 * バッチ処理システム (◕‿◕)
 *
 * マルチコアCPUの並列処理をシミュレートするための
 * 高度なバッチ処理アルゴリズムを実装します。
 * 1ティックで最大100タスクを処理し、
 * リアルタイムスケジューリングの性能を最大化します (◡‿◡)
 */
func (s *Scheduler) processBatchTasks() {
	// Process up to 100 tasks per tick for ultra-fast throughput
	maxTasksPerTick := 100
	tasksProcessed := 0

	// Process tasks from all queues in priority order
	for i := 0; i < s.numQueues && tasksProcessed < maxTasksPerTick; i++ {
		for !s.queues[i].IsEmpty() && tasksProcessed < maxTasksPerTick {
			t := s.queues[i].Pop()
			if t != nil {
				// Execute the task
				finished := t.Execute()
				if finished {
					if t.ID <= 20 {
						fmt.Printf("[DEBUG] Batch finished: %s\n", t.Name)
					}
					s.handleTaskFinish(t)
				} else if t.IsFinished() {
					if t.ID <= 20 {
						fmt.Printf("[DEBUG] Batch finished (IsFinished): %s\n", t.Name)
					}
					s.handleTaskFinish(t)
				} else if t.GetState() == task.Blocked {
					if t.ID <= 20 {
						fmt.Printf("[DEBUG] Batch blocked: %s\n", t.Name)
					}
					s.handleTaskBlock(t)
				} else {
					// Task still needs more time, put it back in queue
					s.queues[i].Push(t)
				}
				tasksProcessed++
			}
		}
	}
}

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
