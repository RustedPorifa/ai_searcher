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

var apiText = `Найди 5 СВЕЖИХ новостей за ПОСЛЕДНЮЮ НЕДЕЛЮ (с %s по %s) по теме: "%s".
Верни ТОЛЬКО ВАЛИДНЫЙ JSON без каких-либо дополнительных символов, пояснений или markdown.
Каждый URL должен быть чистым, без пробелов и без сокращений.
Формат: {"topic":"...", "total_results":кол-во новостей в news, "news":[{"title":"...","url":"...","summary":"...","source":"...","published_at":"...","relevance_score":0.0}]}`

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
		return &newstypes.NewsResponse{}, jsonErr
	}

	httpReq, httpGenErr := http.NewRequest("POST", PERPLEXITY_URL, bytes.NewBuffer(jsonData))
	if httpGenErr != nil {
		return &newstypes.NewsResponse{}, httpGenErr
	}

	httpReq.Header.Set("Authorization", "Bearer "+perplexity_api)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &newstypes.NewsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &newstypes.NewsResponse{}, fmt.Errorf("perplexity API error: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &newstypes.NewsResponse{}, err
	}

	var apiResponse PerplexityAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return &newstypes.NewsResponse{}, err
	}

	if len(apiResponse.Choices) > 0 {
		content := apiResponse.Choices[0].Message.Content
		println(content)
		var response newstypes.NewsResponse
		if err := json.Unmarshal([]byte(content), &response); err != nil {
			return &newstypes.NewsResponse{}, err
		}

		redidb.CacheNews(topic, &response, time.Duration(24*time.Hour))

		redidb.AddToHistory(user_id, topic)
		return &response, nil
	} else {
		return &newstypes.NewsResponse{}, fmt.Errorf("no choices in response")
	}

	//return &NewsResponse{}, fmt.Errorf("no choices in response")
}
