package v1

var (
	// common errors
	ErrSuccess             = NewError(0, "ok")
	ErrBadRequest          = NewError(400, "Bad Request")
	ErrUnauthorized        = NewError(401, "Unauthorized")
	ErrNotFound            = NewError(404, "Not Found")
	ErrInternalServerError = NewError(500, "Internal Server Error")

	// more biz errors
	ErrEmailAlreadyUse = NewError(1001, "The email is already in use.")
	ErrInvalidParams   = NewError(1002, "Invalid params.")
)
