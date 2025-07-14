package models

type Link struct {
	URL        string
	Source     string
	Depth      int
	IsExternal bool
}

type Result struct {
	Link       Link
	StatusCode int32
	Status     string
}
