/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: constants.go
Description: Named constants for the Aurene scheduler to replace magic numbers
and improve code maintainability and readability.
*/

package constants

/**
 * Scheduler Configuration Constants
 *
 * These constants define the core behavior of the MLFQ scheduler.
 * DefaultSchedulerQueues: Number of priority levels for task aging
 * DefaultTickRateHz: Operating frequency for real-time task dispatching
 * DefaultTickDurationMs: Calculated as 1000ms / DefaultTickRateHz = 4ms
 * PriorityAgingInterval: Prevents starvation by aging priorities every N ticks
 * IOUnblockProbability: Percentage chance of IO completion per tick (10%)
 * IOBlockingThreshold: Threshold for realistic IO blocking simulation
 */
const (
	DefaultSchedulerQueues = 3
	DefaultTickRateHz      = 250
	DefaultTickDurationMs  = 4
	PriorityAgingInterval  = 100
	IOUnblockProbability   = 50
	IOBlockingThreshold    = 100
)

/**
 * Queue Time Slice Constants
 *
 * Each queue has an exponential time slice following the formula:
 * time_slice = 2^(queue_index + TimeSliceExponent)
 * This creates a fair scheduling hierarchy where higher priority
 * tasks get shorter time slices for responsiveness.
 */
const (
	Queue0TimeSlice = 50
	Queue1TimeSlice = 100
	Queue2TimeSlice = 200
)

/**
 * Task Priority Constants
 *
 * Priority levels for MLFQ scheduling algorithm.
 * Lower numbers = higher priority (0 = highest, 2 = lowest)
 * This enables preemption and priority-based task selection.
 */
const (
	HighestPriority = 0
	MediumPriority  = 1
	LowestPriority  = 2
)

/**
 * Memory Constants (in bytes)
 *
 * Memory footprint simulation for realistic task modeling.
 * Enables memory-aware scheduling and resource management.
 */
const (
	DefaultTaskMemory = 128
	SmallTaskMemory   = 64
	LargeTaskMemory   = 1024
	MaxTaskMemory     = 8192
)

/**
 * IO Probability Constants
 *
 * Probability values for realistic IO blocking simulation.
 * Determines how often tasks block on IO operations,
 * creating realistic workload patterns for scheduler testing.
 */
const (
	NoIOChance     = 0.0
	LowIOChance    = 0.1
	MediumIOChance = 0.3
	HighIOChance   = 0.5
	MaxIOChance    = 1.0
)

/**
 * Task Duration Constants (in ticks)
 *
 * Duration categories for realistic task modeling.
 * Enables testing of different workload types and
 * scheduler behavior under various task lengths.
 */
const (
	ShortTaskDuration  = 50
	MediumTaskDuration = 200
	LongTaskDuration   = 500
	MaxTaskDuration    = 10000
)

// Logging Constants
const (
	// LogTickInterval is how often to log tick information (every 1000 ticks)
	LogTickInterval = 1000

	// StatusUpdateInterval is the interval for status updates (5 seconds)
	StatusUpdateIntervalSeconds = 5
)

// Mathematical Constants for Scheduler Algorithms
const (
	// PriorityAgingFactor is the factor by which priorities age
	// Formula: new_priority = min(old_priority + 1, max_priority)
	PriorityAgingFactor = 1

	// TimeSliceExponent is the base for exponential time slices
	// Formula: time_slice = 2^(queue_index + TimeSliceExponent)
	TimeSliceExponent = 3

	// CPUUtilizationScale is the scale factor for CPU utilization calculation
	// Formula: utilization = (running_tasks / total_ticks) * CPUUtilizationScale
	CPUUtilizationScale = 100.0
)
