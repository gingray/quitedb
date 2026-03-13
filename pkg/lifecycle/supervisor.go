package lifecycle

import (
	"context"
	"errors"

	"github.com/gingray/quitedb/pkg/config"
)

const (
	ReadyCheckStart  = "ready-check-start"
	ReadyCheckFinish = "ready-check-finish"
	Run              = "running"
	ShutdownStart    = "shutdown-start"
	ShutdownFinish   = "shutdown-finish"
)

var componentStopErr = errors.New("component stop")

type strategy interface {
	Process(ctx context.Context, baseNode *Node) error
}

type Supervisor struct {
	logger config.Logger
}

func NewSupervisor(logger config.Logger) *Supervisor {
	return &Supervisor{logger: logger}
}

func (n *Node) AddNode(node *Node) {
	n.Nodes = append(n.Nodes, node)
}

func (n *Node) Shutdown(ctx context.Context) error {
	for _, node := range n.Nodes {
		err := node.Shutdown(ctx)
		if err != nil {
			return err
		}
	}
	return n.Component.Shutdown(ctx)
}

type Node struct {
	Component Component
	Nodes     []*Node
	logger    config.Logger
	strategy  strategy
}

func (n *Node) Name() string {
	return n.Component.Name()
}

func (n *Node) Ready(ctx context.Context) error {
	return n.Component.Ready(ctx)
}

func (s *Supervisor) CreateRootNode() *Node {
	return &Node{Component: NewRootComponent(), Nodes: []*Node{}, logger: s.logger, strategy: &DefaultStrategy{}}
}

func (s *Supervisor) CreateNode(component Component) *Node {
	return &Node{Component: component, Nodes: []*Node{}, logger: s.logger, strategy: &DefaultStrategy{}}
}

func (n *Node) Run(ctx context.Context) error {
	n.logger.Info("supervisor", "status", ReadyCheckStart, "component", n.Component.Name())
	err := n.Component.Ready(ctx)
	n.logger.Info("supervisor", "status", ReadyCheckFinish, "component", n.Component.Name())

	if err != nil {
		return err
	}
	return n.strategy.Process(ctx, n)
}
