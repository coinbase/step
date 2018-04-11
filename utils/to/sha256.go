package to

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
)

func SHA256Struct(str interface{}) string {
	raw, err := json.Marshal(str)
	if err != nil {
		// No deterministic error
		return RandomString(10)
	}

	return SHA256AByte(&raw)
}

// SHA256Str returns a hex string of the SHA256 of a string
func SHA256Str(str *string) string {
	byt := []byte(*str)
	return SHA256AByte(&byt)
}

// SHA256AByte returns a hex string of the SHA256 of a byte array
func SHA256AByte(b *[]byte) string {
	sum := sha256.Sum256(*b)
	sha := hex.EncodeToString(sum[:])
	return sha
}

// SHA256File returns a hex string of the SHA256 of a file
func SHA256File(file_path string) (string, error) {
	f, err := os.Open(file_path)
	if err != nil {
		// No deterministic error
		return RandomString(10), err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		// No deterministic error
		return RandomString(10), err
	}
	sha := hex.EncodeToString(hasher.Sum(nil))
	return sha, nil
}
