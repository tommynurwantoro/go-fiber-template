package formatter

type SuccessResponse struct {
	Status   string    `json:"status" example:"success"`
	Message  string    `json:"message" example:"Operation completed successfully"`
	Data     any       `json:"data,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

type ErrorResponse struct {
	Status    string `json:"status" example:"APP05"`
	Message   string `json:"message" example:"Unexpected error"`
	TraceID   string `json:"trace_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	ErrorList any    `json:"error_list,omitempty"`
}

type Metadata struct {
	Page         int   `json:"page" example:"1"`
	Limit        int   `json:"limit" example:"10"`
	TotalPages   int64 `json:"total_pages" example:"1"`
	TotalResults int64 `json:"total_results" example:"1"`
}

func NewSuccessResponse(status Status, message string, data any) *SuccessResponse {
	return &SuccessResponse{
		Status:  status.String(),
		Message: message,
		Data:    data,
	}
}

func NewSuccessResponseWithMetadata(status Status, message string, data any, metadata Metadata) *SuccessResponse {
	return &SuccessResponse{
		Status:   status.String(),
		Message:  message,
		Data:     data,
		Metadata: &metadata,
	}
}

func NewErrorResponse(status Status, message string, traceID string) *ErrorResponse {
	return &ErrorResponse{
		Status:  status.String(),
		Message: message,
		TraceID: traceID,
	}
}

func NewErrorResponseList(status Status, message string, id string, err any) *ErrorResponse {
	return &ErrorResponse{
		Status:    status.String(),
		Message:   message,
		TraceID:   id,
		ErrorList: err,
	}
}
