package cert

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1" //nolint:gosec
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/local-deploy/dl/helper"
	"github.com/pterm/pterm"
)

// CaRootName certificate file name
const CaRootName = "rootCA.pem"

// CaRootKeyName certificate key file name
const CaRootKeyName = "rootCA-key.pem"

// Cert certificate structure
type Cert struct {
	CertutilPath  string
	CaFileName    string
	CaFileKeyName string
	CaPath        string
	CaCert        *x509.Certificate
	CaKey         crypto.PrivateKey

	keyFile, certFile string
}

// LoadCA certificate reading
func (c *Cert) LoadCA() error {
	if !pathExists(filepath.Join(c.CaPath, c.CaFileName)) {
		return nil
	}

	certPEMBlock, err := os.ReadFile(filepath.Join(c.CaPath, c.CaFileName))
	if err != nil {
		return fmt.Errorf("failed to read the CA certificate: %w", err)
	}
	certDERBlock, _ := pem.Decode(certPEMBlock)
	if certDERBlock == nil || certDERBlock.Type != "CERTIFICATE" {
		return errors.New("failed to read the CA certificate: unexpected content")
	}
	c.CaCert, err = x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse the CA certificate: %w", err)
	}

	if !pathExists(filepath.Join(c.CaPath, c.CaFileKeyName)) {
		return nil
	}

	keyPEMBlock, err := os.ReadFile(filepath.Join(c.CaPath, c.CaFileKeyName))
	if err != nil {
		return fmt.Errorf("failed to read the CA key: %w", err)
	}
	keyDERBlock, _ := pem.Decode(keyPEMBlock)
	if keyDERBlock == nil || keyDERBlock.Type != "PRIVATE KEY" {
		return errors.New("failed to read the CA key: unexpected content")
	}
	c.CaKey, err = x509.ParsePKCS8PrivateKey(keyDERBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse the CA key: %w", err)
	}
	return nil
}

// CreateCA creating a root certificate
func (c *Cert) CreateCA() error {
	privateKey, err := c.generateKey(true)
	if err != nil {
		return fmt.Errorf("failed to generate the CA key: %w", err)
	}
	publicKey := privateKey.(crypto.Signer).Public()

	pkixPublicKey, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	var keyIdentifier struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	_, err = asn1.Unmarshal(pkixPublicKey, &keyIdentifier)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	checksum := sha1.Sum(keyIdentifier.SubjectPublicKey.Bytes) //nolint:gosec

	template := &x509.Certificate{
		SerialNumber: randomSerialNumber(),
		Subject: pkix.Name{
			Organization:       []string{"Local Deploy CA"},
			OrganizationalUnit: []string{"DL Certificate Authority"},
			CommonName:         "DL Certificate",
		},
		SubjectKeyId:          checksum[:],
		NotAfter:              time.Now().AddDate(10, 0, 0),
		NotBefore:             time.Now(),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}

	certificate, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to generate CA certificate: %w", err)
	}

	pkcs8PrivateKey, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to encode CA key: %w", err)
	}
	err = os.WriteFile(filepath.Join(c.CaPath, c.CaFileKeyName), pem.EncodeToMemory(
		&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8PrivateKey}), 0400)
	if err != nil {
		return fmt.Errorf("failed to save CA key: %w", err)
	}

	err = os.WriteFile(filepath.Join(c.CaPath, c.CaFileName), pem.EncodeToMemory( //nolint:gosec
		&pem.Block{Type: "CERTIFICATE", Bytes: certificate}), 0644)
	if err != nil {
		return fmt.Errorf("failed to save CA certificate: %w", err)
	}
	return nil
}

// MakeCert Create certificates for domains
func (c *Cert) MakeCert(hosts []string, path string) error {
	if c.CaKey == nil {
		return fmt.Errorf("can't create new certificates because the CA key (%s) is missing", c.CaFileKeyName)
	}

	privateKey, err := c.generateKey(false)
	if err != nil {
		return fmt.Errorf("failed to generate certificate key: %w", err)
	}
	publicKey := privateKey.(crypto.Signer).Public()
	expiration := time.Now().AddDate(2, 3, 0)

	dir, _ := os.Getwd()
	template := &x509.Certificate{
		SerialNumber: randomSerialNumber(),
		Subject: pkix.Name{
			Organization:       []string{filepath.Base(dir) + " development certificate"},
			OrganizationalUnit: []string{filepath.Base(dir) + " Certificate"},
		},

		NotBefore: time.Now(), NotAfter: expiration,
		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else if uriName, err := url.Parse(h); err == nil && uriName.Scheme != "" && uriName.Host != "" {
			template.URIs = append(template.URIs, uriName)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if len(template.IPAddresses) > 0 || len(template.DNSNames) > 0 || len(template.URIs) > 0 {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, c.CaCert, publicKey, c.CaKey)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	certFile, keyFile := c.fileNames()

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	pkcs8PrivateKey, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to encode certificate key: %w", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8PrivateKey})

	if certFile == keyFile {
		err = os.WriteFile(filepath.Join(helper.CertDir(), path, keyFile), append(certPEM, privatePEM...), 0600)
		if err != nil {
			return fmt.Errorf("failed to save certificate and key: %w", err)
		}
	} else {
		err = os.WriteFile(filepath.Join(helper.CertDir(), path, certFile), certPEM, 0644) //nolint:gosec
		if err != nil {
			return fmt.Errorf("failed to save certificate: %w", err)
		}
		err = os.WriteFile(filepath.Join(helper.CertDir(), path, keyFile), privatePEM, 0600)
		if err != nil {
			return fmt.Errorf("failed to save certificate key: %w", err)
		}
	}

	// TODO: add debug

	// if certFile == keyFile {
	// 	log.Printf("\nThe certificate and key are at \"%s\"\n\n", certFile)
	// } else {
	// 	log.Printf("\nThe certificate is at \"%s\" and the key at \"%s\"\n\n", certFile, keyFile)
	// }
	//
	// log.Printf("It will expire on %s\n\n", expiration.Format("2 January 2006"))
	return nil
}

func (c *Cert) fileNames() (certFile, keyFile string) {
	certFile = "./cert.pem"
	if c.certFile != "" {
		certFile = c.certFile
	}
	keyFile = "./key.pem"
	if c.keyFile != "" {
		keyFile = c.keyFile
	}

	return
}

func (c *Cert) generateKey(rootCA bool) (crypto.PrivateKey, error) {
	if rootCA {
		return rsa.GenerateKey(rand.Reader, 3072)
	}
	return rsa.GenerateKey(rand.Reader, 2048)
}

func (c *Cert) verifyCert() bool {
	_, err := c.CaCert.Verify(x509.VerifyOptions{})
	return err == nil
}

func (c *Cert) caUniqueName() string {
	return "DL development CA " + c.CaCert.SerialNumber.String()
}

func randomSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		pterm.FgRed.Printfln("failed to generate serial number: %s", err)
		os.Exit(1)
	}
	return serialNumber
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
