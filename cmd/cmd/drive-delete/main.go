package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// helper for fatal errors
func fatalf(format string, a ...any) { log.Fatalf(format, a...) }

func main() {
	if len(os.Args) != 3 {
		fatalf("usage: drive-delete <credentials.json> <fileID>")
	}
	creds, fileID := os.Args[1], os.Args[2]

	// 1 — Google Drive client
	ctx := context.Background()
	svc, err := drive.NewService(ctx, option.WithCredentialsFile(creds))
	if err != nil {
		fatalf("drive service: %v", err)
	}

	// 2 — fetch metadata before deletion
	f, err := svc.Files.Get(fileID).
		Fields("id", "name", "size", "md5Checksum").Do()
	if err != nil {
		fatalf("get meta: %v", err)
	}
	meta := fmt.Sprintf("%s:%s:%d:%s", f.Id, f.Name, f.Size, f.Md5Checksum)
	hash := sha256.Sum256([]byte(meta))
	hashHex := hex.EncodeToString(hash[:])

	// 3 — hard delete
	_, err = svc.Files.Delete(fileID).SupportsAllDrives(true).Do()
	if err != nil {
		fatalf("delete: %v", err)
	}

	// 4 — verify removal (expect error 404)
	_, err = svc.Files.Get(fileID).Do()
	if err == nil {
		fatalf("verify failed – file still present")
	}

	// 5 — RFC 3161 timestamp via FreeTSA
	tsr, err := timestamp(hashHex)
	if err != nil {
		fatalf("tsa: %v", err)
	}

	// 6 — write a minimal receipt text file
	receipt := "receipt.txt"
	fh, _ := os.Create(receipt)
	defer fh.Close()
	io.WriteString(fh, "DeletionOps – Proof-of-Deletion\n")
	io.WriteString(fh, "File ID: "+f.Id+"\n")
	io.WriteString(fh, "Pre-delete SHA256: "+hashHex+"\n")
	io.WriteString(fh, fmt.Sprintf("TSA response bytes: %d\n", len(tsr)))
	fmt.Println("✔︎  receipt saved as", receipt)
}
