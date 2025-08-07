package api

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	appError "github.com/fireops-software/fireops-edge-printer/error"
	"github.com/signintech/gopdf"
)

const (
	PRINT_CMD = "/usr/bin/lpr -o portrait -o fit-to-page -o media=A4 -P %s -#%d %s"
)

type PrinterApi struct {
	printerName string
}

// IsOnline implements IPrinterApi.
func (p *PrinterApi) IsOnline(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "/usr/bin/lpstat", "-p", p.printerName)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	state := string(out)
	return strings.Contains(state, "idle") || strings.Contains(state, "printing")

}

// PrintPdf implements IPrinterApi.
func (p *PrinterApi) PrintPdf(ctx context.Context, pdf *gopdf.GoPdf, copies int) error {
	// Get bytes of pdf document
	doc := pdf.GetBytesPdf()
	// Create temp file for printing
	f, err := os.CreateTemp("", "job_*.pdf")
	if err != nil {
		return appError.NewErrPrinter("failed to create pdf file on filesystem for printing - %v", err)
	}
	// Cleanup
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	// Write pdf data to file
	if _, err := f.Write(doc); err != nil {
		return appError.NewErrPrinter("failed to write pdf data to tempfile - %v", err)
	}
	// Close file after pdf content has been written
	f.Close()
	// Create print command
	printCmd := exec.CommandContext(ctx, "/usr/bin/lpr", "-o", "portrait", "-o", "fit-to-page", "-o", "media=A4", "-P", p.printerName, fmt.Sprintf("-#%d", copies), f.Name())
	// Execute print command
	if stdOut, err := printCmd.Output(); err != nil {
		return appError.NewErrPrinter("%s - %v", string(stdOut), err)
	}
	return nil
}

func NewPrinterApi(printerName string, opts ...func(*PrinterApi)) IPrinterApi {
	p := &PrinterApi{
		printerName: printerName,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}
