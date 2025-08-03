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
 * スケジューラ設定定数 (◕‿◕)
 *
 * MLFQスケジューラの核となる動作を定義する定数群です。
 * DefaultSchedulerQueues: タスクエイジングの優先度レベル数
 * DefaultTickRateHz: リアルタイムタスク配信の動作周波数
 * DefaultTickDurationMs: 1000ms / DefaultTickRateHz = 4msで計算
 * PriorityAgingInterval: Nティックごとに優先度をエイジングして飢餓を防止
 * IOUnblockProbability: ティックごとのIO完了確率（10%）
 * IOBlockingThreshold: リアルなIOブロックシミュレーションの閾値
 */
const (
	DefaultSchedulerQueues = 3
	DefaultTickRateHz      = 1000000 // 1MHz - REAL SCHEDULER SPEED!
	DefaultTickDurationMs  = 1       // 1 millisecond - FAST PRECISION!
	PriorityAgingInterval  = 1000    // Increased for faster processing
	IOUnblockProbability   = 90      // Much higher to reduce blocking
	IOBlockingThreshold    = 10      // Much lower threshold
)

/**
 * Queue Time Slice Constants
 *
 * キュー時間スライス定数 (｡♥‿♥｡)
 *
 * 各キューは指数関数的な時間スライスを持ちます：
 * time_slice = 2^(queue_index + TimeSliceExponent)
 * これにより、高優先度タスクが短い時間スライスを
 * 得て応答性を確保する公平なスケジューリング階層を作成します。
 */
const (
	Queue0TimeSlice = 10 // 10 ticks - Plenty of time!
	Queue1TimeSlice = 15 // 15 ticks - Generous time slice!
	Queue2TimeSlice = 20 // 20 ticks - Plenty of time!
)

/**
 * Task Priority Constants
 *
 * タスク優先度定数 (◡‿◡)
 *
 * MLFQスケジューリングアルゴリズムの優先度レベル。
 * 小さい数字 = 高優先度（0 = 最高、2 = 最低）
 * これによりプリエンプションと優先度ベースのタスク選択が可能になります。
 */
const (
	HighestPriority = 0
	MediumPriority  = 1
	LowestPriority  = 2
)

/**
 * Memory Constants (in bytes)
 *
 * メモリ定数（バイト単位）(◕‿◕)
 *
 * リアルなタスクモデリングのためのメモリフットプリントシミュレーション。
 * メモリを意識したスケジューリングとリソース管理を可能にします。
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
 * IO確率定数 (｡♥‿♥｡)
 *
 * リアルなIOブロックシミュレーションのための確率値。
 * タスクがIO操作でブロックする頻度を決定し、
 * スケジューラテストのためのリアルなワークロードパターンを作成します。
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
 * タスク持続時間定数（ティック単位）(◡‿◡)
 *
 * リアルなタスクモデリングのための持続時間カテゴリ。
 * 異なるワークロードタイプと様々なタスク長での
 * スケジューラ動作のテストを可能にします。
 */
const (
	ShortTaskDuration  = 1  // 1 tick - INSTANT!
	MediumTaskDuration = 2  // 2 ticks - FAST!
	LongTaskDuration   = 5  // 5 ticks - Still fast!
	MaxTaskDuration    = 10 // 10 ticks max!
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
