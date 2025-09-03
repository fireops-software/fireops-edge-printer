package api

import (
	"context"

	"github.com/fireops-software/fireops-edge-printer/domain"
)

type IPrinterApi interface {
	IsOnline(ctx context.Context) bool
	PrintEvents(ctx context.Context, events []domain.Event, copies int) error
}
