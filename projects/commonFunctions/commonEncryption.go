package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
)

// EncryptString encrypts a string using AES-256 encryption with a random IV.
func EncryptString(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	//fmt.Println(iv)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a string using AES-256 encryption.
func DecryptString(key []byte, ciphertext string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(data) < aes.BlockSize {
		return "", errors.New("Ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return string(data), nil
}

func PullKey(keyURL string, userAgentString string, xAbility string) string {
	url := keyURL
	req, err := http.NewRequest("POST", url, nil)
	CheckError("Unable to pull key from URL...", err, true)
	req.Header.Set("User-Agent", userAgentString)
	req.Header.Set("X-Content-Type-Abilities", xAbility)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	CheckError("Unable to get response from Amazon...", err, true)
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	respBodyString := string(respBody)
	//fmt.Println(string(respBody))
	return respBodyString
}
