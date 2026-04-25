package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"netease-music-rag/backend/internal/config"
	"netease-music-rag/backend/internal/model"
)

type LoginAuth struct {
	cookie        string
	qrCodeKey     string
	qrCodeMessage string
	qrCodeImgStr  string
	qrCodeStatus  int
}

type NeteaseClient struct {
	baseURL    string
	httpClient *http.Client
	auth       *LoginAuth
}

func NewNeteaseClient(cfg *config.Config) *NeteaseClient {

	jar, _ := cookiejar.New(nil)

	client := &NeteaseClient{
		baseURL: cfg.NeteaseAPIURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
		auth: &LoginAuth{},
	}
	return client
}

func (c *NeteaseClient) generateQrKey() (string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/login/qr/key", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to generate QR key, status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Unikey string `json:"unikey"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	c.auth.qrCodeKey = result.Data.Unikey
	return result.Data.Unikey, nil
}

func (c *NeteaseClient) createQrCode() (string, error) {
	params := url.Values{}
	params.Set("key", c.auth.qrCodeKey)
	params.Set("qrimg", "true")

	req, err := http.NewRequest("GET", c.baseURL+"/login/qr/create?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create QR code, status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Qrimg string `json:"qrimg"` // base64 PNG data URL, returned when qrimg=true
			Qrurl string `json:"qrurl"` // login URL (kept for reference)
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Data.Qrimg == "" {
		return "", fmt.Errorf("qrimg field missing from API response")
	}

	c.auth.qrCodeImgStr = result.Data.Qrimg
	return result.Data.Qrimg, nil
}

// GenerateLoginQR creates a fresh QR key + code and returns (base64PngDataURL, key, error).
func (c *NeteaseClient) GenerateLoginQR() (string, string, error) {
	key, err := c.generateQrKey()
	if err != nil {
		return "", "", fmt.Errorf("generateQrKey: %w", err)
	}
	imgDataURL, err := c.createQrCode()
	if err != nil {
		return "", "", fmt.Errorf("createQrCode: %w", err)
	}
	return imgDataURL, key, nil
}

func (c *NeteaseClient) checkLoginStatus() error {
	params := url.Values{}
	params.Set("key", c.auth.qrCodeKey)
	// Add timestamp to prevent caching
	params.Set("t", fmt.Sprintf("%d", time.Now().UnixNano()/1e6))

	req, err := http.NewRequest("GET", c.baseURL+"/login/qr/check?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Cookie  string `json:"cookie"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	c.auth.qrCodeStatus = result.Code
	c.auth.qrCodeMessage = result.Message
	if result.Cookie != "" {
		c.auth.cookie = result.Cookie
	}

	return nil
}

// CheckLoginStatus polls the QR scan status for a given key.
// Returns (statusCode, message, error). Status codes: 800=expired, 801=waiting, 802=scanned, 803=success.
func (c *NeteaseClient) CheckLoginStatus(key string) (int, string, error) {
	prev := c.auth.qrCodeKey
	c.auth.qrCodeKey = key
	err := c.checkLoginStatus()
	c.auth.qrCodeKey = prev
	if err != nil {
		return 0, "", err
	}
	return c.auth.qrCodeStatus, c.auth.qrCodeMessage, nil
}

// 二维码登陆
func (c *NeteaseClient) Login(phone string) error {
	// 1. generate qr key
	_, err := c.generateQrKey()
	if err != nil {
		return fmt.Errorf("failed to generate QR key: %w", err)
	}

	// 2. create qr code
	_, err = c.createQrCode()
	if err != nil {
		return fmt.Errorf("failed to create QR code: %w", err)
	}

	ShowQRCodeInTerminal(c.auth.qrCodeImgStr)

	// 3. iteratively check login status
	success := Poll(5, 120, func() bool {
		err := c.checkLoginStatus()
		if err != nil {
			log.Printf("Check login status error: %v", err)
			return false
		}
		log.Printf("QR Status: %d, Message: %s", c.auth.qrCodeStatus, c.auth.qrCodeMessage)
		// 800: expired, 801: waiting, 802: scanning, 803: success
		return c.auth.qrCodeStatus == 803
	})

	if !success {
		return errors.New("login timeout or failed")
	}

	log.Printf("Login successful!")
	return nil
}

// 获取每日推荐歌单
func (c *NeteaseClient) GetDailyRecommendPlaylist() (*[]model.RecommendPlaylistData, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/recommend/resource", nil)
	if err != nil {
		return nil, err
	}

	if c.auth.cookie != "" {
		// Some endpoints prefer cookie in header, some in params.
		// NeteaseCloudMusicApi usually handles both but header is standard.
		req.Header.Set("Cookie", c.auth.cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Code      int                           `json:"code"`
		Recommend []model.RecommendPlaylistData `json:"recommend"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	return &result.Recommend, nil
}

func (c *NeteaseClient) GetDetailPlaylist(playlistId int64) (*model.DetailPlaylistData, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/playlist/detail?id="+fmt.Sprintf("%d", playlistId), nil)
	if err != nil {
		return nil, err
	}

	if c.auth.cookie != "" {
		// Some endpoints prefer cookie in header, some in params.
		// NeteaseCloudMusicApi usually handles both but header is standard.
		req.Header.Set("Cookie", c.auth.cookie);
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Code     int                      `json:"code"`
		Playlist model.DetailPlaylistData `json:"playlist"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	return &result.Playlist, nil
}

func (c *NeteaseClient) GetSongLyrics(songId int64) (*string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/lyric?id="+fmt.Sprintf("%d", songId), nil)
	if err != nil {
		return nil, err
	}

	if c.auth.cookie != "" {
		// Some endpoints prefer cookie in header, some in params.
		// NeteaseCloudMusicApi usually handles both but header is standard.
		req.Header.Set("Cookie", c.auth.cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Code int `json:"code"`
		LRC  struct {
			Lyric string `json:"lyric"`
		} `json:"lrc"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	return &result.LRC.Lyric, nil
}

func (c *NeteaseClient) GetDailyRecommendations() ([]model.NeteaseSongDTO, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/recommend/songs", nil)
	if err != nil {
		return nil, err
	}

	if c.auth.cookie != "" {
		// Some endpoints prefer cookie in header, some in params.
		// NeteaseCloudMusicApi usually handles both but header is standard.
		req.Header.Set("Cookie", c.auth.cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			DailySongs []model.NeteaseSongDTO `json:"dailySongs"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	return result.Data.DailySongs, nil
}
