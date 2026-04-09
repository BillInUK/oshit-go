package pos

import (
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/common/pkg/entity"
)

// HandeSnapShot 处理快照任务扫描到的 snapshot 快照
func (l *PosSnapShotLogic) HandeSnapShot(msg entity.KafkaNewSnapShotMsg) error {
	log.Infof("%s 处理快照消息: %v", l.prefix, msg)
	l.ProcessSnapShot(msg)
	return nil
}
