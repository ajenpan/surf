package rsagen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func GenerateRsaPem(bits int) ([]byte, []byte, error) {
	pk, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	priv_pem := ExportRsaPrivateKeyAsPem(pk)
	pub_pem, err := ExportRsaPublicKeyAsPem(&pk.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	return priv_pem, pub_pem, nil
}

func GenerateRsaKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, _ := rsa.GenerateKey(rand.Reader, bits)
	return privkey, &privkey.PublicKey
}

func ExportRsaPrivateKeyAsPem(privkey *rsa.PrivateKey) []byte {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return privkey_pem
}

func ExportRsaPublicKeyAsPem(pubkey *rsa.PublicKey) ([]byte, error) {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return nil, err
	}
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)
	return pubkey_pem, nil
}

func ParseRsaPrivateKeyFromPem(privPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func ParseRsaPublicKeyFromPem(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break // fall through
	}
	return nil, errors.New("key type is not RSA")
}

func LoadRsaPublicKeyFromFile(fname string) (*rsa.PublicKey, error) {
	publicRaw, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	return ParseRsaPublicKeyFromPem(publicRaw)
}

func LoadRsaPrivateKeyFromFile(fname string) (*rsa.PrivateKey, error) {
	privateRaw, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	return ParseRsaPrivateKeyFromPem(privateRaw)
}

func ReadFromUrl(path string) ([]byte, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	var raw []byte
	if u.Scheme == "http" || u.Scheme == "https" {
		resp, nerr := http.Get(u.String())
		if nerr != nil {
			return nil, nerr
		}
		defer resp.Body.Close()
		raw, err = io.ReadAll(resp.Body)
	} else if u.Scheme == "file" {
		raw, err = os.ReadFile(u.Host + u.Path)
	} else {
		err = fmt.Errorf("unsupported scheme: %v", u.Scheme)
	}
	return raw, err
}

func LoadRsaPublicKeyFromUrl(path string) (*rsa.PublicKey, error) {
	raw, err := ReadFromUrl(path)
	if err != nil {
		return nil, err
	}
	return ParseRsaPublicKeyFromPem(raw)
}

func LoadRsaPrivateKeyFromUrl(path string) (*rsa.PrivateKey, error) {
	raw, err := ReadFromUrl(path)
	if err != nil {
		return nil, err
	}
	return ParseRsaPrivateKeyFromPem(raw)
}
