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

type TaskState int

const (
	Ready TaskState = iota
	Running
	Blocked
	Finished
)

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

type Task struct {
	ID          int64
	Name        string
	Duration    int64
	Remaining   int64
	State       TaskState
	Priority    int
	IOChance    float64
	Memory      int64
	Group       string
	ArrivalTime time.Time
	StartTime   *time.Time
	EndTime     *time.Time

	WaitTime       int64
	TurnaroundTime int64

	mu          sync.RWMutex
	blockedAt   *time.Time
	lastRunTime *time.Time
}

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

/**
 * Execute runs the task for one tick
 *
 * タスク実行システム (｡•̀ᴗ-)✧
 *
 * 1ティック分のタスク実行を行い、
 * 残り時間を減算して完了判定を行います。
 * IOブロッキングの確率も計算して、
 * リアルなプロセス動作をシミュレートします (๑˃̵ᴗ˂̵)و
 */
func (t *Task) Execute() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State != Running {
		return false
	}

	remaining := atomic.AddInt64(&t.Remaining, -1)

	now := time.Now()
	t.lastRunTime = &now

	if remaining <= 0 {
		t.State = Finished
		t.EndTime = &now
		t.TurnaroundTime = int64(now.Sub(t.ArrivalTime).Milliseconds())
		return true
	}

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
 * IOブロッキング判定システム (๑•́ ₃ •̀๑)
 *
 * リアルなIOブロッキングシミュレーションのための
 * 適切な乱数生成を使用します。
 * 各タスクは設定されたIO確率パラメータに基づいて、
 * IO操作でブロックする確率ベースのチャンスを持ちます (｡•ㅅ•｡)♡
 */
func (t *Task) shouldBlock() bool {
	if t.IOChance <= 0 {
		return false
	}
	return rand.Float64() < t.IOChance
}

func (t *Task) Unblock() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State == Blocked {
		t.State = Ready
		t.blockedAt = nil
	}
}

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

func (t *Task) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State == Running {
		t.State = Ready
	}
}

func (t *Task) IsFinished() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.State == Finished
}

func (t *Task) GetState() TaskState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.State
}

func (t *Task) GetRemaining() int64 {
	return atomic.LoadInt64(&t.Remaining)
}

func (t *Task) GetPriority() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Priority
}

func (t *Task) SetPriority(priority int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Priority = priority
}

func (t *Task) GetWaitTime() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.WaitTime
}

func (t *Task) AddWaitTime(ticks int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.WaitTime += ticks
}

func (t *Task) String() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return fmt.Sprintf("Task{ID: %d, Name: %s, State: %s, Priority: %d, Remaining: %d/%d}",
		t.ID, t.Name, t.State, t.Priority, t.GetRemaining(), t.Duration)
}
