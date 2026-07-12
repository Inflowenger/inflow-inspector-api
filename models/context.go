package models

type ContextChangeType string

const (
	ByFlow ContextChangeType = "flow"
	ByAPI  ContextChangeType = "manual"
)

type ContextRecord struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Context   string         `json:"context"`
	Header    map[string]any `json:"header"`
	UpdatedAt int64          `json:"updatedAt"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedBy LastChange     `json:"updatedBy"`
}
type LastChange struct {
	By      ContextChangeType `json:"by"`
	Address string            `json:"address"`
}
type ContextDoc struct {
	Data   string `json:"data"`
	Header string `json:"header"`
}
