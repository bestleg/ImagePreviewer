package processor

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"

	"github.com/bestleg/ImagePreviewer/pkg/services/cropper"
	"github.com/bestleg/ImagePreviewer/pkg/services/fetcher"
	"github.com/bestleg/ImagePreviewer/pkg/utils"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Processor struct {
	cacheDir string
	logger   *zap.SugaredLogger
	fetcher  fetcher.Fetcher
	cropper  cropper.Transformer
	cache    simplelru.LRUCache
}

func NewProcessor(
	cacheDir string,
	l *zap.SugaredLogger,
	f fetcher.Fetcher,
	t cropper.Transformer,
	c simplelru.LRUCache,
) *Processor {
	return &Processor{cacheDir: cacheDir, logger: l, fetcher: f, cropper: t, cache: c}
}

func (p *Processor) ProcessorHandler(ctx context.Context) http.Handler {
	r := httprouter.New()
	crop := uint8(0)
	r.GET("/:cropFormat/:width/:height/*url", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		cropFormat := ps.ByName("cropFormat")
		rawWidth := ps.ByName("width")
		rawHeight := ps.ByName("height")
		url := ps.ByName("url")
		url = "http://" + url[1:]
		p.logger.Infow("app request",
			"url", url,
			"width", rawWidth,
			"height", rawHeight,
			"headers", r.Header,
		)
		switch cropFormat {
		case "fill":
			crop = utils.Fill
		case "resize":
			crop = utils.Resize
		default:
			{
				p.logger.Errorf("wrong type of crop image: %s", cropFormat)
				http.Error(w, fmt.Sprintf("wrong type of crop image: %s", cropFormat), http.StatusBadRequest)
				return
			}
		}
		width, err := strconv.Atoi(rawWidth)
		if err != nil {
			p.logger.Errorf("failed to parse width: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		height, err := strconv.Atoi(rawHeight)
		if err != nil {
			p.logger.Errorf("failed to parse height: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		img, err := p.process(ctx, url, r.Header, width, height, crop)
		if err != nil {
			p.logger.Errorf("failed to handle request: %v", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Add("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(img)))

		if _, err := w.Write(img); err != nil {
			p.logger.Errorf("failed to write response: %v", err)
		}
	})
	return r
}

func (p *Processor) process(
	ctx context.Context,
	url string,
	header http.Header,
	width, height int,
	cropFormat uint8,
) ([]byte, error) {
	cacheKey, err := utils.GetHash(fmt.Sprintf("%s|%d|%d|%d", url, width, height, cropFormat))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cacheKey hash")
	}
	if imgPath, found := p.cache.Get(cacheKey); found {
		img, err := ioutil.ReadFile(imgPath.(string))
		return img, err
	}

	img, err := p.fetcher.Fetch(ctx, url, header)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch image")
	}

	img, err = p.cropper.Crop(img, width, height, cropFormat)
	if err != nil {
		return nil, errors.Wrap(err, "failed to crop image")
	}

	imgPath := path.Join(p.cacheDir, cacheKey+".jpeg")
	err = ioutil.WriteFile(imgPath, img, fs.FileMode(utils.WritePerm))
	if err != nil {
		return nil, errors.Wrap(err, "failed to save image")
	}

	p.cache.Add(cacheKey, imgPath)

	return img, nil
}
