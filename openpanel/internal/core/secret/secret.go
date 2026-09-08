// Package secret loads the panel's signing key from secret.key, or generates an ephemeral one for this process if the file doesn't exist yet.
package secret

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

const DefaultPath = "/etc/openpanel/openpanel/secret.key"

// Load returns the panel's secret key for session/CSRF signing. If path doesn't exist, it generates a random one for this process only, so sessions won't survive a restart until the file exists.
func Load(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return []byte(strings.TrimSpace(string(data))), nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	fmt.Printf("WARNING - %s not found, generating a random ephemeral secret key for this process. "+
		"Sessions will not survive a restart until this file is created.\n", path)

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return []byte(hex.EncodeToString(buf)), nil
}
