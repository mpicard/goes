package graph

type Todo struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Author User   `json:"author"`
}
type Filter struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
