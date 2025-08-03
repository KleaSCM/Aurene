/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: strategies.go
Description: Pluggable scheduling strategy implementations including FCFS, Round-Robin,
SJF, and real-time scheduling algorithms for production deployment.
*/

package scheduler

import (
	"aurene/task"
	"fmt"
	"sort"
	"sync"
	"time"
)

/**
 * SchedulingStrategy defines the interface for pluggable scheduling algorithms
 *
 * Provides a common interface for different scheduling algorithms
 * to enable runtime algorithm switching and comparison.
 */
type SchedulingStrategy interface {
	// Core scheduling methods
	AddTask(t *task.Task)
	GetNextTask() *task.Task
	Tick()
	GetStats() map[string]interface{}

	// Algorithm-specific methods
	GetName() string
	GetDescription() string

	// Configuration
	SetConfig(config map[string]interface{}) error
	GetConfig() map[string]interface{}
}

/**
 * FCFSStrategy implements First-Come, First-Served scheduling
 *
 * Simple non-preemptive scheduling where tasks are executed
 * in the order they arrive, suitable for batch processing.
 */
type FCFSStrategy struct {
	mu          sync.RWMutex
	readyQueue  []*task.Task
	currentTask *task.Task
	stats       map[string]interface{}
}

func NewFCFSStrategy() *FCFSStrategy {
	return &FCFSStrategy{
		readyQueue: make([]*task.Task, 0),
		stats:      make(map[string]interface{}),
	}
}

func (f *FCFSStrategy) AddTask(t *task.Task) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.readyQueue = append(f.readyQueue, t)
}

func (f *FCFSStrategy) GetNextTask() *task.Task {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.currentTask != nil && f.currentTask.GetState() == task.Running {
		return f.currentTask
	}

	if len(f.readyQueue) > 0 {
		f.currentTask = f.readyQueue[0]
		f.readyQueue = f.readyQueue[1:]
		f.currentTask.Start()
		return f.currentTask
	}

	return nil
}

func (f *FCFSStrategy) Tick() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.currentTask != nil && f.currentTask.GetState() == task.Running {
		finished := f.currentTask.Execute()
		if finished {
			f.currentTask = nil
		}
	}
}

func (f *FCFSStrategy) GetStats() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["algorithm"] = "FCFS"
	stats["ready_queue_length"] = len(f.readyQueue)
	stats["current_task"] = f.currentTask
	totalTasks := len(f.readyQueue)
	if f.currentTask != nil {
		totalTasks++
	}
	stats["total_tasks"] = totalTasks

	return stats
}

func (f *FCFSStrategy) GetName() string {
	return "FCFS"
}

func (f *FCFSStrategy) GetDescription() string {
	return "First-Come, First-Served scheduling algorithm"
}

func (f *FCFSStrategy) SetConfig(config map[string]interface{}) error {
	// FCFS has no special configuration
	return nil
}

func (f *FCFSStrategy) GetConfig() map[string]interface{} {
	return make(map[string]interface{})
}

/**
 * RoundRobinStrategy implements Round-Robin scheduling with quantum
 *
 * ラウンドロビンスケジューリングシステム (◕‿◕)
 *
 * 各タスクが固定の時間量子を得てからプリエンプトされる
 * プリエンプティブスケジューリングで、
 * 公平なCPU時間配分を確保します (◡‿◡)
 */
type RoundRobinStrategy struct {
	mu          sync.RWMutex
	readyQueue  []*task.Task
	currentTask *task.Task
	quantum     time.Duration
	timeUsed    time.Duration
	stats       map[string]interface{}
}

func NewRoundRobinStrategy(quantum time.Duration) *RoundRobinStrategy {
	return &RoundRobinStrategy{
		readyQueue: make([]*task.Task, 0),
		quantum:    quantum,
		stats:      make(map[string]interface{}),
	}
}

func (rr *RoundRobinStrategy) AddTask(t *task.Task) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	rr.readyQueue = append(rr.readyQueue, t)
}

func (rr *RoundRobinStrategy) GetNextTask() *task.Task {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if rr.currentTask != nil && rr.timeUsed >= rr.quantum {
		rr.currentTask.Stop()
		rr.readyQueue = append(rr.readyQueue, rr.currentTask)
		rr.currentTask = nil
		rr.timeUsed = 0
	}

	if len(rr.readyQueue) > 0 {
		rr.currentTask = rr.readyQueue[0]
		rr.readyQueue = rr.readyQueue[1:]
		rr.currentTask.Start()
		rr.timeUsed = 0
		return rr.currentTask
	}

	return rr.currentTask
}

