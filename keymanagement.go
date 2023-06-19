package pineweb

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func GetKeysFromKeyfile(keyfile string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	_, err := os.Stat(keyfile)
	if os.IsNotExist(err) {
		err := newKey(keyfile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate a new PEM key: " + err.Error())
		}
	}

	_, sk, err := loadKey(keyfile, os.ReadFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load PEM key: " + err.Error())
	}
	if len(sk) != ed25519.PrivateKeySize {
		return nil, nil, fmt.Errorf("the private key is not long enough")
	}

	return sk.Public().(ed25519.PublicKey), sk, nil
}

func newKey(keypath string) error {
	var data [35]byte
	_, err := rand.Read(data[:])
	if err != nil {
		return err
	}
	return saveKey(keypath, data[3:])
}

func saveKey(keypath string, data ed25519.PrivateKey) error {
	keyOut, err := os.OpenFile(keypath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	keyID := base64.RawURLEncoding.EncodeToString(data[:])
	keyID = strings.ReplaceAll(keyID, "-", "")
	keyID = strings.ReplaceAll(keyID, "_", "")

	err = pem.Encode(keyOut, &pem.Block{
		Type: "PINECONE BROWSER PRIVATE KEY",
		Headers: map[string]string{
			"Key-ID": fmt.Sprintf("ed25519:%s", keyID[:6]),
		},
		Bytes: data,
	})

	err = keyOut.Close()
	return err
}

func loadKey(privateKeyPath string, readFile func(string) ([]byte, error)) (string, ed25519.PrivateKey, error) {
	privateKeyData, err := readFile(privateKeyPath)
	if err != nil {
		return "", nil, err
	}
	return readKeyPEM(privateKeyPath, privateKeyData, true)
}

func readKeyPEM(path string, data []byte, enforceKeyIDFormat bool) (string, ed25519.PrivateKey, error) {
	var keyIDRegexp = regexp.MustCompile("^ed25519:[a-zA-Z0-9_]+$")
	for {
		var keyBlock *pem.Block
		keyBlock, data = pem.Decode(data)
		if data == nil {
			return "", nil, fmt.Errorf("no private key PEM data in %q", path)
		}
		if keyBlock == nil {
			return "", nil, fmt.Errorf("keyBlock is nil %q", path)
		}
		if keyBlock.Type == "PINECONE BROWSER PRIVATE KEY" {
			keyID := keyBlock.Headers["Key-ID"]
			if keyID == "" {
				return "", nil, fmt.Errorf("missing key ID in PEM data in %q", path)
			}
			if !strings.HasPrefix(keyID, "ed25519:") {
				return "", nil, fmt.Errorf("key ID %q doesn't start with \"ed25519:\" in %q", keyID, path)
			}
			if enforceKeyIDFormat && !keyIDRegexp.MatchString(keyID) {
				return "", nil, fmt.Errorf("key ID %q in %q contains illegal characters (use a-z, A-Z, 0-9 and _ only)", keyID, path)
			}
			_, privKey, err := ed25519.GenerateKey(bytes.NewReader(keyBlock.Bytes))
			if err != nil {
				return "", nil, err
			}
			return keyID, privKey, nil
		}
	}
}
