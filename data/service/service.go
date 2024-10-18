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
	// Close RabbitMQ service
	if err := s.RabbitMQ.Close(); err != nil {
		errs = append(errs, err)
	}
	// Close Kafka service
	if err := s.Kafka.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
