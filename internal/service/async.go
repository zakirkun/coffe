package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"log/slog"
)

func (a *App) StartPlan(ctx context.Context, rawID string) (domain.ActionPlan, error) {
	id, err := parsePlanID(rawID)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	plan, target, workflow, err := a.startExecution(ctx, id)
	if err != nil {
		return domain.ActionPlan{}, err
	}
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	a.runMu.Lock()
	a.runs[id] = cancel
	a.runWG.Add(1)
	a.runMu.Unlock()
	go func() {
		defer a.runWG.Done()
		defer cancel()
		defer func() {
			a.runMu.Lock()
			delete(a.runs, id)
			a.runMu.Unlock()
		}()
		if _, err := a.runStartedPlan(runCtx, plan, target, workflow); err != nil {
			a.logger.ErrorContext(
				context.WithoutCancel(runCtx), "asynchronous plan failed",
				slog.String("plan_id", string(id)), slog.Any("err", err),
			)
		}
	}()
	return plan, nil
}

func (a *App) CancelPlan(ctx context.Context, rawID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("cancel plan: %w", err)
	}
	id, err := parsePlanID(rawID)
	if err != nil {
		return err
	}
	a.runMu.Lock()
	cancel, ok := a.runs[id]
	a.runMu.Unlock()
	if !ok {
		return fmt.Errorf("plan %s: %w", id, ErrNotRunning)
	}
	cancel()
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.cancelAll()
	return a.Wait(ctx)
}

func (a *App) cancelAll() {
	a.runMu.Lock()
	cancels := make([]context.CancelFunc, 0, len(a.runs)+len(a.agentRuns))
	for _, cancel := range a.runs {
		cancels = append(cancels, cancel)
	}
	for _, cancel := range a.agentRuns {
		cancels = append(cancels, cancel)
	}
	a.runMu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
}

func (a *App) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		a.runWG.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("wait for running plans: %w", ctx.Err())
	}
}
