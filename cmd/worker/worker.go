// cmd/worker/worker.go
package main

import (
	"fundivest/internal/config"
	"fundivest/internal/queue"
	repoPostgres "fundivest/internal/repository/postgres"
	"fundivest/internal/worker"

	"go.uber.org/zap"
)

func SetupWorker(cfg *config.Config, infra *Infrastructure, logger *zap.Logger) *queue.AsynqConsumer {
    certRepo := repoPostgres.NewCertificateRepo(infra.DB)
    
    pdfGen := worker.NewGotenbergClient(cfg.GotenbergURL)
    tmpl, _ := worker.NewHTMLTemplateRenderer(cfg.TemplatePath, cfg.FrontendURL)
    
    processor := worker.NewProcessor(certRepo, pdfGen, infra.FileStorage, tmpl, logger)
    
    return queue.NewAsynqConsumer(cfg.ValkeyURL, processor)
}