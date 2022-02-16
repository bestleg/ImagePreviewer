package integrationtests

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const imageType = "image/jpeg"

type TestSuite struct {
	suite.Suite
	client *http.Client
}

func NewTestSuite() *TestSuite {
	return &TestSuite{client: http.DefaultClient}
}

func TestFill(t *testing.T) {
	s := NewTestSuite()

	url := "nginx:80/orig_gopher.jpg"
	width, height := 333, 666

	// nolint:bodyclose
	res, body, err := s.doRequest(t, url, "fill", width, height)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.True(t, res.Header.Get("Content-Type") == imageType)

	config, _, err := image.DecodeConfig(bytes.NewReader(body))
	require.NoError(t, err)

	require.Equal(t, config.Width, width)
	require.Equal(t, config.Height, height)
}

func TestResize(t *testing.T) {
	s := NewTestSuite()

	url := "nginx:80/orig_fox.jpg"
	width, height := 111, 222

	// nolint:bodyclose
	res, body, err := s.doRequest(t, url, "resize", width, height)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.True(t, res.Header.Get("Content-Type") == imageType)

	config, _, err := image.DecodeConfig(bytes.NewReader(body))
	require.NoError(t, err)

	require.Equal(t, config.Width, width)
	require.Equal(t, config.Height, height)
}

func TestBadResize(t *testing.T) {
	s := NewTestSuite()

	url := "nginx:80/orig_fox.jpg"
	width, height := 111111, 2

	// nolint:bodyclose
	res, body, err := s.doRequest(t, url, "resize", width, height)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadGateway, res.StatusCode)
	require.Equal(t, "failed to crop image: jpeg: image is too large to encode\n", string(body))
}

func TestServerDoesntExist(t *testing.T) {
	s := NewTestSuite()

	url := "not_exist.com/orig_gopher.jpg"
	width, height := 333, 666

	// nolint:bodyclose
	res, _, err := s.doRequest(t, url, "fill", width, height)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadGateway, res.StatusCode)
}

func TestCropNotImage(t *testing.T) {
	s := NewTestSuite()

	url := "nginx:80/text.txt"
	width, height := 555, 111

	// nolint:bodyclose
	res, _, err := s.doRequest(t, url, "fill", width, height)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadGateway, res.StatusCode)
}

func TestFillFromCache(t *testing.T) {
	s := NewTestSuite()

	url := "nginx:80/orig_gopher.jpg"
	width, height := 100, 200

	// nolint:bodyclose
	res, body, err := s.doRequest(t, url, "fill", width, height)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.True(t, res.Header.Get("Content-Type") == imageType)

	config, _, err := image.DecodeConfig(bytes.NewReader(body))
	require.NoError(t, err)

	require.Equal(t, config.Width, width)
	require.Equal(t, config.Height, height)

	// nolint:bodyclose
	resFromCache, body, err := s.doRequest(t, url, "fill", width, height)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resFromCache.StatusCode)
	require.True(t, resFromCache.Header.Get("Content-Type") == imageType)

	configFromCache, _, err := image.DecodeConfig(bytes.NewReader(body))
	require.NoError(t, err)

	require.Equal(t, configFromCache.Width, width)
	require.Equal(t, configFromCache.Height, height)
}

// nolint:thelper
func (s TestSuite) doRequest(t *testing.T,
	imageURL string,
	cropType string,
	width, height int,
) (*http.Response, []byte, error) {
	url := "http://image-previewer:8081/" +
		cropType + "/" +
		strconv.FormatInt(int64(width), 10) + "/" +
		strconv.FormatInt(int64(height), 10) + "/" +
		imageURL

	reqReady, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		"http://image-previewer:8081/ready",
		nil)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	require.NoError(t, err)

	res, err := s.client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	resReady, err := s.client.Do(reqReady)
	require.NoError(t, err)
	defer resReady.Body.Close()
	require.Equal(t, http.StatusOK, resReady.StatusCode)

	b, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	return res, b, err
}
