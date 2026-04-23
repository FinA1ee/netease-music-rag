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
		c.auth.cookie = "
MUSIC_U=00227B40AAEF28381E124839968725BF7C9F6298E08018CA970A3A635BD157EA46CC383947125B11E2B06F5997DC57F7EEC21678BE872B8458D4585BA14F8861C9A8159CB628729346C20262F601F142EC3F96BBD058AAF03188A821D74403A43B84147DF6470C6DCB0B8FCFBB5024C83674A6BCE45AFD344FBB15A0E55FC5445D3A35A514D51B246B7CC19E3315655070DE9AEB14B28E0AAFC14938739CE49CD804C79BD17ECEE88E9B23E6CC058BB78D0315A26FBBA707EA9BCEC254ADAAC3AEE895605A53AF18A6B4C7200A0538E989A790B00898E6DAD651ADCD3BC57984F87072417AAF98FE989D1791B5B89231C1AFF21D09C6836F0CF00A8E901163149C144A3B07BB765BBFEC09BAEBAB3DB6B73FEF5FF10120E909EE48E1C17133E4A225627109FAB6E56A0F94B62E152E40D6FFF36C767DD7C63E750F544413A0C2E41A417AB10303A86C1B4F813CDFF1A6343183F6194E174646E4B7992C0192BDCA16D4EDE5A1E630D2835ACD3E89A2A1C0; NMTID=00OzXtVepSesCeZUE7WmR9fW8IJ0WUAAAGddj3GKQ"
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
		req.Header.Set("Cookie", "
MUSIC_U=00227B40AAEF28381E124839968725BF7C9F6298E08018CA970A3A635BD157EA46CC383947125B11E2B06F5997DC57F7EEC21678BE872B8458D4585BA14F8861C9A8159CB628729346C20262F601F142EC3F96BBD058AAF03188A821D74403A43B84147DF6470C6DCB0B8FCFBB5024C83674A6BCE45AFD344FBB15A0E55FC5445D3A35A514D51B246B7CC19E3315655070DE9AEB14B28E0AAFC14938739CE49CD804C79BD17ECEE88E9B23E6CC058BB78D0315A26FBBA707EA9BCEC254ADAAC3AEE895605A53AF18A6B4C7200A0538E989A790B00898E6DAD651ADCD3BC57984F87072417AAF98FE989D1791B5B89231C1AFF21D09C6836F0CF00A8E901163149C144A3B07BB765BBFEC09BAEBAB3DB6B73FEF5FF10120E909EE48E1C17133E4A225627109FAB6E56A0F94B62E152E40D6FFF36C767DD7C63E750F544413A0C2E41A417AB10303A86C1B4F813CDFF1A6343183F6194E174646E4B7992C0192BDCA16D4EDE5A1E630D2835ACD3E89A2A1C0; NMTID=00OzXtVepSesCeZUE7WmR9fW8IJ0WUAAAGddj3GKQ")
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
		req.Header.Set("Cookie", "
MUSIC_U=00227B40AAEF28381E124839968725BF7C9F6298E08018CA970A3A635BD157EA46CC383947125B11E2B06F5997DC57F7EEC21678BE872B8458D4585BA14F8861C9A8159CB628729346C20262F601F142EC3F96BBD058AAF03188A821D74403A43B84147DF6470C6DCB0B8FCFBB5024C83674A6BCE45AFD344FBB15A0E55FC5445D3A35A514D51B246B7CC19E3315655070DE9AEB14B28E0AAFC14938739CE49CD804C79BD17ECEE88E9B23E6CC058BB78D0315A26FBBA707EA9BCEC254ADAAC3AEE895605A53AF18A6B4C7200A0538E989A790B00898E6DAD651ADCD3BC57984F87072417AAF98FE989D1791B5B89231C1AFF21D09C6836F0CF00A8E901163149C144A3B07BB765BBFEC09BAEBAB3DB6B73FEF5FF10120E909EE48E1C17133E4A225627109FAB6E56A0F94B62E152E40D6FFF36C767DD7C63E750F544413A0C2E41A417AB10303A86C1B4F813CDFF1A6343183F6194E174646E4B7992C0192BDCA16D4EDE5A1E630D2835ACD3E89A2A1C0; NMTID=00OzXtVepSesCeZUE7WmR9fW8IJ0WUAAAGddj3GKQ")
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
		req.Header.Set("Cookie", "
MUSIC_U=00227B40AAEF28381E124839968725BF7C9F6298E08018CA970A3A635BD157EA46CC383947125B11E2B06F5997DC57F7EEC21678BE872B8458D4585BA14F8861C9A8159CB628729346C20262F601F142EC3F96BBD058AAF03188A821D74403A43B84147DF6470C6DCB0B8FCFBB5024C83674A6BCE45AFD344FBB15A0E55FC5445D3A35A514D51B246B7CC19E3315655070DE9AEB14B28E0AAFC14938739CE49CD804C79BD17ECEE88E9B23E6CC058BB78D0315A26FBBA707EA9BCEC254ADAAC3AEE895605A53AF18A6B4C7200A0538E989A790B00898E6DAD651ADCD3BC57984F87072417AAF98FE989D1791B5B89231C1AFF21D09C6836F0CF00A8E901163149C144A3B07BB765BBFEC09BAEBAB3DB6B73FEF5FF10120E909EE48E1C17133E4A225627109FAB6E56A0F94B62E152E40D6FFF36C767DD7C63E750F544413A0C2E41A417AB10303A86C1B4F813CDFF1A6343183F6194E174646E4B7992C0192BDCA16D4EDE5A1E630D2835ACD3E89A2A1C0; NMTID=00OzXtVepSesCeZUE7WmR9fW8IJ0WUAAAGddj3GKQ")
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
		req.Header.Set("Cookie", "MUSIC_U=00227B40AAEF28381E124839968725BF7C9F6298E08018CA970A3A635BD157EA46CC383947125B11E2B06F5997DC57F7EEC21678BE872B8458D4585BA14F8861C9A8159CB628729346C20262F601F142EC3F96BBD058AAF03188A821D74403A43B84147DF6470C6DCB0B8FCFBB5024C83674A6BCE45AFD344FBB15A0E55FC5445D3A35A514D51B246B7CC19E3315655070DE9AEB14B28E0AAFC14938739CE49CD804C79BD17ECEE88E9B23E6CC058BB78D0315A26FBBA707EA9BCEC254ADAAC3AEE895605A53AF18A6B4C7200A0538E989A790B00898E6DAD651ADCD3BC57984F87072417AAF98FE989D1791B5B89231C1AFF21D09C6836F0CF00A8E901163149C144A3B07BB765BBFEC09BAEBAB3DB6B73FEF5FF10120E909EE48E1C17133E4A225627109FAB6E56A0F94B62E152E40D6FFF36C767DD7C63E750F544413A0C2E41A417AB10303A86C1B4F813CDFF1A6343183F6194E174646E4B7992C0192BDCA16D4EDE5A1E630D2835ACD3E89A2A1C0; NMTID=00OzXtVepSesCeZUE7WmR9fW8IJ0WUAAAGddj3GKQ")
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
