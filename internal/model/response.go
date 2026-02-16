package model

// ResponseModel is a standardized API response structure (similar to .NET's ResponseModel).
// All API responses follow this format for consistency.
type ResponseModel struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`    // "00" = success, "01" = error
	Message string      `json:"message"`
	Result  interface{} `json:"result,omitempty"`
}

// NewSuccessResponse creates a successful response.
func NewSuccessResponse(result interface{}) *ResponseModel {
	return &ResponseModel{
		Success: true,
		Code:    "00",
		Message: "Success",
		Result:  result,
	}
}

// NewErrorResponse creates an error response.
func NewErrorResponse(message string) *ResponseModel {
	return &ResponseModel{
		Success: false,
		Code:    "01",
		Message: message,
		Result:  nil,
	}
}
