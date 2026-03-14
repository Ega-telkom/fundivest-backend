package queue

import (
    "context"
    "encoding/json"
    
    "github.com/hibiken/asynq"
)

const TypeCertificateGeneration = "certificate:generate"

type AsynqPublisher struct {
    client *asynq.Client
}

func NewAsynqPublisher(valkeyAddr string) *AsynqPublisher {
    client := asynq.NewClient(asynq.RedisClientOpt{Addr: valkeyAddr})
    return &AsynqPublisher{client: client}
}

func (p *AsynqPublisher) Publish(ctx context.Context, certID string) error {
    payload, _ := json.Marshal(map[string]string{"certificate_id": certID})
    
    task := asynq.NewTask(TypeCertificateGeneration, payload)
    
    _, err := p.client.EnqueueContext(ctx, task)
    return err
}

func (p *AsynqPublisher) Close() error {
    return p.client.Close()
}