package domain

type Index struct {
	DocID   string
	Tree    []Node
	Version int
}

type Node struct {
	ID        string
	Title     string
	Summary   string
	Text      string
	StartPage int
	EndPage   int
	Children  []Node
}
