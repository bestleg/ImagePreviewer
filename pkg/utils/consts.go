package utils

const (
	Fill      uint8 = 0b01
	Resize    uint8 = 0b11
	WritePerm int   = 600

	SupportedContentTypes = "image/jpeg"
	SupportedHeader       = "HTTP/1.1"

	ErrNotSupportedContentType    = "got not supported content type"
	ErrNotSupportedHeader         = "got not supported header"
	ErrFailedToReadRequestBody    = "failed to read request body"
	ErrFailedToPerformRequest     = "failed to perform request"
	ErrFailedToParseImageURL      = "failed to parse image url"
	ErrFailedToCreateProxyRequest = "failed to create proxy request"
	ErrMakingRequest              = "error making request"
	ErrPrepareRequest             = "failed to prepare request"
)
