package svc

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func (s *ServiceContext) initKafkaConsumer() error {
	cfg := s.Config.Kafka
	// 优先使用独立的 topic 字段，兼容旧的 topics[0] 写法
	topic := cfg.Consumer.Topic
	if topic == "" && len(cfg.Consumer.Topics) > 0 {
		topic = cfg.Consumer.Topics[0]
	}
	if len(cfg.Brokers) == 0 || cfg.Consumer.GroupID == "" || topic == "" {
		return nil
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.Consumer.GroupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	s.KafkaConsumer = reader
	fmt.Println("Kafka consumer initialized successfully, topic:", topic)
	return nil
}

func (s *ServiceContext) initSnapShotKafkaConsumer() error {
	cfg := s.Config.Kafka
	if len(cfg.Brokers) == 0 || cfg.Consumer.GroupID == "" {
		return nil
	}

	// 优先使用独立的 snapshot_topics 字段，兼容旧的硬编码写法
	snapshotTopics := cfg.Consumer.SnapshotTopics
	if len(snapshotTopics) == 0 {
		snapshotTopics = []string{"PosTopic", "StakeTopic"}
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.Consumer.GroupID + "-snapshot",
		GroupTopics: snapshotTopics,
		MinBytes:    10e3,
		MaxBytes:    10e6,
	})

	s.SnapShotKafkaConsumer = reader
	fmt.Println("Snapshot Kafka consumer initialized successfully, topics:", snapshotTopics)
	return nil
}

func (s *ServiceContext) closeKafkaConsumer() error {
	if s.KafkaConsumer != nil {
		if reader, ok := s.KafkaConsumer.(*kafka.Reader); ok {
			if err := reader.Close(); err != nil {
				return fmt.Errorf("关闭Kafka消费者错误: %v", err)
			}
		}
	}
	return nil
}

func (s *ServiceContext) initKafkaProducer() error {
	if s.Config == nil || len(s.Config.Kafka.Brokers) == 0 {
		// 如果没有配置Kafka，则跳过初始化
		return nil
	}

	brokers := s.Config.Kafka.Brokers

	// 创建Kafka生产者
	producer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	s.KafkaProducer = producer
	fmt.Println("Kafka producer initialized successfully")
	return nil
}

func (s *ServiceContext) closeKafkaProducer() error {
	if s.KafkaProducer != nil {
		if writer, ok := s.KafkaProducer.(*kafka.Writer); ok {
			if err := writer.Close(); err != nil {
				return fmt.Errorf("关闭Kafka生产者错误: %v", err)
			}
		}
	}
	return nil
}

// SendKafkaMessage 发送Kafka消息
func (s *ServiceContext) SendKafkaMessage(topic string, key, value []byte) error {
	if s.KafkaProducer == nil {
		return fmt.Errorf("Kafka生产者未初始化")
	}

	writer, ok := s.KafkaProducer.(*kafka.Writer)
	if !ok {
		return fmt.Errorf("Kafka生产者类型错误")
	}

	message := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	return writer.WriteMessages(context.Background(), message)
}
