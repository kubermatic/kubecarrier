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

package main

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	corezap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kubermatic/kubecarrier/pkg/anchor"
)

func main() {
	level := zap.NewAtomicLevel()
	ctrl.SetLogger(corezap.New(func(options *corezap.Options) {
		options.Level = &level
		options.Development = true
	}))

	log := ctrl.Log.WithName("anchor")
	if err := anchor.NewAnchor(&logWrapper{Logger: log, level: &level}).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type logWrapper struct {
	logr.Logger
	level *zap.AtomicLevel
}

func (w *logWrapper) GetLevel() *zap.AtomicLevel {
	return w.level
}
