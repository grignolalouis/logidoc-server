package domain

type SearchHit struct {
	DocID     string
	DocName   string
	NodeID    string
	NodeTitle string
	Summary   string
	StartPage int
	EndPage   int
	Score     float64
}
