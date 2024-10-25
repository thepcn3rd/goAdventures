package main

/*
Generates server.crt and server.key for a self-signed certificate for servers that need them

Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/createCerts"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o createCerts.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o createCerts.exe -ldflags "-w -s" main.go

References:
Generated the base of the code using ChatGPT 10/7/2023
URL: https://github.com/kgretzky/evilginx2/blob/master/core/certdb.go - Learned about the x509 template and extended it

Example JSON config file used to generate new certificates
{
        "DNSNames": [
                "blog.example.com",
                "www.example.com"
        ],
        "Org": "Example Inc",
        "OrgUnit": "",
        "CommonName": "example.com",
        "City": "Lewiston",
        "State": "ID",
        "Country": "US",
        "Email": "admin@example.com"
}

*/

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"
)

type certConfig struct {
	DNSNames   []string `json:"DNSNames"`
	Org        string   `json:"Org"`
	OrgUnit    string   `json:"OrgUnit"`
	CommonName string   `json:"CommonName"`
	City       string   `json:"City"`
	State      string   `json:"State"`
	Country    string   `json:"Country"`
	Email      string   `json:"Email"`
}

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

func createDirectory(createDir string) {
	currentDir, err := os.Getwd()
	checkError("Unable to get the working directory", err)
	newDir := currentDir + createDir
	if _, err := os.Stat(newDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(newDir, os.ModePerm)
		checkError("Unable to create directory "+createDir, err)
	}
}

func main() {
	ConfigPtr := flag.String("config", "certConfig.json", "Location of the configuration file for certificate generation")
	flag.Parse()

	// Read the configuration file
	configFile, err := os.ReadFile(*ConfigPtr)
	checkError("Failed to read configuration file", err)

	// Parse the configuration file
	var certConfig certConfig
	err = json.Unmarshal(configFile, &certConfig)
	checkError("Failed to parse configuration file", err)

	fmt.Println("Generating certificate and private key...")

	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	checkError("Failed to generate private key", err)

	// Generate values for the below options in the x509 Cert
	// Evaluated how evilginx2 did it in certdb.go
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Printf("Failed to generate serial number: %v\n", err)
		return
	}

	// Create a certificate template with DNSNames (FQDN)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{certConfig.Org},
			OrganizationalUnit: []string{certConfig.OrgUnit},
			CommonName:         certConfig.CommonName,
			Locality:           []string{certConfig.City},    // City
			Province:           []string{certConfig.State},   // State
			Country:            []string{certConfig.Country}, // Country,
		},
		NotBefore:             time.Now().Add(-37 * 24 * time.Hour),
		NotAfter:              time.Now().Add((365 - 37) * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		DNSNames:              []string{certConfig.DNSNames[0], certConfig.DNSNames[1]},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		EmailAddresses:        []string{certConfig.Email},
	}

	// Create a self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	checkError("Failed to create certificate: %v\n", err)

	// Create the directory of keys in the current directory
	createDirectory("/keys")

	// Save the private key to a file
	privateKeyFile, err := os.Create("keys/server.key")
	checkError("Failed to create private key file: %v\n", err)
	defer privateKeyFile.Close()

	// Save the private key to a file
	privKeyBytes, _ := x509.MarshalECPrivateKey(privateKey)
	pem.Encode(privateKeyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privKeyBytes})

	// Save the certificate to a file
	certificateFile, err := os.Create("keys/server.crt")
	checkError("Failed to create certificate file: %v\n", err)
	defer certificateFile.Close()
	pem.Encode(certificateFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	fmt.Println("Saved the certificate and key in the keys directory")
	fmt.Println("\nCertificate and private key generated successfully!")
}
