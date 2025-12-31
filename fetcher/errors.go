package fetcher

import (
	"errors"

	"github.com/r3dpixel/toolkit/trace"
)

// ErrCode Error codes for fetcher
type ErrCode byte

// Error Custom error type for fetcher
type Error = *trace.CodedErr[ErrCode]

// Error codes
const (
	InvalidCredentialsErr ErrCode = iota
	FetchMetadataErr
	MalformedMetadataErr
	FetchCardDataErr
	MalformedCardDataErr
	FetchBookDataErr
	MalformedBookDataErr
	FetchAvatarErr
	DecodeErr
	MissingCookieProviderErr
	None
	errCodeSize
)

// Error messages
var errMessages = [errCodeSize]string{
	InvalidCredentialsErr:    "invalid credentials",
	FetchMetadataErr:         "failed to fetch metadata",
	MalformedMetadataErr:     "malformed metadata response",
	FetchCardDataErr:         "failed to fetch card data",
	MalformedCardDataErr:     "malformed card data response",
	FetchBookDataErr:         "failed to fetch book data",
	MalformedBookDataErr:     "malformed book data response",
	FetchAvatarErr:           "failed to fetch avatar",
	DecodeErr:                "failed to decode card png",
	MissingCookieProviderErr: "missing cookie provider",
	None:                     "",
}

// Ensure error messages are initialized
var _ = errMessages[None]

// String returns the error message
func (e ErrCode) String() string {
	if int(e) < len(errMessages) {
		return errMessages[e]
	}
	return "unknown error"
}

// NewError creates a new error with the given code
func NewError(cause error, code ErrCode) Error {
	return trace.CodedError[ErrCode]().Wrap(cause).Msg(code.String()).Code(code)
}

// GetErrCode returns the error code from the given error
func GetErrCode(err error) ErrCode {
	var codedErr Error
	if errors.As(err, &codedErr) {
		return codedErr.GetCode()
	}
	return None
}
