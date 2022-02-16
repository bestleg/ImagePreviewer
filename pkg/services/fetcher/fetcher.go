package fetcher

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/bestleg/ImagePreviewer/pkg/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Fetcher interface {
	Fetch(ctx context.Context, url string, header http.Header) ([]byte, error)
}

type HTTPFetcher struct {
	logger         *zap.SugaredLogger
	transport      http.RoundTripper
	requestTimeout time.Duration
}

func NewFetcher(l *zap.SugaredLogger, connectTimeout time.Duration, requestTimeout time.Duration) *HTTPFetcher {
	return &HTTPFetcher{
		logger:         l,
		requestTimeout: requestTimeout,
		transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: connectTimeout,
			}).DialContext,
		},
	}
}

func (f HTTPFetcher) Fetch(ctx context.Context, url string, header http.Header) ([]byte, error) {
	proxyRequest, err := prepareRequest(ctx, url, header)
	if err != nil {
		return nil, errors.Wrap(err, utils.ErrPrepareRequest)
	}
	responseBody, err := f.doRequest(proxyRequest)
	if err != nil {
		return nil, errors.Wrap(err, utils.ErrMakingRequest)
	}
	return responseBody, nil
}

func prepareRequest(ctx context.Context, rawURL string, header http.Header) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, utils.ErrFailedToCreateProxyRequest)
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrap(err, utils.ErrFailedToParseImageURL)
	}
	if parsedURL.Scheme != "http" {
		return nil, errors.New(utils.ErrNotSupportedScheme)
	}
	request.URL = parsedURL
	request.Header = header
	return request, nil
}

func (f *HTTPFetcher) doRequest(request *http.Request) ([]byte, error) {
	client := http.Client{
		Timeout:   f.requestTimeout,
		Transport: f.transport,
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, utils.ErrFailedToPerformRequest)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			f.logger.Errorf("failed to close body %v", errClose)
		}
	}()

	if !utils.Contains([]string{utils.SupportedContentTypes}, resp.Header.Get("Content-type")) {
		return nil, errors.New(utils.ErrNotSupportedContentType)
	}

	if resp.Proto != utils.SupportedHeader {
		return nil, errors.New(utils.ErrNotSupportedHeader)
	}

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, utils.ErrFailedToReadRequestBody)
	}
	return buff, nil
}