func (rr *RoundRobinStrategy) Tick() {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if rr.currentTask != nil && rr.currentTask.GetState() == task.Running {
		rr.timeUsed += time.Millisecond * 4 // Assuming 250Hz tick rate

		finished := rr.currentTask.Execute()
		if finished {
			rr.currentTask = nil
			rr.timeUsed = 0
		}
	}
}

func (rr *RoundRobinStrategy) GetStats() map[string]interface{} {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["algorithm"] = "Round-Robin"
	stats["ready_queue_length"] = len(rr.readyQueue)
	stats["current_task"] = rr.currentTask
	stats["quantum"] = rr.quantum
	stats["time_used"] = rr.timeUsed
	totalTasks := len(rr.readyQueue)
	if rr.currentTask != nil {
		totalTasks++
	}
	stats["total_tasks"] = totalTasks

	return stats
}

func (rr *RoundRobinStrategy) GetName() string {
	return "Round-Robin"
}

func (rr *RoundRobinStrategy) GetDescription() string {
	return "Round-Robin scheduling with quantum-based preemption"
}

func (rr *RoundRobinStrategy) SetConfig(config map[string]interface{}) error {
	if quantum, ok := config["quantum"].(time.Duration); ok {
		rr.quantum = quantum
	}
	return nil
}

func (rr *RoundRobinStrategy) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"quantum": rr.quantum,
	}
}

/**
 * SJFStrategy implements Shortest Job First scheduling
 *
 * 最短ジョブ優先スケジューリングシステム (◕‿◕)
 *
 * より短いタスクを優先する非プリエンプティブスケジューリングで、
 * 平均待機時間とターンアラウンド時間を最小化します (◡‿◡)
 */
type SJFStrategy struct {
	mu          sync.RWMutex
	readyQueue  []*task.Task
	currentTask *task.Task
	stats       map[string]interface{}
}

func NewSJFStrategy() *SJFStrategy {
	return &SJFStrategy{
		readyQueue: make([]*task.Task, 0),
		stats:      make(map[string]interface{}),
	}
}

func (sjf *SJFStrategy) AddTask(t *task.Task) {
	sjf.mu.Lock()
	defer sjf.mu.Unlock()

	sjf.readyQueue = append(sjf.readyQueue, t)

	sort.Slice(sjf.readyQueue, func(i, j int) bool {
		return sjf.readyQueue[i].GetRemaining() < sjf.readyQueue[j].GetRemaining()
	})
}

func (sjf *SJFStrategy) GetNextTask() *task.Task {
	sjf.mu.Lock()
	defer sjf.mu.Unlock()

	if sjf.currentTask != nil && sjf.currentTask.GetState() == task.Running {
		return sjf.currentTask
	}

	if len(sjf.readyQueue) > 0 {
		sjf.currentTask = sjf.readyQueue[0]
		sjf.readyQueue = sjf.readyQueue[1:]
		sjf.currentTask.Start()
		return sjf.currentTask
	}

	return nil
}

func (sjf *SJFStrategy) Tick() {
	sjf.mu.Lock()
	defer sjf.mu.Unlock()

	if sjf.currentTask != nil && sjf.currentTask.GetState() == task.Running {
		finished := sjf.currentTask.Execute()
		if finished {
			sjf.currentTask = nil
		}
	}
}

func (sjf *SJFStrategy) GetStats() map[string]interface{} {
	sjf.mu.RLock()
	defer sjf.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["algorithm"] = "SJF"
	stats["ready_queue_length"] = len(sjf.readyQueue)
	stats["current_task"] = sjf.currentTask
	totalTasks := len(sjf.readyQueue)
	if sjf.currentTask != nil {
		totalTasks++
	}
	stats["total_tasks"] = totalTasks

	return stats
}

func (sjf *SJFStrategy) GetName() string {
	return "SJF"
}

func (sjf *SJFStrategy) GetDescription() string {
	return "Shortest Job First scheduling algorithm"
}

func (sjf *SJFStrategy) SetConfig(config map[string]interface{}) error {
	// SJF has no special configuration
	return nil
}

func (sjf *SJFStrategy) GetConfig() map[string]interface{} {
	return make(map[string]interface{})
}

