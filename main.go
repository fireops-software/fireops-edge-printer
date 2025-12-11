package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fireops-software/fireops-edge-printer/api"
	"github.com/fireops-software/fireops-edge-printer/services"
	"github.com/uoul/go-common/config"
	"github.com/uoul/go-common/log"
	"github.com/uoul/go-common/messaging"
)

const (
	VERSION      = "{VERSION}"
	SERVICE_NAME = "fireops-edge-printer"
	DISPLAY_NAME = "Drucker"
)

func main() {
	// Create application context
	appCtx, appCtxCancel := context.WithCancel(context.Background())

	// Create ConfigProvider
	cp := config.NewEnvVarProvider()
	// Create Logger
	logger := log.NewConsoleLogger(
		log.StringToLogLevel(
			cp.StringOrDefault("LOG_LEVEL", "INFO"),
			log.INFO,
		),
	)
	// Create RabbitMq client
	rabbitMq := messaging.NewRabbitMqMessenger(
		appCtx,
		logger,
		cp.StringOrDefault("RABBITMQ_HOST", ""),
		cp.UInt16OrDefault("RABBITMQ_PORT", 5672),
		cp.StringOrDefault("RABBITMQ_USER", ""),
		cp.StringOrDefault("RABBITMQ_PW", ""),
	)
	// Create PrinterApi
	printerApi := api.NewPrinterApi(
		appCtx,
		logger,
		cp.StringOrDefault("PRINTER_NAME", ""),
		cp.StringOrDefault("PRINTER_URL", ""),
		api.WithPrinterDriver(
			cp.StringOrDefault("PRINTER_DRIVER", "everywhere"),
		),
	)
	// Create PrintService
	services.NewEventPrinter(
		appCtx,
		logger,
		rabbitMq,
		messaging.RabbitMqExchange{
			Type:       "topic",
			Exchange:   cp.StringOrDefault("RABBITMQ_EVENTS_EXCHANGE", "fireops-edge-events"),
			RoutingKey: cp.StringOrDefault("RABBITMQ_EVENTS_ROUTING_KEY", "new"),
		},
		printerApi,
		cp.StringOrDefault("FIREDEP_ADDR", ""),
		services.WithEventPrinterCopies(
			cp.IntOrDefault("PRINTER_COPIES", 1),
		),
	)
	// Create HealthReporter
	services.NewHealthReporter(
		appCtx,
		logger,
		rabbitMq,
		messaging.RabbitMqExchange{
			Type:       "topic",
			Exchange:   cp.StringOrDefault("RABBITMQ_HEALTH_EXCHANGE", "fireops-edge-health"),
			RoutingKey: cp.StringOrDefault("RABBITMQ_HEALTH_ROUTING_KEY", ""),
		},
		SERVICE_NAME,
		DISPLAY_NAME,
	)
	// Wait until stop
	osSig := make(chan os.Signal, 1)
	signal.Notify(osSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-osSig
	appCtxCancel()
	logger.Infof("Shutting down...")
}
