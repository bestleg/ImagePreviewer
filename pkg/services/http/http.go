package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	ready     bool
	readyLock sync.RWMutex
)

func setReady(state bool) {
	readyLock.Lock()
	defer readyLock.Unlock()
	ready = state
}

func isReady() bool {
	readyLock.RLock()
	defer readyLock.RUnlock()
	return ready
}

type Server struct {
	Server          *http.Server
	shutdownTimeout time.Duration
	done            chan struct{}
}

type StartStopper interface {
	Start(logger *zap.SugaredLogger)
	Stop(logger *zap.SugaredLogger)
}

func NewHTTPServer(addr string, shutdownTimeout time.Duration, router http.Handler) *Server {
	r := http.NewServeMux()
	r.Handle("/", router)
	r.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if isReady() {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &Server{
		shutdownTimeout: shutdownTimeout,
		Server:          srv,
		done:            make(chan struct{}),
	}
}

func (s *Server) Run(logger *zap.SugaredLogger, appName string) {
	logger.Infof("starting %s", appName)
	s.start(logger)

	setReady(true)
	logger.Infof("%s ready", appName)
	logger.Infof("received %s", wait([]os.Signal{syscall.SIGTERM, syscall.SIGINT}))
	logger.Infof("stopping %s", appName)
	setReady(false)

	s.stop(logger)
}

func wait(signals []os.Signal) os.Signal {
	sig := make(chan os.Signal, len(signals))
	signal.Notify(sig, signals...)
	s := <-sig
	signal.Stop(sig)
	return s
}

func (s *Server) start(logger *zap.SugaredLogger) {
	logger.Infof("starting http server on %s", s.Server.Addr)
	go func() {
		defer close(s.done)

		err := s.Server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("http server failure: %v", err)
		}
		logger.Infof("http server on %s stopped listening", s.Server.Addr)
	}()
}

func (s *Server) stop(logger *zap.SugaredLogger) {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	if err := s.Server.Shutdown(ctx); err != nil {
		logger.Errorf("http shutdown error :%v", err)
	}
	logger.Infof("http server on %s stopped", s.Server.Addr)
	<-s.done
	cancel()
}
