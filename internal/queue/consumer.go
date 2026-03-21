// internal/queue/consumer.go
package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type JobProcessor interface {
    Process(ctx context.Context, certID string) error
}

type AsynqConsumer struct {
    server    *asynq.Server
    mux       *asynq.ServeMux
    processor JobProcessor
    logger *zap.Logger
    concur int
    queues map[string]int
}

func NewAsynqConsumer(
	valkeyAddr string, 
	password string, 
	processor JobProcessor, 
	logger *zap.Logger, 
	concur int,
	queues map[string]int,
) *AsynqConsumer {
    server := asynq.NewServer(
        asynq.RedisClientOpt{
        	Addr: valkeyAddr,
        	Password: password,
        },
        asynq.Config{
            Concurrency: concur,
            Queues: queues,
        },
    )
    
    mux := asynq.NewServeMux()
    
    consumer := &AsynqConsumer{
        server:    server,
        mux:       mux,
        processor: processor,
        logger: logger,
        concur: concur,
        queues: queues,
    }
    
    // Register handler
    mux.HandleFunc(TypeCertificateGeneration, consumer.handleCertificateGeneration)
    
    return consumer
}

func (c *AsynqConsumer) handleCertificateGeneration(ctx context.Context, task *asynq.Task) error {
    var payload map[string]string
    if err := json.Unmarshal(task.Payload(), &payload); err != nil {
        return fmt.Errorf("unmarshal payload: %w", err)
    }
    
    certID := payload["certificate_id"]
    log.Printf("Processing certificate job: %s", certID)
    
    return c.processor.Process(ctx, certID)
}

func (c *AsynqConsumer) Start() error {
    return c.server.Start(c.mux)
}

func (c *AsynqConsumer) Shutdown() {
    c.server.Shutdown()
}