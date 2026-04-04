package service

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"netease-music-rag/backend/internal/config"
)

type NeteaseOpenAPIClient struct {
	AppId      string
	AppKey     string // RSA Private Key String
	HttpClient *http.Client
}

func NewNeteaseOpenAPIClient(cfg *config.Config) *NeteaseOpenAPIClient {
	return &NeteaseOpenAPIClient{
		AppId:  cfg.NeteaseAppId,
		AppKey: cfg.NeteaseAppKey,
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GenerateSignature sorts parameters alphabetically, hashes with SHA256, and signs with RSA Private Key
func (c *NeteaseOpenAPIClient) GenerateSignature(params map[string]string) (string, error) {
	// 1. Sort the keys alphabetically
	var keys []string
	for k := range params {
		if k != "sign" { // omit sign itself if present
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 2. Build the parameter string: k1=v1&k2=v2...
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	rawStr := sb.String()

	// 3. Parse private key
	// Ensure the AppKey is wrapped in standard PEM boundaries if missing
	privKeyStr := c.AppKey
	if !strings.Contains(privKeyStr, "BEGIN PRIVATE KEY") && !strings.Contains(privKeyStr, "BEGIN RSA PRIVATE KEY") {
		privKeyStr = "-----BEGIN PRIVATE KEY-----\n" + c.AppKey + "\n-----END PRIVATE KEY-----"
	}

	block, _ := pem.Decode([]byte(privKeyStr))
	if block == nil {
		return "", errors.New("failed to parse PEM block from AppKey")
	}

	var rsaPriv *rsa.PrivateKey
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// fallback to PKCS1
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %v", err)
		}
		rsaPriv = priv.(*rsa.PrivateKey)
	} else {
		var ok bool
		rsaPriv, ok = priv.(*rsa.PrivateKey)
		if !ok {
			return "", errors.New("parsed key is not an RSA private key")
		}
	}

	// 4. Hash and Sign (RSA_SHA256)
	hashed := sha256.Sum256([]byte(rawStr))
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPriv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign: %v", err)
	}

	// 5. Return Base64 encoded signature
	return base64.StdEncoding.EncodeToString(signature), nil
}

// GetRecommendSonglist gets the recommended playlist via the official OpenAPI
func (c *NeteaseOpenAPIClient) GetRecommendSonglist(accessToken string, limit int) (string, error) {
	apiEndpoint := "http://openapi.music.163.com/openapi/music/basic/recommend/songlist/get/v2"

	bizContentBytes, _ := json.Marshal(map[string]interface{}{
		"limit": limit,
	})

	deviceBytes, _ := json.Marshal(map[string]interface{}{
		"deviceType": "andrwear",
		"os":         "andrwear",
		"appVer":     "0.1",
		"channel":    "hm",
		"model":      "kys",
		"deviceId":   "321",
		"brand":      "hm",
		"osVer":      "8.1.0",
	})

	// Prepare the raw parameters
	params := map[string]string{
		"appId":       c.AppId,
		"bizContent":  string(bizContentBytes),
		"signType":    "RSA_SHA256",
		"accessToken": accessToken,
		"device":      string(deviceBytes),
		"timestamp":   strconv.FormatInt(time.Now().UnixNano()/1e6, 10), // current ms
	}

	// Generate the RSA signature
	signature, err := c.GenerateSignature(params)
	if err != nil {
		return "", fmt.Errorf("signature error: %v", err)
	}

	// Construct final query string with url encoding
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("sign", signature)

	fullURL := fmt.Sprintf("%s?%s", apiEndpoint, q.Encode())

	// Execute HTTP GET request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
