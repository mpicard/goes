package graph

type Todo struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Author User   `json:"author"`
}
