package models

type Link struct {
	URL        string
	Source     string
	Depth      int
	IsExternal bool
}

type Result struct {
	Link       Link
	Status     string
	StatusCode int32
}
