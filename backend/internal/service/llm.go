package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"netease-music-rag/backend/internal/model"

	"google.golang.org/genai"
)

type LLMClient struct {
	client   *genai.Client
	llmModel string
	// embeddingModel string
}

func NewLLMClient(apiKey, llmModel string) (*LLMClient, error) {

	// 1. 配置 Clash 代理
	proxyURL, err := url.Parse("http://127.0.0.1:7897")
	if err != nil {
		panic(err)
	}

	// 2. 创建带代理的 HTTP Client
	hc := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 2. 设置代理（你的 Clash 端口 7897）
	if proxyURL != nil {
		proxy, err := url.Parse(proxyURL.String())
		if err != nil {
			return nil, err
		}
		hc.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:     apiKey,
		HTTPClient: hc,
	})
	if err != nil {
		return nil, err
	}

	return &LLMClient{
		client:   client,
		llmModel: llmModel,
		// embeddingModel: embeddingModel,
	}, nil
}

func (l *LLMClient) AnalyzeSong(ctx context.Context, song *model.NeteaseSongDTO, lyrics string) (*model.LLMAnalysisResult, error) {

	if lyrics == "" {
		return nil, fmt.Errorf("歌词为空，跳过LLM分析")
	}

	var artistNames []string
	for _, ar := range song.Ar {
		artistNames = append(artistNames, ar.Name)
	}
	artistStr := strings.Join(artistNames, "、")

	// ==============================
	// 超级增强版 Prompt（高区分度核心）
	// 包含：语言、歌手性别、声线、配器、曲风、情绪、主题、关键词
	// ==============================
	prompt := fmt.Sprintf(`
你是专业音乐全维度分析师，请严格按要求分析歌曲，只返回JSON，无任何多余内容。

输出字段说明：
1. keywords: 5-10个歌词代表性核心实词（区分度高）
2. style: 2-4个精准细分曲风标签（Mandopop、粤语流行、民谣、R&B、抒情、摇滚等）
3. mood: 3-6个情绪标签（治愈、孤独、伤感、甜蜜、热血、平静等）
4. theme: 3-6个主题标签（爱情、成长、思念、离别、梦想、自我等）
5. features: 必须包含以下6类特征，用简短标签存放
   - 歌曲语言（普通话/粤语/英语/日语等）
   - 歌手性别（男/女/组合）
   - 歌手声线（清亮/烟嗓/磁性/甜嗓/浑厚/温柔）
   - 核心配器器乐（钢琴/吉他/鼓/弦乐/合成器/古筝等）
   - 歌曲速度（快/慢/中等）
   - 歌曲年代（1980/1990/2020）

规则：
- 标签精准、简短、高区分度
- 不使用无效通用词
- 严格返回JSON格式

歌曲：%s
歌手：%s
歌词：
%s
`,
		song.Name,
		artistStr,
		song.Lyric,
	)

	// 调用 Gemini API
	var temperature float32 = 0.2
	result, err := l.client.Models.GenerateContent(ctx, l.llmModel, []*genai.Content{
		{Parts: []*genai.Part{{Text: prompt}}},
	}, &genai.GenerateContentConfig{
		Temperature: &temperature,
	})

	if err != nil {
		return nil, err
	}

	text := extractTextFromResult(result)
	text = cleanJSONWrapper(text)

	var analysis model.LLMAnalysisResult
	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		log.Printf("Failed to unmarshal JSON: %s", text)
		return nil, err
	}
	log.Printf("✅ LLM 完成歌曲分析")
	return &analysis, nil
}

func (l *LLMClient) DryRun(ctx context.Context) {
	resp, err := l.client.Models.GenerateContent(ctx, l.llmModel, genai.Text("你好，测试一下"), nil)
	if err != nil {
		panic(fmt.Sprintf("调用 Gemini 失败: %v", err))
	}

	// 5. 打印结果
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		fmt.Println("Gemini 响应:", resp.Candidates[0].Content.Parts[0].Text)
	}
}

// func (l *LLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
// 	resp, err := l.client.Models.EmbedContent(ctx, l.embeddingModel, &genai.EmbedContentRequest{
// 		Contents: []*genai.Content{
// 			{Parts: []*genai.Part{{Text: text}}},
// 		},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(resp.Embeddings) == 0 {
// 		return nil, fmt.Errorf("no embeddings returned")
// 	}
// 	return resp.Embeddings[0].Values, nil
// }

func extractTextFromResult(result *genai.GenerateContentResponse) string {
	if result != nil && len(result.Candidates) > 0 {
		cand := result.Candidates[0]
		if cand.Content != nil && len(cand.Content.Parts) > 0 {
			return cand.Content.Parts[0].Text
		}
	}
	return ""
}

func cleanJSONWrapper(js string) string {
	re := regexp.MustCompile(`(?s)^\s*` + "```" + `(?:json)?\s*(.*?)\s*` + "```" + `\s*$`)
	if match := re.FindStringSubmatch(js); len(match) == 2 {
		return match[1]
	}
	return js
}
