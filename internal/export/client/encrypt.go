/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

// Encrypt string to base64 crypto using AES
func Encrypt(text string) (cryptoText string, err error) {
	pass, err := getShadow()
	if err != nil {
		return
	}
	// key := []byte(keyText)
	plaintext := []byte(text)

	block, err := aes.NewCipher([]byte(pass))
	if err != nil {
		return
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	cryptoText = base64.URLEncoding.EncodeToString(ciphertext)
	return
}

// Decrypt from base64 to decrypted string
func Decrypt(cryptoText string) (text string, err error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	pass, err := getShadow()
	if err != nil {
		return
	}

	block, err := aes.NewCipher([]byte(pass))
	if err != nil {
		return
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		err = errors.New("ciphertext too short")
		return
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	text = fmt.Sprintf("%s", ciphertext)
	return
}

func getShadow() (string, error) {
	_, err := os.Stat("/etc/shadow.toml")
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("Shadow file not found: %v", err)
		}
	}

	contents, err := ioutil.ReadFile("/etc/shadow.toml")
	if err != nil {
		return "", fmt.Errorf("Failed to read shadow file: %v", err)
	}
	var psk struct {
		Shadow string
	}
	err = toml.Unmarshal(contents, &psk)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal shadow file: %v", err)
	}

	return psk.Shadow, nil
}
