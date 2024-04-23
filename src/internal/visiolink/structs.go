package visiolink

type Catalog struct {
	Customer        string `json:"customer"`
	PublicationDate string `json:"publication_date"`
	Title           string `json:"title"`
	Sections        []struct {
		FrontPage int `json:"front_page"`
	} `json:"sections"`
	FolderID int `json:"folder_id"`
	Catalog  int `json:"catalog"`
	Pages    int `json:"pages"`
}

type Content struct {
	Generated      string    `json:"generated"`
	TeaserImageURL string    `json:"teaser_image_url"`
	CatalogURL     string    `json:"catalog_url"`
	Catalogs       []Catalog `json:"catalogs"`
}

type TokenResponse struct {
	AccessURL string `json:"access_url"`
	Success   bool   `json:"success"`
}

type Paper struct {
	Customer     string
	Domain       string
	LoginDomain  string
	LoginUrl     string
	ReaderDomain string
	CatalogId    int16
}
