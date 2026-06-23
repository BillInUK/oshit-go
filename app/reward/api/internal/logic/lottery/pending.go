package lottery

import "sync"

// lotteryPendingMap 保存等待链上确认的 channel。
// key: txId (string)，value: chan int32（buffered size 1，防止 Kafka 消费者阻塞）
var lotteryPendingMap sync.Map
