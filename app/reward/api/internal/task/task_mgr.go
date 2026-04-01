package task

type TaskManager struct {
	taskCtx *TaskContext
}

func NewTaskManager(taskCtx *TaskContext) *TaskManager {
	return &TaskManager{
		taskCtx: taskCtx,
	}
}

// RegisterScannedTxHandler 注册已确认交易处理器，key 为 SubService 名称
func (m *TaskManager) RegisterScannedTxHandler(subService string, handler ScannedTxHandler) {
	if m.taskCtx.ScannedHandlers == nil {
		m.taskCtx.ScannedHandlers = make(map[string]ScannedTxHandler)
	}
	m.taskCtx.ScannedHandlers[subService] = handler
}

// RegisterExpiredTxHandler 注册超时交易处理器，key 为 SubService 名称
func (m *TaskManager) RegisterExpiredTxHandler(subService string, handler ExpiredTxHandler) {
	if m.taskCtx.ExpiredHandlers == nil {
		m.taskCtx.ExpiredHandlers = make(map[string]ExpiredTxHandler)
	}
	m.taskCtx.ExpiredHandlers[subService] = handler
}

func (m *TaskManager) StartAllTasks() {
	NewKafkaConsumerTask(m.taskCtx).Start()
}
