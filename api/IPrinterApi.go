package api

import (
	"context"

	"github.com/signintech/gopdf"
)

type IPrinterApi interface {
	PrintPdf(ctx context.Context, pdf *gopdf.GoPdf, copies int) error
}
