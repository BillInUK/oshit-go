package task

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"sync"
)

// TaskManager 任务管理器
type TaskManager struct {
	tasks []Task
	mu    sync.RWMutex
}

// NewTaskManager 创建任务管理器
func NewTaskManager(db *gorm.DB, redis *redis.Client, rpcClient *rpc.Client) *TaskManager {
	return &TaskManager{
		tasks: []Task{
			NewFeeTask(db, redis, rpcClient),
			NewComputeUnitTask(redis),
			NewPriceTask(redis),
		},
	}
}

// StartAll 启动所有任务
func (m *TaskManager) StartAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, task := range m.tasks {
		if err := task.Start(ctx); err != nil {
			return fmt.Errorf("start task %s error: %v", task.Name(), err)
		}
		fmt.Printf("Task %s started\n", task.Name())
	}

	return nil
}

// StopAll 停止所有任务
func (m *TaskManager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, task := range m.tasks {
		if err := task.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("stop task %s error: %v", task.Name(), err))
		} else {
			fmt.Printf("Task %s stopped\n", task.Name())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping tasks: %v", errs)
	}

	return nil
}

// GetTask 获取指定任务
func (m *TaskManager) GetTask(name string) Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, task := range m.tasks {
		if task.Name() == name {
			return task
		}
	}

	return nil
}

// GetFeeTask 获取手续费任务
func (m *TaskManager) GetFeeTask() *FeeTask {
	if task := m.GetTask("fee_task"); task != nil {
		if feeTask, ok := task.(*FeeTask); ok {
			return feeTask
		}
	}
	return nil
}

// GetComputeUnitTask 获取计算单元任务
func (m *TaskManager) GetComputeUnitTask() *ComputeUnitTask {
	if task := m.GetTask("compute_unit_task"); task != nil {
		if computeUnitTask, ok := task.(*ComputeUnitTask); ok {
			return computeUnitTask
		}
	}
	return nil
}

// GetPriceTask 获取价格任务
func (m *TaskManager) GetPriceTask() *PriceTask {
	if task := m.GetTask("price_task"); task != nil {
		if priceTask, ok := task.(*PriceTask); ok {
			return priceTask
		}
	}
	return nil
}
