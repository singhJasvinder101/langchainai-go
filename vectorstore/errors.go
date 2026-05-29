package vectorstore

import "errors"

var (
	ErrDocumentsRequired = errors.New("vectorstore: documents are required")
	ErrQueryRequired     = errors.New("vectorstore: query is required")
	ErrInvalidK          = errors.New("vectorstore: k must be greater than zero")
	ErrIDsRequired       = errors.New("vectorstore: ids are required")
	ErrDocumentContent   = errors.New("vectorstore: document content is required")
)
