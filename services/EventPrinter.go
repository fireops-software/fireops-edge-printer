package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fireops-software/fireops-edge-printer/api"
	"github.com/fireops-software/fireops-edge-printer/domain"
	"github.com/fireops-software/fireops-edge-printer/utils"
	"github.com/rabbitmq/amqp091-go"
	"github.com/uoul/go-common/log"
	"github.com/uoul/go-common/messaging"

	appError "github.com/fireops-software/fireops-edge-printer/error"
)

//--------------------------------------------------------------------------------------------
// Types
//--------------------------------------------------------------------------------------------

type EventPrinter struct {
	ctx      context.Context
	logger   log.ILogger
	rabbitMq messaging.IMessenger[messaging.RabbitMqExchange, amqp091.Delivery]
	exchange messaging.RabbitMqExchange
	printer  api.IPrinterApi

	copies        int
	printTimeOut  time.Duration
	retryInterval time.Duration
}

//--------------------------------------------------------------------------------------------
// Public
//--------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------
// Private
// --------------------------------------------------------------------------------------------

func (e *EventPrinter) run() error {
	// Subscribe for new events
	eventsCh := e.rabbitMq.Subscribe(e.exchange)
	defer e.rabbitMq.Unsubscribe(eventsCh)

	for {
		select {
		case <-e.ctx.Done():
			return nil
		case msg := <-eventsCh:
			if msg.Error != nil {
				return msg.Error
			}
			events := []domain.Event{}
			if err := json.Unmarshal(msg.Result.Body, &events); err != nil {
				return appError.NewErrDataParsing("failed to parse incomming data - %v", err)
			}
			e.logger.Debugf("New incomming events: %s", string(msg.Result.Body))
			// Create PDF
			pdf, err := utils.CreatePdfFromEvents(events)
			if err != nil {
				return appError.NewErrDataParsing("failed to create pdf - %v", err)
			}
			// Create context for time
			pCtx, cancel := context.WithTimeout(e.ctx, e.printTimeOut)
			// Print PDF
			if err := e.printer.PrintPdf(pCtx, pdf, e.copies); err != nil {
				cancel()
				return err
			}
			cancel()
			e.logger.Debugf("Printjob sent successfully")
		}
	}
}

func WithEventPrinterCopies(copies int) func(*EventPrinter) {
	return func(ep *EventPrinter) {
		ep.copies = copies
	}
}

func WithEventPrinterTimeout(timeout time.Duration) func(*EventPrinter) {
	return func(ep *EventPrinter) {
		ep.printTimeOut = timeout
	}
}

//--------------------------------------------------------------------------------------------
// Constructor
//--------------------------------------------------------------------------------------------

func NewEventPrinter(ctx context.Context, logger log.ILogger, rabbitMq messaging.IMessenger[messaging.RabbitMqExchange, amqp091.Delivery], exchange messaging.RabbitMqExchange, printer api.IPrinterApi, opts ...func(*EventPrinter)) *EventPrinter {
	e := &EventPrinter{
		ctx:      ctx,
		logger:   logger,
		rabbitMq: rabbitMq,
		exchange: exchange,
		printer:  printer,

		retryInterval: 10 * time.Second,
		copies:        1,
		printTimeOut:  30 * time.Second,
	}
	for _, o := range opts {
		o(e)
	}
	go func() {
		for {
			err := e.run()
			if err == nil {
				// Context exeeded
				break
			}
			logger.Errorf("%v", err)
			time.Sleep(e.retryInterval)
		}
	}()
	return e
}
