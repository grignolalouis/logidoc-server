package domain

import "errors"

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrIndexNotFound    = errors.New("index not found for document")
	ErrDocumentNotReady = errors.New("document is not yet indexed")
	ErrInvalidDocument  = errors.New("invalid document format")
	ErrIndexingFailed   = errors.New("indexing failed")
)
