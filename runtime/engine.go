/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: engine.go
Description: Runtime engine that drives the Aurene scheduler at configurable tick rates.
Manages the main tick loop, performance monitoring, and provides callbacks for
task lifecycle events and system integration.
*/

package runtime

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"aurene/internal/constants"
	"aurene/internal/logger"
	"aurene/scheduler"
	"aurene/task"
)

// Engine represents the runtime engine that drives the scheduler
type Engine struct {
	scheduler *scheduler.Scheduler
	logger    *logger.Logger
	tickRate  time.Duration
	running   int32
	stopChan  chan struct{}

	// Performance monitoring
	startTime      time.Time
	totalTicks     int64
	cpuUtilization float64

	// Callbacks
	onTick      func(int64)
	onTaskEvent func(string, *task.Task)

	// Thread safety
	mu sync.RWMutex

	// IPC server for task injection
	ipcServer net.Listener
	ipcPort   int
}

// NewEngine creates a new runtime engine
func NewEngine(sched *scheduler.Scheduler, tickRate time.Duration) *Engine {
	if tickRate <= 0 {
		tickRate = time.Duration(constants.DefaultTickDurationMs) * time.Millisecond
	}

	eng := &Engine{
		scheduler: sched,
		logger:    logger.New(),
		tickRate:  tickRate,
		stopChan:  make(chan struct{}),
		ipcPort:   8080,
	}

	// Set up scheduler callbacks
	sched.SetCallbacks(
		eng.onSchedulerTick,
		eng.onTaskStart,
		eng.onTaskStop,
		eng.onTaskFinish,
		eng.onTaskBlock,
		eng.onTaskUnblock,
	)

	return eng
}

// Start begins the runtime engine
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if atomic.LoadInt32(&e.running) == 1 {
		return fmt.Errorf("engine is already running")
	}

	e.startTime = time.Now()
	atomic.StoreInt32(&e.running, 1)

	e.logger.Info("Starting Aurene runtime engine at %v tick rate", e.tickRate)

	// Start IPC server for task injection
	if err := e.startIPCServer(); err != nil {
		e.logger.Error("Failed to start IPC server: %v", err)
		return err
	}

	// Start the tick loop in a goroutine
	go e.tickLoop()

	return nil
}

// Stop gracefully stops the runtime engine
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if atomic.LoadInt32(&e.running) == 0 {
		return
	}

	e.logger.Info("Stopping Aurene runtime engine...")

	// Stop IPC server
	if e.ipcServer != nil {
		e.ipcServer.Close()
	}

	atomic.StoreInt32(&e.running, 0)
	close(e.stopChan)
}

/**
 * startIPCServer launches the IPC server for task injection
 *
 * Starts a TCP server that accepts task injection commands
 * from CLI tools, enabling real-time task management.
 */
func (e *Engine) startIPCServer() error {
	// Try different ports if 8080 is busy
	ports := []int{8080, 8081, 8082, 8083, 8084}

	for _, port := range ports {
		listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			e.ipcServer = listener
			e.ipcPort = port
			e.logger.Info("IPC server started on localhost:%d", port)

			// Handle IPC connections in background
			go e.handleIPCConnections()
			return nil
		}
		e.logger.Warn("Port %d busy, trying next port...", port)
	}

	// If all ports are busy, just log a warning and continue
	e.logger.Warn("All IPC ports busy, continuing without IPC server")
	return nil
}

/**
 * handleIPCConnections processes incoming IPC connections
 *
 * Accepts TCP connections and processes task injection commands
 * from CLI tools, enabling real-time task management.
 */
func (e *Engine) handleIPCConnections() {
	for {
		conn, err := e.ipcServer.Accept()
		if err != nil {
			if atomic.LoadInt32(&e.running) == 0 {
				return // Engine stopped
			}
			e.logger.Error("IPC connection error: %v", err)
			continue
		}

		go e.handleIPCConnection(conn)
	}
}

/**
 * handleIPCConnection processes a single IPC connection
 *
 * Reads task injection commands and adds tasks to the scheduler
 * in real-time without stopping the scheduler operation.
 */
