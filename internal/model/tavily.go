package model

// TavilyQuery is the input type for the Tavily search tool.
type TavilyQuery struct {
	Query string `json:"query" jsonschema:"description=The search query to look up"`
}
