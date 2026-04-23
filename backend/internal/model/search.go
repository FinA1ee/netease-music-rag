package model

// SearchRequest is the input for a natural-language song recommendation query.
type SearchRequest struct {
	Query string `json:"query"`           // e.g. "我想听一首快节奏男声强劲的歌"
	Limit int    `json:"limit,omitempty"` // number of results, defaults to 5

	// Optional deterministic filters parsed from prompt directives.
	ExactArtist   string  `json:"exact_artist,omitempty"`   // exact artist name match
	ExactAlbum    string  `json:"exact_album,omitempty"`    // exact album name match
	MinPopularity float32 `json:"min_popularity,omitempty"` // global popularity floor
}

// SongResult is a single recommendation returned by the search API.
type SongResult struct {
	SongID     int64    `json:"song_id"`
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	Album      Album    `json:"album"`
	Style      any      `json:"style"`
	Mood       any      `json:"mood"`
	Keywords   any      `json:"keywords"`
	Popularity float32  `json:"popularity"`
}

// SearchResponse is the JSON body returned by GET /api/search.
type SearchResponse struct {
	Query   string       `json:"query"`
	Results []SongResult `json:"results"`
}
