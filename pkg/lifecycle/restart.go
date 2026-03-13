package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

var restartErr = errors.New("restart")

type RestartStrategy struct {
	restartSignalCh chan struct{}
}

func (r *RestartStrategy) Process(ctx context.Context, node *Node) error {
	for {
		restartCtx, cancel := context.WithCancelCause(ctx)
		go func() {
			select {
			case <-r.restartSignalCh:
				cancel(restartErr)
			case <-ctx.Done():
				return
			}
		}()
		err := r.process(restartCtx, node)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !errors.Is(err, restartErr) {
			return err
		}
		err = node.Component.Ready(ctx)
		if err != nil {
			return err
		}
	}
}

func (r *RestartStrategy) process(ctx context.Context, node *Node) error {
	g, errCtx := errgroup.WithContext(ctx)
	wg := sync.WaitGroup{}
	wg.Add(1)
	g.Go(func() error {
		node.logger.Info("supervisor", "status", Run, "component", node.Component.Name())
		wg.Done()
		return node.Component.Run(errCtx)
	})
	wg.Wait()

	for _, node := range node.Nodes {
		g.Go(func() error {
			return node.Run(errCtx)
		})
	}
	err := g.Wait()
	for _, component := range node.Nodes {
		compErr := component.Shutdown(ctx)
		err = errors.Join(err, compErr)
	}
	node.logger.Info("supervisor", "status", ShutdownStart, "component", node.Component.Name())
	stopErr := fmt.Errorf("component: %s, %w", node.Component.Name(), componentStopErr)
	err = errors.Join(err, node.Component.Shutdown(ctx), stopErr)
	node.logger.Info("supervisor", "status", ShutdownFinish, "component", node.Component.Name())
	return err
}
