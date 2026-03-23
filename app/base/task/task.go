package task

import (
	"context"
	"fmt"
	"time"
)

// Task 任务接口
type Task interface {
	Start(ctx context.Context) error
	Stop() error
	Name() string
}

// BaseTask 基础任务结构
type BaseTask struct {
	name     string
	interval time.Duration
	stopChan chan struct{}
}

// NewBaseTask 创建基础任务
func NewBaseTask(name string, interval time.Duration) *BaseTask {
	return &BaseTask{
		name:     name,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Name 返回任务名称
func (t *BaseTask) Name() string {
	return t.name
}

// Stop 停止任务
func (t *BaseTask) Stop() error {
	close(t.stopChan)
	return nil
}

// RunWithInterval 以固定间隔运行任务
func (t *BaseTask) RunWithInterval(ctx context.Context, fn func(context.Context) error) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	// 立即执行一次
	if err := fn(ctx); err != nil {
		fmt.Printf("Task %s first run error: %v\n", t.name, err)
	}

	for {
		select {
		case <-t.stopChan:
			fmt.Printf("Task %s stopped\n", t.name)
			return
		case <-ticker.C:
			if err := fn(ctx); err != nil {
				fmt.Printf("Task %s error: %v\n", t.name, err)
			}
		}
	}
}
