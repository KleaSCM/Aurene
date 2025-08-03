/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: math_tasks.go
Description: Math computation workload generator for testing Aurene scheduler.
Creates millions of mathematical computation tasks with varying complexity,
priority levels, and memory requirements for comprehensive scheduler testing.
*/

package workloads

import (
	"aurene/task"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

/**
 * MathTask represents a mathematical computation task
 *
 * 数学計算タスクシステム (◕‿◕)✨
 *
 * 複雑な数学計算を実行するタスクを定義し、
 * 様々な計算タイプと難易度レベルをサポートします。
 * リアルなCPU負荷を生成するための
 * 数学的アルゴリズムを実装します (｡•̀ᴗ-)✧
 */
type MathTask struct {
	ID         int64
	Type       MathTaskType
	Complexity int
	Data       []float64
	Result     float64
	Completed  bool
	StartTime  time.Time
	EndTime    time.Time
}

type MathTaskType int

const (
	Addition MathTaskType = iota
	Multiplication
	Division
	SquareRoot
	Power
	Factorial
	PrimeCheck
	MatrixMultiply
	Integration
	Differentiation
)

func (mtt MathTaskType) String() string {
	switch mtt {
	case Addition:
		return "Addition"
	case Multiplication:
		return "Multiplication"
	case Division:
		return "Division"
	case SquareRoot:
		return "SquareRoot"
	case Power:
		return "Power"
	case Factorial:
		return "Factorial"
	case PrimeCheck:
		return "PrimeCheck"
	case MatrixMultiply:
		return "MatrixMultiply"
	case Integration:
		return "Integration"
	case Differentiation:
		return "Differentiation"
	default:
		return "Unknown"
	}
}

/**
 * MathWorkloadGenerator creates mathematical computation tasks
 *
 * 数学ワークロード生成システム (๑•́ ₃ •̀๑)
 *
 * 数百万の数学計算タスクを生成し、
 * 様々な複雑度と優先度でスケジューラーをテストします。
 * リアルなCPU負荷パターンをシミュレートするための
 * 多様な数学的計算を提供します (｡•ㅅ•｡)♡
 */
type MathWorkloadGenerator struct {
	mu          sync.RWMutex
	taskCounter int64
	config      MathWorkloadConfig
}

type MathWorkloadConfig struct {
	TotalTasks      int64
	TaskTypes       []MathTaskType
	ComplexityRange [2]int
	PriorityRange   [2]int
	MemoryRange     [2]int64
	IOChanceRange   [2]float64
	BatchSize       int
}

func NewMathWorkloadGenerator(config MathWorkloadConfig) *MathWorkloadGenerator {
	return &MathWorkloadGenerator{
		config: config,
	}
}

func (mwg *MathWorkloadGenerator) GenerateTasks() []*task.Task {
	mwg.mu.Lock()
	defer mwg.mu.Unlock()

	var tasks []*task.Task
	batchSize := mwg.config.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	for i := int64(0); i < mwg.config.TotalTasks; i++ {
		taskType := mwg.config.TaskTypes[rand.IntN(len(mwg.config.TaskTypes))]
		complexity := rand.IntN(mwg.config.ComplexityRange[1]-mwg.config.ComplexityRange[0]) + mwg.config.ComplexityRange[0]
		priority := rand.IntN(mwg.config.PriorityRange[1]-mwg.config.PriorityRange[0]) + mwg.config.PriorityRange[0]
		memory := rand.Int64N(mwg.config.MemoryRange[1]-mwg.config.MemoryRange[0]) + mwg.config.MemoryRange[0]
		ioChance := rand.Float64()*(mwg.config.IOChanceRange[1]-mwg.config.IOChanceRange[0]) + mwg.config.IOChanceRange[0]

		duration := mwg.calculateDuration(taskType, complexity)

		task := task.NewTask(
			mwg.taskCounter,
			fmt.Sprintf("math_%s_%d", taskType.String(), mwg.taskCounter),
			duration,
			priority,
			ioChance,
			memory,
			"math_workload",
		)

		tasks = append(tasks, task)
		mwg.taskCounter++

		if len(tasks) >= batchSize {
			break
		}
	}

	return tasks
}

func (mwg *MathWorkloadGenerator) calculateDuration(taskType MathTaskType, complexity int) int64 {
	baseDuration := int64(10)

	switch taskType {
	case Addition:
		return baseDuration * int64(complexity/10)
	case Multiplication:
		return baseDuration * int64(complexity/8)
	case Division:
		return baseDuration * int64(complexity/6)
	case SquareRoot:
		return baseDuration * int64(complexity/4)
	case Power:
		return baseDuration * int64(complexity/3)
	case Factorial:
		return baseDuration * int64(complexity*2)
	case PrimeCheck:
		return baseDuration * int64(complexity*3)
	case MatrixMultiply:
		return baseDuration * int64(complexity*5)
	case Integration:
		return baseDuration * int64(complexity*8)
	case Differentiation:
		return baseDuration * int64(complexity*6)
	default:
		return baseDuration
	}
}

func (mwg *MathWorkloadGenerator) GenerateStreamingTasks(channel chan<- *task.Task) {
	go func() {
		defer close(channel)

		for i := int64(0); i < mwg.config.TotalTasks; i++ {
			taskType := mwg.config.TaskTypes[rand.IntN(len(mwg.config.TaskTypes))]
			complexity := rand.IntN(mwg.config.ComplexityRange[1]-mwg.config.ComplexityRange[0]) + mwg.config.ComplexityRange[0]
			priority := rand.IntN(mwg.config.PriorityRange[1]-mwg.config.PriorityRange[0]) + mwg.config.PriorityRange[0]
			memory := rand.Int64N(mwg.config.MemoryRange[1]-mwg.config.MemoryRange[0]) + mwg.config.MemoryRange[0]
			ioChance := rand.Float64()*(mwg.config.IOChanceRange[1]-mwg.config.IOChanceRange[0]) + mwg.config.IOChanceRange[0]

			duration := mwg.calculateDuration(taskType, complexity)

			task := task.NewTask(
				mwg.taskCounter,
				fmt.Sprintf("math_%s_%d", taskType.String(), mwg.taskCounter),
				duration,
				priority,
				ioChance,
				memory,
				"math_workload",
			)

			channel <- task
			mwg.taskCounter++

			time.Sleep(time.Microsecond * time.Duration(rand.IntN(100)))
		}
	}()
}

func (mwg *MathWorkloadGenerator) GetStats() map[string]interface{} {
	mwg.mu.RLock()
	defer mwg.mu.RUnlock()

	return map[string]interface{}{
		"total_tasks_generated": mwg.taskCounter,
		"config":                mwg.config,
		"task_types":            len(mwg.config.TaskTypes),
	}
}

func CreateDefaultMathWorkload() *MathWorkloadGenerator {
	return NewMathWorkloadGenerator(MathWorkloadConfig{
		TotalTasks:      1000000, // 1 million tasks
		TaskTypes:       []MathTaskType{Addition, Multiplication, Division, SquareRoot, Power, Factorial, PrimeCheck, MatrixMultiply, Integration, Differentiation},
		ComplexityRange: [2]int{1, 100},
		PriorityRange:   [2]int{0, 10},
		MemoryRange:     [2]int64{1024 * 1024, 100 * 1024 * 1024}, // 1MB to 100MB
		IOChanceRange:   [2]float64{0.0, 0.3},
		BatchSize:       10000,
	})
}
