package errors

import (
	"fmt"
	"net/http"

	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

type Type string

const (
	JsonUnmarshalError          Type = "JsonUnmarshalError"
	JsonMarshalError            Type = "JsonMarshalError"
	NotInitializedError         Type = "NotInitialized"
	RequestTimeoutError         Type = "RequestTimeoutError"
	BadRequestError             Type = "BadRequestError"
	ValidationError             Type = "ValidationError"
	InternalServerError         Type = "InternalServerError"
	NotFoundError               Type = "NotFoundError"
	UnauthorizedError           Type = "UnauthorizedError"
	ForbiddenError              Type = "ForbiddenError"
	CannotReadResponseBodyError Type = "CannotReadResponseBodyError"
	ResponseBodyNilError        Type = "ResponseBodyNilError"
)

type Fields map[string]interface{}

type Error struct {
	Code        int32  `json:"code"`
	Message     string `json:"message"`
	Type        Type   `json:"type"`
	ServiceName string `json:"service_name"`
	Internal    error  `json:"-"`

	Fields Fields `json:"fields"`
}

var nilError = Error{Code: http.StatusOK}

func New(message string, errorType Type, fields ...interface{}) Error {
	return Error{
		Message: message,
		Type:    errorType,
		Fields:  argsToFields(fields),
	}
}

func NewLinkedinError(code int32, message string, errorType Type, fields ...interface{}) Error {
	return Error{
		Type:    errorType,
		Message: message,
		Code:    code,
		Fields:  argsToFields(fields),
	}
}

func Nil() Error {
	return nilError
}

func NewInternalError(message string, errorType Type, err error, fields ...interface{}) Error {
	return New(message, errorType, fields).WithInternal(pkgerrors.Wrap(err, message))
}

func (e Error) FromService(service string) Error {
	e.ServiceName = service
	return e
}

func (e Error) WithStatusCode(code int32) Error {
	e.Code = code
	return e
}

func (e Error) WithInternal(err error) Error {
	e.Internal = err
	return e
}

func (e Error) Wrap(message string) Error {
	e.Internal = pkgerrors.Wrap(e.Internal, message)
	return e
}

func (e Error) WithStackTrace() Error {
	e.Internal = pkgerrors.WithStack(e.Internal)
	return e
}

func (e Error) IsInternal() bool {
	if e.Code == http.StatusInternalServerError && e.Internal != nil {
		return true
	}
	return false
}

func (e Error) IsNil() bool {
	if e.Code == http.StatusOK && e.Internal == nil && e.Message == "" && e.Type == "" {
		return true
	}
	return false
}

func (e Error) IsNotNil() bool {
	return !e.IsNil()
}

func (e Error) Error() string {
	return fmt.Sprintf("status: %d, message: %s, internal: %s, type: %s, service: %s", e.Code, e.Message, e.Internal, e.Type, e.ServiceName)
}

func (e Error) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	oe.AddString("message", e.Message)
	oe.AddString("type", string(e.Type))
	oe.AddInt32("code", e.Code)
	oe.AddString("service", e.ServiceName)
	if e.Internal != nil {
		oe.AddString("internal", e.Internal.Error())
	}
	return oe.AddReflected("fields", e.Fields)
}

func argsToFields(args []interface{}) Fields {
	if len(args) == 0 {
		return Fields{}
	}

	fields := make(Fields)

	for i := 0; i < len(args); {
		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			break
		}

		key, val := args[i], args[i+1]
		i += 2

		keyStr, ok := key.(string)
		if !ok {
			continue
		}
		fields[keyStr] = val
	}

	return fields
}
