package googlebooks

// volumeListResponseDTO represents the JSON envelope from Google Books /volumes endpoint.
type volumeListResponseDTO struct {
	TotalItems int                 `json:"totalItems"`
	Items      []volumeResponseDTO `json:"items"`
}

// volumeResponseDTO represents a single volume in Google Books API.
type volumeResponseDTO struct {
	ID         string        `json:"id"`
	ETag       string        `json:"etag"`
	SelfLink   string        `json:"selfLink"`
	VolumeInfo volumeInfoDTO `json:"volumeInfo"`
}

// volumeInfoDTO holds the bibliographic metadata of a volume.
type volumeInfoDTO struct {
	Title               string                  `json:"title"`
	Subtitle            string                  `json:"subtitle"`
	Authors             []string                `json:"authors"`
	Publisher           string                  `json:"publisher"`
	PublishedDate       string                  `json:"publishedDate"` // e.g. "2006-05-23", "2006-05", or "2006"
	Description         string                  `json:"description"`
	IndustryIdentifiers []industryIdentifierDTO `json:"industryIdentifiers"`
	PageCount           int                     `json:"pageCount"`
	Categories          []string                `json:"categories"`
	ImageLinks          imageLinksDTO           `json:"imageLinks"`
	Language            string                  `json:"language"`
	PrintType           string                  `json:"printType"` // e.g. "BOOK"
}

// industryIdentifierDTO holds standard identifiers like ISBN_10 and ISBN_13.
type industryIdentifierDTO struct {
	Type       string `json:"type"` // "ISBN_10", "ISBN_13", "OTHER"
	Identifier string `json:"identifier"`
}

// imageLinksDTO provides cover image thumbnails across available resolutions.
type imageLinksDTO struct {
	SmallThumbnail string `json:"smallThumbnail"`
	Thumbnail      string `json:"thumbnail"`
	Small          string `json:"small"`
	Medium         string `json:"medium"`
	Large          string `json:"large"`
	ExtraLarge     string `json:"extraLarge"`
}
