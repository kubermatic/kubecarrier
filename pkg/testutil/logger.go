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
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/go-logr/logr"
)

// Logger implements logr.Logger and logs to testing.T to preserve the order of log lines in tests.
type Logger struct {
	t      *testing.T
	names  []string
	values map[string]interface{}
}

var _ logr.Logger = (*Logger)(nil)

// NewLogger returns a new Logger flushing to testing.T.
func NewLogger(t *testing.T) *Logger {
	return &Logger{
		t:      t,
		values: map[string]interface{}{},
	}
}

// Info implements logr.Logger.Info
func (l *Logger) Info(msg string, kvs ...interface{}) {
	// marks this function as a helper method, so it will be excluded in the log stacktrace
	l.t.Helper()

	values := addValues(l.values, kvs...)

	j, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}
	l.t.Log(fmt.Sprintf("%-15s %-20s %s", strings.Join(l.names, "."), msg, string(j)))
}

// Error implements logr.Logger.Error
func (l *Logger) Error(err error, msg string, kvs ...interface{}) {
	// marks this function as a helper method, so it will be excluded in the log stacktrace
	l.t.Helper()
	l.Info(msg, append(kvs, "error", err.Error())...)
}

// Enabled implements logr.Logger.Enabled
func (l *Logger) Enabled() bool {
	return true
}

// V implements logr.Logger.V
func (l *Logger) V(level int) logr.InfoLogger {
	return l
}

// WithValues implements logr.Logger.WithValues
func (l *Logger) WithValues(kvs ...interface{}) logr.Logger {
	return &Logger{
		t:      l.t,
		names:  l.names,
		values: addValues(l.values, kvs...),
	}
}

// WithName implements logr.Logger.WithName
func (l *Logger) WithName(name string) logr.Logger {
	return &Logger{
		t:      l.t,
		names:  append(l.names, name),
		values: l.values,
	}
}

func addValues(base map[string]interface{}, kvs ...interface{}) map[string]interface{} {
	values := map[string]interface{}{}
	// add existing k/v pairs
	for k := range base {
		values[k] = base[k]
	}
	// add new k/v pairs
	for i := 0; i < len(kvs); i += 2 {
		values[fmt.Sprint(kvs[i])] = kvs[i+1]
	}
	return values
}
