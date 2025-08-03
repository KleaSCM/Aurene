/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: file_loader.go
Description: External task file loader for Aurene scheduler. Enables loading task definitions
from external files (TOML, JSON, CSV) to make the scheduler ready for real system integration.
*/

package workloads

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aurene/task"

	"github.com/BurntSushi/toml"
)

/**
 * TaskDefinition represents an external task specification
 *
 * Defines task parameters that can be loaded from external files,
 * enabling real system integration and dynamic workload management.
 */
type TaskDefinition struct {
	ID         int64   `json:"id" toml:"id"`
	Name       string  `json:"name" toml:"name"`
	Duration   int64   `json:"duration" toml:"duration"`
	Priority   int     `json:"priority" toml:"priority"`
	IOChance   float64 `json:"io_chance" toml:"io_chance"`
	Memory     int64   `json:"memory" toml:"memory"`
	Group      string  `json:"group" toml:"group"`
	Type       string  `json:"type" toml:"type"`
	Complexity int     `json:"complexity" toml:"complexity"`
}

/**
 * WorkloadFile represents a complete workload specification
 *
 * 外部ワークロードファイルシステム (◕‿◕)✨
 *
 * 外部ファイルからタスク定義を読み込み、
 * リアルシステム統合のための動的ワークロード管理を提供します。
 * 様々なファイル形式をサポートし、
 * 本格的なスケジューラー統合を可能にします (｡•̀ᴗ-)✧
 */
type WorkloadFile struct {
	Metadata WorkloadMetadata `json:"metadata" toml:"metadata"`
	Tasks    []TaskDefinition `json:"tasks" toml:"tasks"`
}

type WorkloadMetadata struct {
	Name        string    `json:"name" toml:"name"`
	Description string    `json:"description" toml:"description"`
	CreatedAt   time.Time `json:"created_at" toml:"created_at"`
	Version     string    `json:"version" toml:"version"`
	Author      string    `json:"author" toml:"author"`
}

/**
 * FileLoader provides external task loading capabilities
 *
 * Supports multiple file formats (TOML, JSON, CSV) and provides
 * validation and error handling for production deployment.
 */
type FileLoader struct {
	supportedFormats map[string]func(string) (*WorkloadFile, error)
}

/**
 * NewFileLoader creates a new file loader instance
 *
 * Initializes the file loader with support for multiple
 * file formats and validation capabilities.
 */
func NewFileLoader() *FileLoader {
	loader := &FileLoader{
		supportedFormats: make(map[string]func(string) (*WorkloadFile, error)),
	}

	loader.supportedFormats[".toml"] = loader.loadTOML
	loader.supportedFormats[".json"] = loader.loadJSON
	loader.supportedFormats[".csv"] = loader.loadCSV

	return loader
}

/**
 * LoadWorkloadFromFile loads task definitions from external file
 *
 * Determines file format and delegates to appropriate loader,
 * providing comprehensive error handling and validation.
 */
func (fl *FileLoader) LoadWorkloadFromFile(filePath string) (*WorkloadFile, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	loader, exists := fl.supportedFormats[ext]
	if !exists {
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	return loader(filePath)
}

/**
 * ConvertToTasks converts workload file to scheduler tasks
 *
 * Transforms external task definitions into scheduler-compatible
 * task objects for immediate integration.
 */
func (fl *FileLoader) ConvertToTasks(workload *WorkloadFile) ([]*task.Task, error) {
	tasks := make([]*task.Task, 0, len(workload.Tasks))

	for _, def := range workload.Tasks {
		if err := fl.validateTaskDefinition(&def); err != nil {
			return nil, fmt.Errorf("invalid task definition %s: %w", def.Name, err)
		}

		t := task.NewTask(
			def.ID,
			def.Name,
			def.Duration,
			def.Priority,
			def.IOChance,
			def.Memory,
			def.Group,
		)
		tasks = append(tasks, t)
	}

	return tasks, nil
}

/**
 * loadTOML loads workload from TOML file format
 *
 * ファイルローダーTOMLシステム (๑•́ ₃ •̀๑)
 *
 * TOMLファイルからワークロード定義を読み込み、
 * 構造化されたタスク仕様を解析します。
 * リアルシステム統合のための
 * 堅牢なファイル処理を提供します (｡•ㅅ•｡)♡
 */
func (fl *FileLoader) loadTOML(filePath string) (*WorkloadFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TOML file: %w", err)
	}

	var workload WorkloadFile
	if err := toml.Unmarshal(data, &workload); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	return &workload, nil
}

