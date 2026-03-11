package model

// TavilyQuery is the input type for the Tavily search tool.
type TavilyQuery struct {
	Query      string `json:"query" jsonschema:"description=The search query to look up"`
	MaxResults int    `json:"max_results" jsonschema:"description=max results to return"`
}

// TavilySearchResponse is the top-level shape of a Tavily search response.
type TavilySearchResponse struct {
	Results []TavilyResult `json:"results"`
}

// TavilyResult holds only the fields we care about from each Tavily result.
type TavilyResult struct {
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}
