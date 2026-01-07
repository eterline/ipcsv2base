// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.

package toolkit

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type AppStarter struct {
	Context  context.Context
	stopFunc context.CancelFunc
	startAt  time.Time
	wTimer   WorkTimerCallback
	wg       sync.WaitGroup
}

// StopApp – cancel root app context and stopping app
func (s *AppStarter) StopApp() {
	if s.Context.Err() == nil {
		s.stopFunc()
	}
}

// MustStopApp – cancel root app context and extremely stops app with exit code
func (s *AppStarter) MustStopApp(exitCode int) {
	s.StopApp()
	<-s.Context.Done()
	os.Exit(exitCode)
}

// Wait – waiting for app root context done
func (s *AppStarter) Wait() {
	<-s.Context.Done()
}

// AddWorker – appends worker
func (s *AppStarter) AddWorker() {
	s.wg.Add(1)
}

// DoneWorker – final worker
func (s *AppStarter) DoneWorker() {
	s.wg.Done()
}

// WrapWorker – wrap worker into root workers
func (s *AppStarter) WrapWorker(f func()) {
	s.AddWorker()
	go func() {
		defer s.DoneWorker()
		f()
	}()
}

// WaitWorkers – wait for workers final or timeout exit
func (s *AppStarter) WaitWorkers(timeout time.Duration) error {

	s.Wait()

	ctx, stop := context.WithTimeout(context.Background(), timeout)
	defer stop()

	go func() {
		s.wg.Wait()
		stop()
	}()

	<-ctx.Done()

	if err := ctx.Err(); err == context.DeadlineExceeded {
		return fmt.Errorf("app workers stop timeout: %w", err)
	}

	return nil
}

// ============================

// FinalThreads – wait for thread final or timeout exit
func (s *AppStarter) WorkTime() time.Duration {
	return s.wTimer()
}

// AddValue – appends to context values with key
func (s *AppStarter) AddValue(key, value any) {
	s.Context = context.WithValue(s.Context, key, value)
}

func (s *AppStarter) Started() time.Time {
	return s.startAt
}

// UseContextAdders – uses func list that returns new context
func (s *AppStarter) UseContextAdders(
	addFunc ...func(context.Context) context.Context,
) {
	for _, add := range addFunc {
		s.Context = add(s.Context)
	}
}

// InitAppStart – create app root context and stop function object
func InitAppStart(preInitFunc func() error) *AppStarter {
	return InitAppStartWithContext(
		context.Background(),
		preInitFunc,
	)
}

// InitAppStartWithContext – create app root context and stop function object form external context.
// Must be used with pre init function. If their init will be errored – panic closes app
func InitAppStartWithContext(ctx context.Context, preInitFunc func() error) *AppStarter {

	start := time.Now()

	if err := preInitFunc(); err != nil {
		fmt.Printf("app starting fatal error: %v\n", err)
		os.Exit(1)
	}

	rootContext, stopFunc := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)

	return &AppStarter{
		Context:  rootContext,
		stopFunc: stopFunc,
		wTimer:   WorkTimer(start),
		startAt:  start,
	}
}

// =================================

// BytesUUID – generates uuid (v5) from bytes array
func BytesUUID(input []byte) (uuid.UUID, bool) {
	hash := sha1.New()

	_, err := hash.Write(input)
	if err != nil {
		return uuid.Nil, false
	}

	id, err := uuid.FromBytes(hash.Sum(nil)[:16])
	if err != nil {
		return uuid.Nil, false
	}

	return id, true
}

// StringUUID – generates uuid (v5) from string value
func StringUUID(input string) (uuid.UUID, bool) {
	return BytesUUID([]byte(input))
}

// BytesUUID – generates uuid (v5) from object
func ObjectUUID(object any) (uuid.UUID, bool) {
	data, err := json.Marshal(object)
	if err != nil {
		return uuid.Nil, false
	}

	return BytesUUID(data)
}

// =================================

type WorkTimerCallback func() time.Duration

func WorkTimer(start time.Time) WorkTimerCallback {
	return func() time.Duration {
		return time.Since(start)
	}
}
