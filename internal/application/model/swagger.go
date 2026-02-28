package model

// API response wrappers for Swagger documentation (match formatter.Response structure)

// SuccessMessageAPIResponse represents API responses with status and message only (logout, forgot-password, reset-password, send-verification-email, verify-email, delete-user)
type SuccessMessageAPIResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Operation completed successfully"`
}

// HealthCheckAPIResponseError represents the health check API response when unhealthy
type HealthCheckAPIResponseError struct {
	Code      int           `json:"code" example:"500"`
	Status    string        `json:"status" example:"error"`
	Message   string        `json:"message" example:"Health check completed"`
	IsHealthy bool          `json:"is_healthy" example:"false"`
	Result    []HealthCheck `json:"result"`
}

// Error response types for Swagger documentation

// ErrorInvalidRequest represents 400 error response (validation or body parse failed)
type ErrorInvalidRequest struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid request body or validation failed"`
	TraceID string `json:"traceId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ErrorNotFound represents 404 error response
type ErrorNotFound struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Not found"`
	TraceID string `json:"traceId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ErrorUnauthorized represents 401 error response
type ErrorUnauthorized struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Please authenticate"`
	TraceID string `json:"traceId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ErrorForbidden represents 403 error response
type ErrorForbidden struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"You don't have permission to access this resource"`
	TraceID string `json:"traceId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ErrorDuplicateEmail represents 409 error when email already exists
type ErrorDuplicateEmail struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Email already taken"`
	TraceID string `json:"traceId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ErrorFailedLogin represents 401 error for invalid credentials
type ErrorFailedLogin struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid email or password"`
	TraceID string `json:"traceId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ErrorFailedResetPassword represents 401 error for reset password failure
type ErrorFailedResetPassword struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Password reset failed"`
	TraceID string `json:"traceId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ErrorFailedVerifyEmail represents 401 error for verify email failure
type ErrorFailedVerifyEmail struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Verify email failed"`
	TraceID string `json:"traceId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}
