package task

type TaskManager struct {
	taskCtx    *TaskContext
	txScanTask *TxScanTask
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

	m.txScanTask = NewTxScanTask(m.taskCtx)
	m.txScanTask.Start()

	txExpireTask := NewTxExpireTask(m.taskCtx)
	txExpireTask.Start()

	ttlPartitionTask := NewTTLPartitionTask(m.taskCtx)
	ttlPartitionTask.Start()

	ttlCleanupTask := NewTTLCleanupTask(m.taskCtx)
	ttlCleanupTask.Start()
}

func (m *TaskManager) ReconcileScanConfigs(configs []ScanConfig) error {
	if m.txScanTask == nil {
		return nil
	}
	return m.txScanTask.Reconcile(configs)
}
