package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"onepractice-golang/internal/agent/llm"
	essayPrompt "onepractice-golang/internal/agent/prompt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// 英语作文批改助手

// maxParseRetries 模型输出无法解析为 JSON 时的最大重试次数。
const maxParseRetries = 2

// essayMaxOutputTokens 限制单次生成的最大 token 数，避免长 JSON 被截断。
const essayMaxOutputTokens = 20000

// NewEssayChatModel 构建作文批改使用的 chat model：开启 JSON 输出模式并限制最大 token 数，
// 减少 markdown 包裹、多余解释与长 JSON 截断，配合 EssayScore 的重试兜底。
func NewEssayChatModel(ctx context.Context, provider llm.Provider, apiKey string) (model.BaseChatModel, error) {
	return llm.NewChatModel(ctx, provider, apiKey)
}

// parseRetryPrompt 解析失败后追加进对话的纠正指令。
const parseRetryPrompt = `你上一次的输出无法被解析为合法 JSON。请严格修正后重新输出：
1. 只输出纯 JSON，不要 markdown 代码块（不要使用 ` + "```" + ` 包裹），不要任何解释或多余文字。
2. 严格使用系统提示中要求的 JSON 结构，不要增删、改名任何字段。
3. 所有字符串使用合法转义，数字不要加引号。`

type Input struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type Output struct {
	RawEssay        string  `json:"rawEssay"`
	Title           string  `json:"title"`
	Type            string  `json:"type"`
	WordNum         int     `json:"wordNum"`
	SentNum         int     `json:"sentNum"`
	ParaNum         int     `json:"paraNum"`
	FullScore       int     `json:"fullScore"`
	TotalScore      float64 `json:"totalScore"`
	TotalEvaluation string  `json:"totalEvaluation"`
	EssayAdvice     string  `json:"essayAdvice"`
	MajorScore      struct {
		GrammarScore    float64 `json:"grammarScore"`
		GrammarAdvice   string  `json:"grammarAdvice"`
		TopicScore      float64 `json:"topicScore"`
		TopicAdvice     string  `json:"topicAdvice"`
		WordScore       float64 `json:"wordScore"`
		WordAdvice      string  `json:"wordAdvice"`
		StructureScore  float64 `json:"structureScore"`
		StructureAdvice string  `json:"structureAdvice"`
	} `json:"majorScore"`
	EssayFeedback struct {
		OverallProblems []string `json:"overallProblems"`
		SentsFeedback   []struct {
			SentId                int    `json:"sentId"`
			ParaId                int    `json:"paraId"`
			RawSent               string `json:"rawSent"`
			CorrectedSent         string `json:"correctedSent"`
			IsContainGrammarError bool   `json:"isContainGrammarError"`
			ErrorPosInfos         []struct {
				OrgChunk       string `json:"orgChunk"`
				CorrectChunk   string `json:"correctChunk"`
				ErrorTypeTitle string `json:"errorTypeTitle"`
				ErrBaseInfo    string `json:"errBaseInfo"`
				DetailReason   string `json:"detailReason"`
			} `json:"errorPosInfos"`
			SentFeedback string `json:"sentFeedback"`
		} `json:"sentsFeedback"`
	} `json:"essayFeedback"`
	RewriteSuggestions []struct {
		Original  string `json:"original"`
		Rewritten string `json:"rewritten"`
		Reason    string `json:"reason"`
	} `json:"rewriteSuggestions"`
	ImprovementAdvice []string `json:"improvementAdvice"`
}

// EssayScore 调用 chat model 批改作文。解析失败时会把模型原文与纠正指令追加进对话后重试，
// 最多重试 maxParseRetries 次。
func EssayScore(ctx context.Context, cm model.BaseChatModel, input Input) (Output, error) {
	if cm == nil {
		return Output{}, errors.New("chat model 未初始化")
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return Output{}, err
	}

	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage(essayPrompt.WriterPrompt),
		schema.UserMessage("{input}"),
	)
	messages, err := template.Format(ctx, map[string]any{"input": string(inputJSON)})
	if err != nil {
		return Output{}, err
	}

	var (
		lastErr error
		output  Output
	)
	for attempt := 0; attempt <= maxParseRetries; attempt++ {
		output = Output{}

		resp, genErr := cm.Generate(ctx, messages)
		if genErr != nil {
			return Output{}, genErr
		}

		if resp != nil {
			if unmarshalErr := json.Unmarshal([]byte(normalizeJSON(resp.Content)), &output); unmarshalErr == nil {
				return output, nil
			} else {
				lastErr = unmarshalErr
			}
		} else {
			lastErr = errors.New("模型返回空响应")
		}

		// 解析失败：把模型上一次的原文与纠正指令追加进对话后再重试。
		if resp != nil {
			messages = append(messages, resp)
		}
		messages = append(messages, schema.UserMessage(parseRetryPrompt))
	}

	return Output{}, fmt.Errorf("作文批改结果解析失败（已重试 %d 次）: %w", maxParseRetries, lastErr)
}

// normalizeJSON 去除模型可能包裹的 markdown 代码块围栏。
func normalizeJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
