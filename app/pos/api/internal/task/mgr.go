package task

type TaskManager struct {
	taskCtx *TaskContext
}

func NewTaskManager(taskCtx *TaskContext) *TaskManager {
	return &TaskManager{
		taskCtx: taskCtx,
	}
}

// RegisterSnapShotHandler 注册已确认交易处理器，key 为 SubService 名称
func (m *TaskManager) RegisterSnapShotHandler(subService string, handler SnapShotHandler) {
	if m.taskCtx.SnapShotHandlers == nil {
		m.taskCtx.SnapShotHandlers = make(map[string]SnapShotHandler)
	}
	m.taskCtx.SnapShotHandlers[subService] = handler
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
	// PosSnapShotTask 和 StakeSnapShotTask 的 Start() 内部会 Sleep 到下次执行时间后进入无限循环，
	// 必须用 goroutine 启动，否则会阻塞后续所有任务（包括 Kafka 消费者）
	go NewPosSnapShotTask(m.taskCtx).Start()
	go NewStakeSnapShotTask(m.taskCtx).Start()
	NewKafkaConsumerTask(m.taskCtx).Start()
	NewSnapShotConsumerTask(m.taskCtx).Start()
	NewStakeBuyTokenExpireTask(m.taskCtx).Start()
}