func (e *Engine) handleIPCConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		e.logger.Error("Failed to read IPC command: %v", err)
		return
	}

	line = strings.TrimSpace(line)
	parts := strings.Split(line, ":")

	if len(parts) < 7 || parts[0] != "ADD_TASK" {
		conn.Write([]byte("ERROR: Invalid command format\n"))
		return
	}

	// Parse task parameters
	id, _ := strconv.ParseInt(parts[1], 10, 64)
	name := parts[2]
	duration, _ := strconv.ParseInt(parts[3], 10, 64)
	priority, _ := strconv.Atoi(parts[4])
	ioChance, _ := strconv.ParseFloat(parts[5], 64)
	memory, _ := strconv.ParseInt(parts[6], 10, 64)
	group := parts[7]

	// Create and add task
	t := task.NewTask(id, name, duration, priority, ioChance, memory, group)
	e.scheduler.AddTask(t)

	e.logger.Info("Task injected via IPC: %s", t.String())
	conn.Write([]byte("OK\n"))
}

// tickLoop is the main tick loop that drives the scheduler
func (e *Engine) tickLoop() {
	ticker := time.NewTicker(e.tickRate)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			return
		case <-ticker.C:
			e.tick()
		}
	}
}

// tick advances the scheduler by one tick
func (e *Engine) tick() {
	// Advance the scheduler
	e.scheduler.Tick()

	// Update CPU utilization
	e.updateCPUUtilization()

	// Call tick callback
	if e.onTick != nil {
		e.onTick(e.totalTicks)
	}
}

// updateCPUUtilization calculates the current CPU utilization
func (e *Engine) updateCPUUtilization() {
	// CPU utilization calculation based on task execution
	// Formula: utilization = (running_tasks / total_ticks) * CPUUtilizationScale
	if e.totalTicks > 0 {
		currentTask := e.scheduler.GetCurrentTask()
		if currentTask != nil && currentTask.GetState() == task.Running {
			e.cpuUtilization = float64(constants.CPUUtilizationScale)
		} else {
			e.cpuUtilization = 0.0
		}
	}
}

// AddTask adds a new task to the scheduler
func (e *Engine) AddTask(t *task.Task) {
	e.scheduler.AddTask(t)
}

// GetStats returns engine statistics
func (e *Engine) GetStats() map[string]interface{} {
	schedulerStats := e.scheduler.GetStats()
	stats := make(map[string]interface{})

	// Copy scheduler stats
	for k, v := range schedulerStats {
		stats[k] = v
	}

	// Add engine-specific stats
	stats["engine_uptime"] = time.Since(e.startTime)
	stats["cpu_utilization"] = e.cpuUtilization
	stats["total_ticks"] = e.totalTicks

	return stats
}

// GetCurrentTask returns the currently running task
func (e *Engine) GetCurrentTask() *task.Task {
	return e.scheduler.GetCurrentTask()
}

// IsRunning returns true if the engine is running
func (e *Engine) IsRunning() bool {
	return atomic.LoadInt32(&e.running) == 1
}

// SetCallbacks sets the engine callbacks
func (e *Engine) SetCallbacks(onTick func(int64), onTaskEvent func(string, *task.Task)) {
	e.onTick = onTick
	e.onTaskEvent = onTaskEvent
}

// Scheduler callbacks
func (e *Engine) onSchedulerTick(tick int64) {
	e.totalTicks = tick
}

func (e *Engine) onTaskStart(t *task.Task) {
	e.logger.Info("Task started: %s", t.Name)
	if e.onTaskEvent != nil {
		e.onTaskEvent("start", t)
	}
}

func (e *Engine) onTaskStop(t *task.Task) {
	if t != nil {
		e.logger.Info("Task stopped: %s", t.Name)
		if e.onTaskEvent != nil {
			e.onTaskEvent("stop", t)
		}
	}
}

func (e *Engine) onTaskFinish(t *task.Task) {
	e.logger.Info("Task finished: %s", t.Name)
	if e.onTaskEvent != nil {
		e.onTaskEvent("finish", t)
	}
}

func (e *Engine) onTaskBlock(t *task.Task) {
	e.logger.Info("Task blocked: %s", t.Name)
	if e.onTaskEvent != nil {
		e.onTaskEvent("block", t)
	}
}

func (e *Engine) onTaskUnblock(t *task.Task) {
	e.logger.Info("Task unblocked: %s", t.Name)
	if e.onTaskEvent != nil {
		e.onTaskEvent("unblock", t)
	}
}

// String returns a string representation of the engine state
func (e *Engine) String() string {
	return fmt.Sprintf("Engine{Running: %t, Ticks: %d, CPU: %.1f%%}",
		e.IsRunning(), e.totalTicks, e.cpuUtilization)
}
