package domain

import "time"

type DocumentStatus string

const (
	StatusUploaded DocumentStatus = "uploaded" // file stored, not yet indexed
	StatusIndexing DocumentStatus = "indexing" // indexation in progress
	StatusReady    DocumentStatus = "ready"    // indexed, searchable
	StatusError    DocumentStatus = "error"    // indexation failed
)

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
