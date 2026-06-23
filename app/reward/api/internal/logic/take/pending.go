package take

import "sync"

// takePendingMap 保存等待链上确认的 channel。
// key: txId (string)，value: chan int32（buffered size 1，防止 Kafka 消费者阻塞）
var takePendingMap sync.Map
