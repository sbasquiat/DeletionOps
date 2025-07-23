package main

import (
	"fmt"
	"os"
)

// writePDF is a placeholder that writes a plain‑text “PDF”.
// Replace later with real gofpdf logic.
func writePDF(fileID, hash string, tsaBytes int) error {
	name := fmt.Sprintf("receipt_%s.pdf", fileID)
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f,
		"DeletionOps – Proof‑of‑Deletion\n"+
			"File ID: %s\n"+
			"Pre‑delete SHA‑256: %s\n"+
			"TSA response bytes: %d\n",
		fileID, hash, tsaBytes)
	return err
}
