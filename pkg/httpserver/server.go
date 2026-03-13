package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gingray/quitedb/pkg/app"
	"github.com/gingray/quitedb/pkg/config"
	"github.com/gingray/quitedb/pkg/lifecycle"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Server struct {
	lifecycle.BaseComponent
	router *gin.Engine
	logger config.Logger
	addr   string
	kafka  *kgo.Client
	ch     chan struct{}
}

func (s *Server) Name() string {
	return "http_server"
}

func NewServer(cfg *config.HTTPServiceConfig, app *app.App) *Server {
	server := &Server{
		router: app.HttpRouter,
		logger: app.Logger,
		addr:   fmt.Sprintf(":%d", cfg.Port),
	}
	server.AddReadyHandler(func(ctx context.Context) error {
		server.ch = make(chan struct{})
		return nil
	})
	return server
}

func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("Starting server")
	errCh := make(chan error)
	go func() {
		s.logger.Info("Server started", "addr", s.addr)

		srv := &http.Server{
			Addr:    s.addr,
			Handler: s.router.Handler(),
		}

		s.AddShutdownHandler(func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		})

		err := srv.ListenAndServe()
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	case <-s.ch:
		return fmt.Errorf("test shutdown")
	}
}
