package errs

type ErrorCode int

const (
	Success            ErrorCode = 0
	BadRequest         ErrorCode = 400
	Unauthorized       ErrorCode = 401
	Forbidden          ErrorCode = 403
	NotFound           ErrorCode = 404
	InternalServer     ErrorCode = 500
	ServiceUnavailable ErrorCode = 503
)

var codeMessages = map[ErrorCode]string{
	Success:            "Success",
	BadRequest:         "Bad Request",
	Unauthorized:       "Unauthorized",
	Forbidden:          "Forbidden",
	NotFound:           "Not Found",
	InternalServer:     "Internal Server Error",
	ServiceUnavailable: "Service Unavailable",
}

func (c ErrorCode) Message() string {
	if msg, ok := codeMessages[c]; ok {
		return msg
	}
	return "Unknown Error"
}
