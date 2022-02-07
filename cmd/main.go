package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/bestleg/ImagePreviewer/pkg/logging"
	transformerPkg "github.com/bestleg/ImagePreviewer/pkg/services/cropper"
	fetcherPkg "github.com/bestleg/ImagePreviewer/pkg/services/fetcher"
	"github.com/bestleg/ImagePreviewer/pkg/services/http"
	"github.com/bestleg/ImagePreviewer/pkg/services/processor"
	lru "github.com/hashicorp/golang-lru"
)

var (
	appName         = "image-previewer"
	addr            string
	connectTimeout  time.Duration
	requestTimeout  time.Duration
	shutdownTimeout time.Duration
	cacheDir        string
	cacheSize       int
)

func init() {
	flag.StringVar(&addr, "addr", ":8081", "App addr")
	flag.DurationVar(&connectTimeout, "connect-timeout", 25*time.Second, "Сonnection timeout")
	flag.DurationVar(&requestTimeout, "request-timeout", 25*time.Second, "Request timeout")
	flag.DurationVar(&shutdownTimeout, "shutdown-timeout", 30*time.Second, "Graceful shutdown timeout")
	flag.StringVar(&cacheDir, "cache-dir", "", "Path to Cache dir")
	flag.IntVar(&cacheSize, "cache-size", 5, "Size of cache")
}

func main() {
	ctx := context.Background()
	flag.Parse()

	logger, err := logging.InitLogger()
	if err != nil {
		log.Fatal(fmt.Sprintf("err to init logger %v", err))
	}
	fetcher := fetcherPkg.NewFetcher(logger, connectTimeout, requestTimeout)
	cropper := transformerPkg.NewCropper()

	if cacheDir == "" {
		var err error
		cacheDir, err = ioutil.TempDir("", "")
		if err != nil {
			logger.Fatalf("error to init cache: %v", err)
		}
		defer func() {
			if err := os.RemoveAll(cacheDir); err != nil {
				logger.Errorf("failed to remove cache dir: %v", err)
			}
		}()
	}

	cache, err := lru.NewWithEvict(cacheSize, func(key interface{}, value interface{}) {
		if path, ok := value.(string); ok {
			defer func() {
				if err := os.Remove(path); err != nil {
					logger.Fatalf("failed to remove item from cache: %v", err)
				}
			}()
		}
	})
	if err != nil {
		logger.Fatalf("failed to setup cache %v", err)
	}

	processor := processor.NewProcessor(cacheDir, logger, fetcher, cropper, cache)
	handlerWithGz := gziphandler.GzipHandler(processor.ProcessorHandler(ctx))
	middleWareLoggerHandler := logging.MiddleWareLogger(logger)
	server := http.NewHTTPServer(addr, shutdownTimeout, middleWareLoggerHandler(handlerWithGz))

	server.Run(logger, appName)
}
