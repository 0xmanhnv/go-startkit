package response

// Common, stable error codes/messages for API responses.
// Keep messages short and safe; avoid leaking internals.

const (
	CodeInvalidRequest      = "invalid_request"
	CodeInvalidCredentials  = "invalid_credentials"
	CodeInvalidRefreshToken = "invalid_refresh_token"
	CodeNotFound            = "not_found"
	CodeConflict            = "conflict"
	CodeUnauthorized        = "unauthorized"
	CodeServerError         = "server_error"
)

const (
	MsgInvalidJSON         = "invalid JSON payload"
	MsgInvalidCredentials  = "email or password is incorrect"
	MsgInvalidRefreshToken = "refresh token invalid or expired"
	MsgNotFound            = "resource not found"
	MsgServerError         = "internal error"
)