/**
 * loadJSON loads workload from JSON file format
 *
 * Parses JSON workload definitions and provides
 * comprehensive error handling for malformed files.
 */
func (fl *FileLoader) loadJSON(filePath string) (*WorkloadFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var workload WorkloadFile
	if err := json.Unmarshal(data, &workload); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &workload, nil
}

/**
 * loadCSV loads workload from CSV file format
 *
 * ファイルローダーCSVシステム (◕‿◕)
 *
 * CSVファイルからタスク定義を読み込み、
 * カンマ区切り形式のタスク仕様を解析します。
 * シンプルな形式で大量のタスクを
 * 効率的に処理できるようにします (｡•̀ᴗ-)✧
 */
func (fl *FileLoader) loadCSV(filePath string) (*WorkloadFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have header and at least one task")
	}

	workload := &WorkloadFile{
		Metadata: WorkloadMetadata{
			Name:        "CSV Workload",
			Description: "Loaded from CSV file",
			CreatedAt:   time.Now(),
			Version:     "1.0",
			Author:      "Aurene Scheduler",
		},
		Tasks: make([]TaskDefinition, 0, len(records)-1),
	}

	for _, record := range records[1:] {
		if len(record) < 7 {
			continue
		}

		id, _ := strconv.ParseInt(record[0], 10, 64)
		duration, _ := strconv.ParseInt(record[2], 10, 64)
		priority, _ := strconv.Atoi(record[3])
		ioChance, _ := strconv.ParseFloat(record[4], 64)
		memory, _ := strconv.ParseInt(record[5], 10, 64)
		complexity, _ := strconv.Atoi(record[6])

		task := TaskDefinition{
			ID:         id,
			Name:       record[1],
			Duration:   duration,
			Priority:   priority,
			IOChance:   ioChance,
			Memory:     memory,
			Group:      record[7],
			Type:       record[8],
			Complexity: complexity,
		}

		workload.Tasks = append(workload.Tasks, task)
	}

	return workload, nil
}

/**
 * validateTaskDefinition validates task definition parameters
 *
 * Ensures all task parameters are within valid ranges
 * and provides comprehensive error reporting.
 */
func (fl *FileLoader) validateTaskDefinition(def *TaskDefinition) error {
	if def.Name == "" {
		return fmt.Errorf("task name cannot be empty")
	}

	if def.Duration <= 0 {
		return fmt.Errorf("task duration must be positive")
	}

	if def.Priority < 0 || def.Priority > 10 {
		return fmt.Errorf("task priority must be between 0 and 10")
	}

	if def.IOChance < 0.0 || def.IOChance > 1.0 {
		return fmt.Errorf("IO chance must be between 0.0 and 1.0")
	}

	if def.Memory <= 0 {
		return fmt.Errorf("task memory must be positive")
	}

	return nil
}

/**
 * CreateSampleWorkloadFile creates a sample workload file
 *
 * Generates a sample TOML workload file for testing
 * and demonstration purposes.
 */
func (fl *FileLoader) CreateSampleWorkloadFile(filePath string) error {
	sampleWorkload := &WorkloadFile{
		Metadata: WorkloadMetadata{
			Name:        "Sample Math Workload",
			Description: "Sample workload for testing Aurene scheduler",
			CreatedAt:   time.Now(),
			Version:     "1.0",
			Author:      "Aurene Scheduler",
		},
		Tasks: []TaskDefinition{
			{
				ID:         1,
				Name:       "Matrix Multiplication",
				Duration:   1000,
				Priority:   0,
				IOChance:   0.1,
				Memory:     1024 * 1024 * 100,
				Group:      "math",
				Type:       "matrix",
				Complexity: 50,
			},
			{
				ID:         2,
				Name:       "Prime Number Check",
				Duration:   500,
				Priority:   1,
				IOChance:   0.05,
				Memory:     1024 * 1024 * 10,
				Group:      "math",
				Type:       "prime",
				Complexity: 30,
			},
			{
				ID:         3,
				Name:       "Integration",
				Duration:   2000,
				Priority:   0,
				IOChance:   0.2,
				Memory:     1024 * 1024 * 200,
				Group:      "math",
				Type:       "integration",
				Complexity: 80,
			},
		},
	}

	data, err := toml.Marshal(sampleWorkload)
	if err != nil {
		return fmt.Errorf("failed to marshal sample workload: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write sample workload: %w", err)
	}

	return nil
}
