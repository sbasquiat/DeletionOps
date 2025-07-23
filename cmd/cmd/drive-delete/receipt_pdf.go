package main

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"

)

func writePDF(id, hash string, tsrBytes int) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Proof‑of‑Deletion", false)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(0, 10, "DeletionOps – Proof‑of‑Deletion")
	pdf.Ln(12)

	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(50, 8, "File ID:")
	pdf.Cell(0, 8, id)
	pdf.Ln(8)

	pdf.Cell(50, 8, "Pre‑delete SHA‑256:")
	pdf.Cell(0, 8, hash)
	pdf.Ln(8)

	pdf.Cell(50, 8, "TSA bytes:")
	pdf.Cell(0, 8, fmt.Sprint(tsrBytes))
	pdf.Ln(8)

	pdf.Cell(50, 8, "Generated:")
	pdf.Cell(0, 8, time.Now().Format(time.RFC3339))
	pdf.Ln(12)

	return pdf.OutputFileAndClose("receipt.pdf")
}
