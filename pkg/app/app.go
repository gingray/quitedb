package app

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gingray/quitedb/pkg/config"
	"github.com/gingray/quitedb/pkg/lifecycle"
)

type App struct {
	lifecycle.BaseComponent
	HttpRouter *gin.Engine
	Logger     config.Logger
}

func (a *App) Name() string {
	return "app"
}

func NewApp(cfg *config.Config) (*App, error) {
	app := &App{}
	err := app.WithLogger()
	if err != nil {
		return nil, err
	}

	err = app.WithHTTPRouter()
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	}
}
