// internal/handler/response_types.go
package handler

// Base responses
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"error message"`
}

// Session responses
type CreateSessionResponse struct {
	SessionID string `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CompleteChapterResponse struct {
	Chapter      int  `json:"chapter" example:"1"`
	Completed    bool `json:"completed" example:"true"`
	CompletedAll bool `json:"completed_all" example:"false"`
}

// Certificate responses
type RequestCertificateResponse struct {
	CertificateID string `json:"certificate_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status        string `json:"status" example:"pending"`
}

type CertificateStatusResponse struct {
	CertificateID string `json:"certificate_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status        string `json:"status" example:"done"`
}

type VerifyCertificateResponse struct {
	ID       string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name     string `json:"name" example:"John Doe"`
	CourseID string `json:"course_id" example:"golang-basics"`
	IssuedAt string `json:"issued_at" example:"2024-01-15T10:30:00Z"`
	Status   string `json:"status" example:"done"`
	Verified bool   `json:"verified" example:"true"`
}

// Request bodies
type CreateSessionRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
	CourseID string `json:"course_id" validate:"required" example:"golang-basics"`
}
