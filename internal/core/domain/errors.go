package domain

import "fmt"

type NotFoundError struct {
	Resource string // "document", "index", "file"
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

type NotReadyError struct {
	DocID  string
	Status DocumentStatus
}

func (e *NotReadyError) Error() string {
	return fmt.Sprintf("document %s is %s, cannot proceed", e.DocID, e.Status)
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func ErrDocumentNotFound(id string) error { return &NotFoundError{Resource: "document", ID: id} }
func ErrIndexNotFound(id string) error    { return &NotFoundError{Resource: "index", ID: id} }
func ErrFileNotFound(id string) error     { return &NotFoundError{Resource: "file", ID: id} }
