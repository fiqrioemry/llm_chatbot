package constant

const (
	UploadDocumentSuccess = "Document uploaded, processing started"
	ListDocumentsSuccess  = "Documents retrieved"
	GetDocumentSuccess    = "Document retrieved"
	DeleteDocumentSuccess = "Document deleted"

	ErrDocumentNotFound       = "Document not found"
	CodeDocumentNotFound      = "DOC_NOT_FOUND"
	ErrDocumentInvalidType    = "File type not supported. Allowed: pdf, txt, md, docx"
	CodeDocumentInvalidType   = "DOC_INVALID_TYPE"
	ErrDocumentTooLarge       = "File size exceeds 20MB limit"
	CodeDocumentTooLarge      = "DOC_TOO_LARGE"
	ErrDocumentIngestionFail  = "Document ingestion failed"
	CodeDocumentIngestionFail = "DOC_INGEST_FAIL"
)