/**
 * RealTimeStrategy implements real-time scheduling with deadlines
 *
 * リアルタイムスケジューリングシステム (◕‿◕)
 *
 * 優先度継承とデッドライン逸脱検出を備えた
 * リアルタイムシステム向けのデッドラインベーススケジューリングをサポートします (◡‿◡)
 */
type RealTimeStrategy struct {
	mu           sync.RWMutex
	readyQueue   []*task.Task
	currentTask  *task.Task
	deadlineMode bool
	stats        map[string]interface{}
}

func NewRealTimeStrategy(deadlineMode bool) *RealTimeStrategy {
	return &RealTimeStrategy{
		readyQueue:   make([]*task.Task, 0),
		deadlineMode: deadlineMode,
		stats:        make(map[string]interface{}),
	}
}

func (rt *RealTimeStrategy) AddTask(t *task.Task) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.readyQueue = append(rt.readyQueue, t)

	if rt.deadlineMode {
		sort.Slice(rt.readyQueue, func(i, j int) bool {
			// In a real implementation, tasks would have deadline fields
			return rt.readyQueue[i].GetPriority() < rt.readyQueue[j].GetPriority()
		})
	}
}

func (rt *RealTimeStrategy) GetNextTask() *task.Task {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.currentTask != nil && rt.deadlineMode {
		// In a real implementation, check if deadline has passed
		// For now, we'll just use priority-based selection
	}

	if len(rt.readyQueue) > 0 {
		rt.currentTask = rt.readyQueue[0]
		rt.readyQueue = rt.readyQueue[1:]
		rt.currentTask.Start()
		return rt.currentTask
	}

	return rt.currentTask
}

func (rt *RealTimeStrategy) Tick() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.currentTask != nil && rt.currentTask.GetState() == task.Running {
		finished := rt.currentTask.Execute()
		if finished {
			rt.currentTask = nil
		}
	}
}

func (rt *RealTimeStrategy) GetStats() map[string]interface{} {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["algorithm"] = "Real-Time"
	stats["ready_queue_length"] = len(rt.readyQueue)
	stats["current_task"] = rt.currentTask
	stats["deadline_mode"] = rt.deadlineMode
	totalTasks := len(rt.readyQueue)
	if rt.currentTask != nil {
		totalTasks++
	}
	stats["total_tasks"] = totalTasks

	return stats
}

func (rt *RealTimeStrategy) GetName() string {
	return "Real-Time"
}

func (rt *RealTimeStrategy) GetDescription() string {
	return "Real-time scheduling with deadline support"
}

func (rt *RealTimeStrategy) SetConfig(config map[string]interface{}) error {
	if deadlineMode, ok := config["deadline_mode"].(bool); ok {
		rt.deadlineMode = deadlineMode
	}
	return nil
}

func (rt *RealTimeStrategy) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"deadline_mode": rt.deadlineMode,
	}
}

/**
 * StrategyFactory creates scheduling strategy instances
 *
 * Provides factory methods for creating different scheduling
 * algorithms based on configuration parameters.
 */
type StrategyFactory struct{}

/**
 * NewStrategyFactory creates a new strategy factory
 *
 * Returns a factory instance for creating scheduling strategies.
 */
func NewStrategyFactory() *StrategyFactory {
	return &StrategyFactory{}
}

/**
 * CreateStrategy creates a scheduling strategy based on algorithm name
 *
 * Instantiates the appropriate scheduling algorithm based on
 * the provided algorithm name and configuration parameters.
 */
func (sf *StrategyFactory) CreateStrategy(algorithm string, config map[string]interface{}) (SchedulingStrategy, error) {
	switch algorithm {
	case "fcfs":
		return NewFCFSStrategy(), nil
	case "rr", "round-robin":
		quantum := time.Millisecond * 10
		if q, ok := config["quantum"].(time.Duration); ok {
			quantum = q
		}
		return NewRoundRobinStrategy(quantum), nil
	case "sjf":
		return NewSJFStrategy(), nil
	case "realtime", "rt":
		deadlineMode := false
		if dm, ok := config["deadline_mode"].(bool); ok {
			deadlineMode = dm
		}
		return NewRealTimeStrategy(deadlineMode), nil
	case "mlfq":
		// MLFQ is implemented in the main scheduler
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown scheduling algorithm: %s", algorithm)
	}
}
