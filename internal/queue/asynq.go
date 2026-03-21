package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const TypeCertificateGeneration = "certificate:generate"

type AsynqPublisher struct {
	client *asynq.Client
	logger *zap.Logger
}

func NewAsynqPublisher(valkeyAddr string, password string, logger *zap.Logger) *AsynqPublisher {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     valkeyAddr,
		Password: password,
	})
	
	logger.Info("Asynq publisher initialized",
		zap.String("valkey_url", valkeyAddr),
	)
	return &AsynqPublisher{
		client: client,
		logger: logger,
	}
}

func (p *AsynqPublisher) Publish(ctx context.Context, certID string) error {
	p.logger.Info("Enqueuing job",
		zap.String("cert_id", certID),
		zap.String("task_type", TypeCertificateGeneration),
	)
	
	payload := map[string]string{
		"certificate_id": certID,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		p.logger.Error("Failed to marshal payload",
			zap.String("cert_id", certID),
			zap.Error(err),
		)
		return fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeCertificateGeneration, data)
	
	info, err := p.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(5),
		asynq.Timeout(2 * time.Minute),
		asynq.Retention(24 * time.Hour),
		asynq.Queue("default"),
	)

	if err != nil {
		p.logger.Error("Failed to enqueue job",
			zap.String("cert_id", certID),
			zap.Error(err),
		)
		return fmt.Errorf("enqueue task: %w", err)
	}
	
	p.logger.Info("Job enqueued successfully",
		zap.String("cert_id", certID),
		zap.String("task_id", info.ID),
		zap.String("queue", info.Queue),
		zap.Int("max_retry", info.MaxRetry),
	)
	
	return nil
}

func (p *AsynqPublisher) Close() error {
	p.logger.Info("Closing asynq publisher")
	
	if err := p.client.Close(); err != nil {
		p.logger.Error("Failed to close asynq client", zap.Error(err))
	}
	
	p.logger.Info("Asynq publisher closed")
	return nil
}
