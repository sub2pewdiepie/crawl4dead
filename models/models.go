package models

type Link struct {
	URL        string
	Source     string
	IsExternal bool
}

type Result struct {
	Link   Link
	Status string
}
