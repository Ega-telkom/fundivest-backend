// cmd/worker/worker.go
package main

import (
	"github.com/Ega-telkom/fundivest-backend/internal/config"
	"github.com/Ega-telkom/fundivest-backend/internal/queue"
	repoPostgres "github.com/Ega-telkom/fundivest-backend/internal/repository/postgres"
	"github.com/Ega-telkom/fundivest-backend/internal/worker"

	"go.uber.org/zap"
)

func SetupWorker(cfg *config.Config, infra *Infrastructure, logger *zap.Logger) *queue.AsynqConsumer {
    certRepo := repoPostgres.NewCertificateRepo(infra.DB)
    
    pdfGen := worker.NewGotenbergClient(cfg.GotenbergURL, logger)
    tmpl, _ := worker.NewHTMLTemplateRenderer(cfg.TemplatePath, cfg.AllowedOrigins)
    
    processor := worker.NewProcessor(certRepo, pdfGen, infra.FileStorage, tmpl, logger)
    
    return queue.NewAsynqConsumer(cfg.ValkeyAddr(), cfg.ValkeyPassword, processor)
}