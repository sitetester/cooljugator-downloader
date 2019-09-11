package entity

type LinksByLanguageCode struct {
	LanguageCode string
	Links        []Link
}

type Link struct {
	Href     string
	HrefText string
}
