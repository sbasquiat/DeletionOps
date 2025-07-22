package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os/exec"
)

// timestamp talks to FreeTSA.org and returns the raw .tsr bytes
func timestamp(payload string) ([]byte, error) {
	sum, _ := hex.DecodeString(payload)

	// build a timestamp-query using openssl
	cmd := exec.Command("openssl", "ts",
		"-query", "-digest", hex.EncodeToString(sum), "-sha256", "-no_nonce")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// POST to the public TSA endpoint
	resp, err := http.Post("https://freetsa.org/tsr",
		"application/timestamp-query", &out)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
