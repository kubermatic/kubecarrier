/*
Copyright 2019 The Kubecarrier Authors.

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
	"os"

	zapcore "go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kubermatic/kubecarrier/pkg/anchor"
	"github.com/kubermatic/kubecarrier/pkg/anchor/cmd"
)

func main() {
	ctrl.SetLogger(zap.New(func(options *zap.Options) {
		l := zapcore.NewAtomicLevelAt(zapcore.DebugLevel)
		options.Level = &l
		options.Development = true
	}))

	log := ctrl.Log.WithName("anchor")
	log.Info("Starting anchor command")

	if err := anchor.NewAnchor(log, cmd.DefaultStreams()).Execute(); err != nil {
		log.Error(err, "cannot perform required action")
		os.Exit(1)
	}
}
