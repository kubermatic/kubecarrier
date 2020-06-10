/*
Copyright 2019 The KubeCarrier Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testutil

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	corezap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

// NewLogger returns a new Logger flushing to testing.T.
func NewLogger(t *testing.T) logr.Logger {
	return zapr.NewLogger(NewZapLogger(t))
}

func NewZapLogger(t *testing.T) *zap.Logger {
	return corezap.NewRaw(func(options *corezap.Options) {
		options.Development = true
		if t != nil {
			options.DestWritter = &logWriter{T: t}
		}
		if level := os.Getenv("TEST_LOG_LEVEL"); level != "" {
			l := zap.NewAtomicLevel()
			if err := l.UnmarshalText([]byte(level)); err != nil {
				panic(err)
			}
			options.Level = &l
		}
	})
}

func LoggingContext(t *testing.T, ctx context.Context) (context.Context, *zap.Logger) {
	logger := NewZapLogger(t)
	ctx = util.InjectLogger(ctx, logger)
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	return ctx, logger
}

type logWriter struct {
	*testing.T
}

var _ io.Writer = (*logWriter)(nil)

func (l logWriter) Write(p []byte) (n int, err error) {
	l.T.Log(string(p))
	return len(p), nil
}
