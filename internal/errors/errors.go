package errors

type AppError struct {
	Msg        string
	StatusCode int
}

func NewError(msg string, statusCode int) *AppError {
	return &AppError{
		Msg:        msg,
		StatusCode: statusCode,
	}
}

var (
	ErrUserAlreadyExists      = NewError("user already exists", 400)
	ErrNotFound               = NewError("not found", 404)
	ErrInvalidEmailOrPassword = NewError("invalid password or email", 400)
	ErrInvalidOTP             = NewError("invalid otp", 400)
	ErrUnathorized            = NewError("unauthorized", 401)
	ErrForbidden              = NewError("forbidden", 403) //client's identity is known but has no permission to access the resource. has to do with roles and permissions
	ErrInternalServer         = NewError("internal server error", 500)
	ErrFileTooLarge           = NewError("file too large", 413)
	ErrInvalidObjectID        = NewError("invalid object id", 400)
	ErrCategoryAlreadyExists  = NewError("category already exists", 400)
	ErrUserAlreadyVerified    = NewError("user already verified", 409)
)

func (e *AppError) Error() string {
	return e.Msg
}
