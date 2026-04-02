package domain

// Index represents the hierarchical JSON tree index for a document.
type Index struct {
	DocID   string
	Tree    []Node
	Version int
}

// Node represents a section in the document tree.
type Node struct {
	ID        string
	Title     string
	Summary   string
	Text      string
	StartPage int
	EndPage   int
	Children  []Node
}
