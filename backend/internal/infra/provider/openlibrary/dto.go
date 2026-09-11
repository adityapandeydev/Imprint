package openlibrary

// searchResponseDTO represents the JSON envelope from Open Library's /search.json endpoint.
type searchResponseDTO struct {
	NumFound int            `json:"numFound"`
	Docs     []searchDocDTO `json:"docs"`
}

// searchDocDTO represents an individual search result item in /search.json.
type searchDocDTO struct {
	Key              string   `json:"key"` // e.g. "/works/OL27479W"
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	FirstPublishYear int      `json:"first_publish_year"`
	ISBN             []string `json:"isbn"`
	CoverI           int      `json:"cover_i"`
	EditionKey       []string `json:"edition_key"`
	Subject          []string `json:"subject"`
}

// workResponseDTO represents the response from /works/{id}.json.
type workResponseDTO struct {
	Key         string   `json:"key"`
	Title       string   `json:"title"`
	Description any      `json:"description"` // string OR {"type": "/type/text", "value": "..."}
	Subjects    []string `json:"subjects"`
	Covers      []int    `json:"covers"`
}

// editionDTO represents an individual edition from /isbn/{isbn}.json or /works/{id}/editions.json.
type editionDTO struct {
	Key            string      `json:"key"` // e.g. "/books/OL7353617M"
	Title          string      `json:"title"`
	Publishers     []string    `json:"publishers"`
	PublishDate    string      `json:"publish_date"`
	NumberOfPages  int         `json:"number_of_pages"`
	ISBN10         []string    `json:"isbn_10"`
	ISBN13         []string    `json:"isbn_13"`
	PhysicalFormat string      `json:"physical_format"`
	Covers         []int       `json:"covers"`
	Description    any         `json:"description"`
	Works          []workRefDTO `json:"works"`
}

type workRefDTO struct {
	Key string `json:"key"` // e.g. "/works/OL27479W"
}

type editionsListResponseDTO struct {
	Size    int          `json:"size"`
	Entries []editionDTO `json:"entries"`
}
