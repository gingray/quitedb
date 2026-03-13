package test

import (
	"context"
	"fmt"
	"io"
	net_http "net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gingray/quitedb/internal/http"
	"github.com/gingray/quitedb/pkg/app"
	"github.com/gingray/quitedb/pkg/config"
	"github.com/gingray/quitedb/pkg/httpserver"
	"github.com/gingray/quitedb/pkg/lifecycle"
	"github.com/stretchr/testify/suite"
)

type QuiteDbTestSuite struct {
	suite.Suite
	serverUrl string
	cancel    context.CancelFunc
}

func (suite *QuiteDbTestSuite) SetupSuite() {
	cfg, err := config.NewConfig()
	cfg.Port = 9191
	suite.Assertions.NotNil(cfg)
	suite.Assertions.NoError(err)

	appClient, err := app.NewApp(cfg)
	suite.Assertions.NoError(err)
	suite.Assertions.NotNil(appClient)
	suite.serverUrl = fmt.Sprintf("http://localhost:%d", cfg.Port)
	server := httpserver.NewServer(&cfg.HTTPServiceConfig, appClient)
	router := http.NewRouter(appClient.Db)
	router.SetupRoutes(appClient.HttpRouter)

	supervisor := lifecycle.NewSupervisor(appClient.Logger)
	rootNode := supervisor.CreateRootNode()
	appNode := supervisor.CreateNode(appClient)
	serverNode := supervisor.CreateNode(server)

	rootNode.AddNode(appNode)
	appNode.AddNode(serverNode)
	ctx, cancel := context.WithCancel(context.Background())
	suite.cancel = cancel
	timer := time.NewTimer(1 * time.Second)
	go func() {
		err = rootNode.Run(ctx)
	}()
	<-timer.C
}

func (suite *QuiteDbTestSuite) TearDownSuite() {
	suite.cancel()
}

func (suite *QuiteDbTestSuite) TestBaseActionLog() {
	readyEndpoint := fmt.Sprintf("%s/ready", suite.serverUrl)
	resp, err := net_http.Get(readyEndpoint)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	suite.Assertions.NoError(err)
	suite.Assertions.Equal(200, resp.StatusCode)

	actions, err := suite.loadActionLines()
	suite.Assertions.NoError(err)
	suite.Assertions.NotEmpty(actions)

	for _, action := range actions {
		body, err := suite.makeRequest(action)
		suite.Assertions.NoError(err)
		if action.Action == "PUT" {
			continue
		}
		if action.IsNotFound {
			suite.Assertions.Equal(NotFound, body)
		} else {
			result, err := strconv.ParseInt(body, 10, 64)
			suite.Assertions.NoError(err)
			suite.Assertions.Equal(action.Value, int(result))
			suite.Assertions.Equal(200, resp.StatusCode)
		}
	}
}

func TestQuiteDbTestSuite(t *testing.T) {
	suite.Run(t, new(QuiteDbTestSuite))
}
