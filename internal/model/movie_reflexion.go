package model

// Movie is a single film suggestion with title, year, and a short reason for the pick.
type Movie struct {
	Title  string `json:"title"`
	Year   int    `json:"year"`
	Reason string `json:"reason"`
}

// CinephileResponse is the raw JSON output from the cinephile agent.
type CinephileResponse struct {
	Movies []Movie `json:"movies"`
}

// ClerkResponse is the fact-checker's verdict on the current draft.
type ClerkResponse struct {
	Critiques   []string `json:"critiques"`
	IsSatisfied bool     `json:"isSatisfied"`
	Summary     string   `json:"summary"`
}

// FindMoviesState is the shared mutable state threaded through the find-movies graph.
type FindMoviesState struct {
	UserQuery       RefinedMovieQuery `json:"userQuery"`
	CurrentDraft    []Movie           `json:"movies"`
	CritiqueHistory []string          `json:"critiqueHistory"`
	LastSummary     string            `json:"last_summary"`
	RetryCount      int               `json:"retryCount"`
	MaxRetries      int               `json:"maxRetries"`
	IsSatisfied     bool              `json:"isSatisfied"`
	FinalAnswer     string            `json:"finalAnswer,omitempty"`
}

// RefinedMovieQuery holds structured search parameters extracted from free-text by the refiner chain.
type RefinedMovieQuery struct {
	PrimaryGenre string   `json:"primary_genre" doc:"The main genre (e.g. Drama)"`
	Secondary    []string `json:"secondary_genres" doc:"Other genres mentioned"`
	StartYear    int      `json:"start_year" doc:"Four digit start year"`
	EndYear      int      `json:"end_year" doc:"Four digit end year"`
	IsClassic    bool     `json:"is_classic" doc:"True if user mentions 'old', 'classic', or 'vintage'"`
	OriginalText string   `json:"original_text" doc:"The raw user input for reference"`
	QueryInfo    string   `json:"query_info" doc:"A concise search-optimised string distilled from the user intent, suitable for use as a web search query (e.g. 'surreal horror underground 1990s')"`
}
