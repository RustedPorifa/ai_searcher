package perplexityapi

import (
	"ai_tg_search/redidb"
	"ai_tg_search/struct_types/newstypes"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const PERPLEXITY_URL = "https://api.perplexity.ai/chat/completions"

var apiText = `Найди 5 СВЕЖИХ новостей за ПОСЛЕДНЮЮ НЕДЕЛЮ по теме: "%s".
Верни ТОЛЬКО ВАЛИДНЫЙ JSON-ОБЪЕКТ в формате, указанном ниже.
НЕ ДОБАВЛЯЙ никаких пояснений, комментариев, markdown-блоков, тройных кавычек, префиксов вроде "json" или иных символов.
НЕ ИСПОЛЬЗУЙ обратные кавычки, звёздочки, тире или любые другие символы вне JSON.
Ответ должен начинаться с символа '{' и заканчиваться символом '}'.
Все URL должны быть чистыми, без пробелов, переносов и сокращений.
Текст исключительно на русском языке.

Формат:
{"topic":"...", "total_results":число, "news":[{"title":"...","url":"...","summary":"...","source":"...","published_at":"ГГГГ-ММ-ДД","relevance_score":число_от_0_до_1}]}`

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PerplexityRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	Format    string    `json:"format,omitempty"`
}

type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Snippet     string `json:"snippet"`
	Date        string `json:"date"`
	LastUpdated string `json:"last_updated"`
}

type PerplexityResponse struct {
	Results []SearchResult `json:"results"`
	ID      string         `json:"id"`
}

type PerplexityAPIResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

var perplexity_api string

func InitPerplexityAPI() {
	perplexity_api = os.Getenv("PERPLEXITY_API")
}

func FindNews(topic string, user_id int64) (*newstypes.NewsResponse, error) {
	reqBody := PerplexityRequest{
		Model: "sonar",
		Messages: []Message{
			{
				Role:    "user",
				Content: fmt.Sprintf(apiText, topic),
			},
		},
		MaxTokens: 2048,
		Format:    "json",
	}

	jsonData, jsonErr := json.Marshal(reqBody)
	if jsonErr != nil {
		log.Println("Ошибка при преобразовании запроса в JSON:", jsonErr)
		return nil, jsonErr
	}

	httpReq, httpGenErr := http.NewRequest("POST", PERPLEXITY_URL, bytes.NewBuffer(jsonData))
	if httpGenErr != nil {
		return nil, httpGenErr
	}

	httpReq.Header.Set("Authorization", "Bearer "+perplexity_api)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("perplexity API error: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse PerplexityAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	if len(apiResponse.Choices) > 0 {
		content := apiResponse.Choices[0].Message.Content
		println(content)
		var response newstypes.NewsResponse
		if err := json.Unmarshal([]byte(content), &response); err != nil {
			return nil, err
		}

		redidb.CacheNews(topic, &response, time.Duration(24*time.Hour))

		redidb.AddToHistory(user_id, topic)
		return &response, nil
	} else {
		return nil, fmt.Errorf("no choices in response")
	}

	//return &NewsResponse{}, fmt.Errorf("no choices in response")
}
