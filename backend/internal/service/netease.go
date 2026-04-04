package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"netease-music-rag/backend/internal/config"
	"netease-music-rag/backend/internal/model"
)

type NeteaseClient struct {
	baseURL    string
	httpClient *http.Client
	cookie     string
}

func NewNeteaseClient(cfg *config.Config) *NeteaseClient {
	client := &NeteaseClient{
		baseURL: cfg.NeteaseAPIURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	if cfg.NeteasePhone != "" && cfg.NeteasePass != "" {
		client.login(cfg.NeteasePhone, cfg.NeteasePass)
	}
	return client
}

func (c *NeteaseClient) login(phone, password string) {
	url := fmt.Sprintf("%s/login/cellphone?phone=%s&password=%s", c.baseURL, phone, password)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("Failed to login Netease: %v", err)
		return
	}
	defer resp.Body.Close()

	if cookies := resp.Header.Get("Set-Cookie"); cookies != "" {
		c.cookie = cookies
		log.Printf("NetEase login successful!")
	}
}

func (c *NeteaseClient) GetDailyRecommendations() ([]model.NeteaseSongDTO, error) {
	req, _ := http.NewRequest("GET", c.baseURL+"/recommend/songs", nil)
	if c.cookie != "" {
		req.Header.Add("Cookie", c.cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			DailySongs []model.NeteaseSongDTO `json:"dailySongs"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data.DailySongs, nil
}

func (c *NeteaseClient) GetLyric(id int64) (string, error) {
	url := fmt.Sprintf("%s/lyric?id=%d", c.baseURL, id)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Lrc struct {
			Lyric string `json:"lyric"`
		} `json:"lrc"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Lrc.Lyric, nil
}
