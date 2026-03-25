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

	unitTask := NewUnitTask(m.taskCtx)
	unitTask.Start()

	priceTask := NewPriceTask(m.taskCtx)
	priceTask.Start()

	klineTask := NewKLineTask(m.taskCtx)
	klineTask.Start()

	holdersTask := NewHoldersTask(m.taskCtx)
	holdersTask.Start()
}
