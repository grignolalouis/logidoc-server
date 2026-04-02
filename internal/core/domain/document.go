package domain

import "time"

// DocumentStatus represents the lifecycle state of a document.
type DocumentStatus string

const (
	StatusUploaded DocumentStatus = "uploaded" // file stored, not yet indexed
	StatusIndexing DocumentStatus = "indexing" // indexation in progress
	StatusReady    DocumentStatus = "ready"    // indexed, searchable
	StatusError    DocumentStatus = "error"    // indexation failed
)

// Document represents a document entity.
type Document struct {
	ID          string
	Name        string
	Description string
	Status      DocumentStatus
	PageCount   int
	NodeCount   int
	Error       string
	CreatedAt   time.Time
	IndexedAt   *time.Time
}
