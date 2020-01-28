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

package util

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type PredicateFn func(obj runtime.Object) bool

func (p PredicateFn) Create(ev event.CreateEvent) bool {
	return p(ev.Object)
}

func (p PredicateFn) Delete(ev event.DeleteEvent) bool {
	return p(ev.Object)
}

func (p PredicateFn) Update(ev event.UpdateEvent) bool {
	return p(ev.ObjectNew)
}

func (p PredicateFn) Generic(ev event.GenericEvent) bool {
	return p(ev.Object)
}

var _ predicate.Predicate = (PredicateFn)(nil)
