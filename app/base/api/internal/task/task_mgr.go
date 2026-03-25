package task

type TaskManager struct {
	taskCtx *TaskContext
}

func NewTaskManager(taskCtx *TaskContext) *TaskManager {
	return &TaskManager{
		taskCtx: taskCtx,
	}
}

func (m *TaskManager) StartAllTasks() {
	feeTask := NewFeeTask(m.taskCtx)
	feeTask.Start()
}
