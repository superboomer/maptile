package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/superboomer/maptile/app/options"
	"github.com/superboomer/maptile/app/server/api"
	"github.com/superboomer/maptile/app/server/middleware"
	"go.uber.org/zap"
)

// Server main app struct that represents http server
type Server struct {
	options *options.Opts
	logger  *zap.Logger
	server  *http.Server
}

// NewServer create new Server
func NewServer(opts *options.Opts, logger *zap.Logger) *Server {
	return &Server{
		options: opts,
		logger:  logger,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%s", opts.APIPort),
			Handler:      nil,
			ReadTimeout:  10 * time.Second, // Set a maximum time to read the request headers
			WriteTimeout: 10 * time.Second, // Set a maximum time to write the response
		},
	}
}

// Run start program with specified parameters
func Run(ctx context.Context, logger *zap.Logger, opts *options.Opts) error {

	apiService, err := api.CreateAPI(logger, &opts.Cache, opts.Schema, opts.MaxSide)
	if err != nil {
		return err
	}

	var s = NewServer(opts, logger)
	var md = &middleware.MD{Logger: logger}

	s.logger.Info("service starting")

	s.SetRoutes(apiService, md)
	err = s.RunHTTP(ctx)
	if err != nil {
		s.logger.Error("http server error occurred while running", zap.Error(err))
	}

	s.logger.Info("service stopped")

	return err
}

// RunHTTP start listen and server http server
func (s *Server) RunHTTP(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down the HTTP server...", zap.String("port", s.options.APIPort))
		_ = s.server.Shutdown(ctx)
	}()

	s.logger.Info("Starting the HTTP server...", zap.String("port", s.options.APIPort))
	err := s.server.ListenAndServe()

	// Shutting down the server is not something bad ffs Go...
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
