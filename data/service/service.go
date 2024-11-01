package service

import "ncobase/common/data/connection"

type Services struct {
	RabbitMQ *RabbitMQService
	Kafka    *KafkaService
}

func New(conn *connection.Connections) *Services {
	return &Services{
		RabbitMQ: NewRabbitMQService(conn.RMQ),
		Kafka:    NewKafkaService(conn.KFK),
	}
}

func (s *Services) Close() (errs []error) {
	// Close RabbitMQ service if connection exists
	if s.RabbitMQ != nil && s.RabbitMQ.conn != nil {
		if err := s.RabbitMQ.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Close Kafka service if connection exists
	if s.Kafka != nil && s.Kafka.conn != nil {
		if err := s.Kafka.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
