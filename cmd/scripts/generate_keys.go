package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

func privateKeyToPEM(privateKey *rsa.PrivateKey) string {
	// Преобразуем приватный ключ в DER-формат
	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// Создаем PEM-блок
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	}

	// Преобразуем в строку
	return string(pem.EncodeToMemory(privateKeyPEM))
}

func publicKeyToPEM(publicKey *rsa.PublicKey) string {
	// Преобразуем публичный ключ в DER-формат
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		log.Fatalf("Failed to marshal public key: %v", err)
	}

	// Создаем PEM-блок
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	}

	// Преобразуем в строку
	return string(pem.EncodeToMemory(publicKeyPEM))
}

func saveKeyToFile(keyPEM string, keyType string) (err error) {
	var file *os.File
	if keyType == "public" {
		file, err = os.Create("/tmp/public")
		if err != nil {
			return
		}
	} else {
		file, err = os.Create("/tmp/private")
		if err != nil {
			return
		}
	}
	_, err = file.Write([]byte(keyPEM))
	return
}

func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Println("Error while generating key:", err)
	}

	publicKey := &privateKey.PublicKey

	err = saveKeyToFile(privateKeyToPEM(privateKey), "private")
	if err != nil {
		fmt.Println("Error while converting private key to string:", err)
	}
	err = saveKeyToFile(publicKeyToPEM(publicKey), "public")
	if err != nil {
		fmt.Println("Error while converting public key to string:", err)
	}

}
