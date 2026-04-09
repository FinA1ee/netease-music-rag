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


	var artistNames []string
	for _, ar := range song.Ar {
		artistNames = append(artistNames, ar.Name)
	}
	artistStr := strings.Join(artistNames, "、")

	// ==============================
	// 超级增强版 Prompt（高区分度核心）
	// 包含：语言、歌手性别、声线、配器、曲风、情绪、主题、关键词
	prompt := fmt.Sprintf(`
你是专业音乐全维度分析师，**必须严格遵守以下所有指令**。

## 输出要求
1. **只返回纯 JSON**，无任何解释、无多余文字、无标点、无注释。
2. **所有字段名（key）必须是英文**，严格与下方格式一致，**禁止出现任何中文 key**。
3. **所有字段值（value）必须是中文**，包括 keywords，**禁止出现英文 value**。
4. **所有字段不允许为空**，不知道就合理推断。

## 输出格式（严格复制结构，一字不差）
{
  "keywords": ["关键词1","关键词2","关键词3","关键词4","关键词5"],
  "style": ["曲风1","曲风2","曲风3"],
  "mood": ["情绪1","情绪2","情绪3","情绪4","情绪5"],
  "theme": ["主题1","主题2","主题3","主题4","主题5"],
  "features": {
    "language": ["歌曲语言"],
    "singerNation": ["歌手国籍"],
    "singerGender": ["歌手性别"],
    "singerVoice": ["歌手声线"],
    "vocalType": ["演唱方式"],
    "instruments": ["配器1","配器2"],
    "speed": ["速度"],
    "intensity": ["情绪强度"],
    "scene": ["适用场景"],
    "arrangement": ["编曲密度"],
    "eraStyle": ["年代风格"],
    "emoColor": "情感色调"
  }
}

## 字段范例
- keywords: 歌词中最具代表性的中文关键词（5个）
- style: 音乐曲风标签，如 流行、民谣、电子、古风、R&B
- mood: 情绪标签，如 悲伤、温暖、思念、欢快、孤独
- theme: 歌词主题，如 爱情、离别、成长、自然、城市
- language: 歌曲语言，如 普通话 / 粤语 / 英语 / 韩语 / 日语
- singerNation: 歌手国籍，如 中国 / 中国台湾 / 中国香港 / 韩国 / 日本
- singerGender: 歌手性别，如 男 / 女 / 男生组合 / 女生组合 / 混搭组合
- singerVoice: 歌手声线，如 温柔 / 磁性 / 清亮 / 沙哑 / 甜美
- vocalType: 演唱方式，如 独唱 / 合唱 / 说唱 / 戏腔 / 念白 / 纯音乐
- instruments: 核心配器，如 钢琴 / 吉他 / 弦乐 / 电子合成器
- speed: 速度，仅选一项：慢板 / 中速 / 轻快 / 快速
- intensity: 情绪强度，仅选一项：平静 / 温和 / 强烈 / 炸裂 / 舒缓 / 激昂
- scene: 适用场景，如 睡前 / 通勤 / 运动 / 学习 / 约会
- arrangement: 编曲密度，仅选一项：极简 / 清淡 / 丰满 / 华丽 / 层次丰富 / 电子感
- eraStyle: 年代风格，仅选一项：现代 / 复古 / 00年代 / 10年代 / 20年代
- emoColor: 情感色调，仅一个字符串，选：冷色调 / 暖色调 / 中性色调

## 强制规则
- **features 内部所有 key 必须是英文，值必须是中文**
- emoColor 是字符串（非数组），其余 features 字段均为字符串数组
- 所有字段必须填写，**禁止返回 null / 空数组 / 空字符串**
- 只输出 JSON，**不要任何其他内容**

-----
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
	// If there were no lyrics, discard any keywords the LLM may have invented
	if strings.TrimSpace(lyrics) == "" {
		analysis.Keywords = []string{}
	}
	log.Printf("✅ LLM 完成歌曲分析 %v", analysis)
	return &analysis, nil
}

// func (l *LLMClient) GetSongEmbedding(ctx context.Context, song *model.NeteaseSongDTO) ([]float32, error) {

// 	var text string
// 	features := json.Unmarshal(song.LlmData.Features, &feat)

//     return fmt.Sprintf(`
// 歌曲名：%s
// 歌手：%s
// 语言：%s
// 性别：%s
// 声线：%s
// 配器：%s
// 风格：%s
// 情绪：%s
// 主题：%s
// 关键词：%s
// `,
//         song.Name,
//         getArtistNames(song),
//         feat.Language,
//         feat.SingerGender,
//         feat.SingerVoice,
//         strings.Join(feat.Instruments, ","),
//         song.Style,
//         song.Mood,
//         song.Theme,
//         song.Keywords,
//     )

// 	resp, err := l.client.Models.EmbedContent(ctx, l.embeddingModel, []*genai.Content{
// 		{Parts: []*genai.Part{{Text: text}}},
// 	}, &genai.EmbedContentConfig{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(resp.Embeddings) == 0 {
// 		return nil, fmt.Errorf("no embeddings returned")
// 	}
// 	return resp.Embeddings[0].Values, nil
// }

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
