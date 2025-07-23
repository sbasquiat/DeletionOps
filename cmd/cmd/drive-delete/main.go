package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

)

func fatalf(format string, a ...any) { log.Fatalf(format, a...) }

func main() {
	if len(os.Args) != 3 {
		fatalf("usage: drive-delete <credentials.json> <fileID>")
	}
	credsFile, fileID := os.Args[1], os.Args[2]

	// 1 ── OAuth config from client_secret_….json
	secret, err := os.ReadFile(credsFile)
	if err != nil { fatalf("read creds: %v", err) }
	conf, err := google.ConfigFromJSON(secret, drive.DriveScope)
	if err != nil { fatalf("parse creds: %v", err) }

	// 2 ── interactive token (one-time)
	tok, err := tokenFromWeb(conf)
	if err != nil { fatalf("token: %v", err) }

	// 3 ── Drive service using the authorised *http.Client*
	httpClient := oauth2.NewClient(context.Background(), conf.TokenSource(context.Background(), tok))
	svc, err := drive.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil { fatalf("drive service: %v", err) }

	// 4 ── fetch metadata, compute SHA-256
	f, err := svc.Files.Get(fileID).Fields("id", "name", "size", "md5Checksum").Do()
	if err != nil { fatalf("get meta: %v", err) }
	meta := fmt.Sprintf("%s:%s:%d:%s", f.Id, f.Name, f.Size, f.Md5Checksum)
	hash := sha256.Sum256([]byte(meta))   // hash is a [32]byte array
    hashHex := hex.EncodeToString(hash[:]) // convert to slice, then hex


	// 5 ── delete & verify 404
	if err = svc.Files.Delete(fileID).SupportsAllDrives(true).Do(); err != nil {
		fatalf("delete: %v", err)
	}
	if _, err = svc.Files.Get(fileID).Do(); err == nil {
		fatalf("verify failed – file still present")
	}

	// 6 ── RFC-3161 timestamp
	tsr, err := timestamp(hashHex)
	if err != nil { fatalf("tsa: %v", err) }

	// 7 ── tiny text receipt
	if err := writePDF(f.Id, hashHex, len(tsr)); err != nil {
    fatalf("pdf: %v", err)
}
fmt.Println("✔︎  receipt saved as receipt.pdf")
}

/* ---------- helpers ---------- */

func tokenFromWeb(conf *oauth2.Config) (*oauth2.Token, error) {
	url := conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Open this URL, authorise, then paste the code:\n%v\n> ", url)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}
	return conf.Exchange(context.Background(), code)
}

func timestamp(hexDigest string) ([]byte, error) {
	// build a TSQ with openssl
	cmd := exec.Command("openssl", "ts",
		"-query", "-digest", hexDigest, "-sha256", "-no_nonce")
	var tsq bytes.Buffer
	cmd.Stdout = &tsq
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// POST to FreeTSA
	resp, err := http.Post(
		"https://freetsa.org/tsr",
		"application/timestamp-query",
		bytes.NewReader(tsq.Bytes()),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}