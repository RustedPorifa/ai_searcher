package newstypes

type NewsItem struct {
	Title          string  `json:"title"`
	URL            string  `json:"url"`
	Summary        string  `json:"summary"`
	Source         string  `json:"source"`
	PublishedAt    string  `json:"published_at"`
	RelevanceScore float64 `json:"relevance_score"`
}

type NewsResponse struct {
	News         []NewsItem `json:"news"`
	TotalResults int        `json:"total_results"`
	Topic        string     `json:"topic"`
}
