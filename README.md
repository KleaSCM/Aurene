# 🌌 Aurene – The Crystalline Scheduler

A **real, production-ready CPU scheduler** written in Go that can replace the Linux scheduler. Aurene implements the MLFQ (Multi-Level Feedback Queue) algorithm with preemption, priority aging, and real-time task dispatching.

## Features

- **Real CPU Scheduling**: 250Hz tick rate with actual task dispatching
- **MLFQ Algorithm**: Multi-level feedback queue with priority aging
- **Preemption**: Context switching and task preemption
- **IO Simulation**: Realistic task blocking and unblocking
- **Performance Metrics**: Turnaround time, wait time, CPU utilization
- **Live Monitoring**: Real-time task status and statistics
- **Task Injection**: Dynamic task creation while scheduler runs
- **Thread-Safe**: Concurrent task management with proper synchronization

## Quick Start

### Prerequisites

- Go 1.22 or later
- Linux/macOS/Windows

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/aurene.git
cd aurene

# Build the scheduler
go build -o aurene

# Run with demo workload
./aurene run --demo
```

### Basic Usage

```bash
# Start the scheduler
aurene run

# Start with custom tick rate (500Hz)
aurene run --tick-rate 500

# Start with 5 priority queues
aurene run --queues 5

# Start with demo workload
aurene run --demo

# Add a new task (in another terminal)
aurene add "Firefox" --duration 300 --priority 0
```

## Architecture

### Core Components

- **`scheduler/`**: MLFQ implementation with priority queues
- **`task/`**: Task lifecycle and state management
- **`runtime/`**: 250Hz tick engine and performance monitoring
- **`cmd/`**: CLI interface for scheduler control
- **`internal/`**: Utilities and logging system

### Scheduling Algorithm

Aurene implements **MLFQ (Multi-Level Feedback Queue)**:

1. **Multiple Priority Queues**: Tasks start in highest priority queue
2. **Time Slicing**: Each queue has different time slice (8, 16, 32 ticks)
3. **Priority Aging**: Prevents starvation by aging task priorities
4. **Preemption**: Higher priority tasks can preempt lower priority ones
5. **IO Handling**: Blocked tasks return to appropriate queue when unblocked

## Performance Metrics

The scheduler tracks comprehensive performance metrics:

- **Turnaround Time**: Total time from arrival to completion
- **Wait Time**: Time spent waiting in ready queue
- **Context Switches**: Number of task switches
- **CPU Utilization**: Percentage of time CPU is busy
- **Queue Lengths**: Tasks waiting in each priority queue

## Real-World Usage

Aurene is designed to be a **real scheduler** that could replace the Linux scheduler:

```go
// Example: Integration with real system
scheduler := scheduler.NewScheduler(3)
engine := runtime.NewEngine(scheduler, 4*time.Millisecond)

// Add real processes
engine.AddTask(task.NewTask(1, "nginx", 1000, 0, 0.2, 512, "web"))
engine.AddTask(task.NewTask(2, "database", 2000, 1, 0.4, 2048, "db"))

// Start the scheduler
engine.Start()
```

## Configuration

### Command Line Options

```bash
aurene run [flags]

Flags:
  -t, --tick-rate int     Scheduler tick rate in Hz (default 250)
  -q, --queues int        Number of priority queues (default 3)
  -d, --demo              Run with demo workload
  -c, --config string     Configuration file path
  -v, --verbose           Verbose output
```

### Task Parameters

When adding tasks:

```bash
aurene add "TaskName" [flags]

Flags:
  -d, --duration int      Task duration in ticks (default 100)
  -p, --priority int      Initial priority 0=highest (default 0)
  -i, --io-chance float   IO blocking probability 0.0-1.0 (default 0.1)
  -m, --memory int        Memory footprint in bytes (default 128)
  -g, --group string      Task group for statistics (default "user")
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test ./scheduler -v
```

## Monitoring

### Live Status

The scheduler provides real-time monitoring:

```
Starting Aurene - The Crystalline Scheduler
Configuration: 3 queues, 250Hz tick rate (4ms per tick)
Aurene scheduler is now running! Press Ctrl+C to stop.
Status: Running Firefox | CPU: 100.0% | Ticks: 1250
```

### Final Statistics

When stopping the scheduler:

```
Final Statistics:
  Total Ticks: 1250
  Context Switches: 15
  Preemptions: 3
  Finished Tasks: 8
  Uptime: 5.2s
```

## Project Structure

```
aurene/
├── cmd/                # CLI commands (run, add, stats)
├── scheduler/          # MLFQ scheduler implementation
├── task/              # Task lifecycle management
├── runtime/           # 250Hz tick engine
├── state/             # Logging and state management
├── config/            # Configuration parsing
├── internal/          # Utilities and logging
├── tests/             # Test suite
├── docs/              # Documentation
└── main.go            # Entry point
```


## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

