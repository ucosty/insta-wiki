package models

type Page struct {
	Id            int
	IsNewPage     bool
	Path          string
	Title         string
	BodyMarkdown  string // Body in raw Markdown format
	BodyFormatted string `json:"-"` // Body in HTML format
	Revision      int
}
