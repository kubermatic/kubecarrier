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

package mockclient

import (
	"context"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is a mock for the controller-runtime dynamic client interface.
type Client struct {
	mock.Mock

	StatusMock *StatusClient
}

var _ client.Client = &Client{}

func NewClient() *Client {
	return &Client{
		StatusMock: &StatusClient{},
	}
}

// StatusClient interface

func (c *Client) Status() client.StatusWriter {
	return c.StatusMock
}

// Reader interface

func (c *Client) Get(ctx context.Context, key types.NamespacedName, obj runtime.Object) error {
	args := c.Called(ctx, key, obj)
	return args.Error(0)
}

func (c *Client) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	args := c.Called(ctx, list, opts)
	return args.Error(0)
}

// Writer interface

func (c *Client) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := c.Called(ctx, obj, patch, opts)
	return args.Error(0)
}

func (c *Client) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

type StatusClient struct {
	mock.Mock
}

var _ client.StatusWriter = &StatusClient{}

func (c *StatusClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *StatusClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := c.Called(ctx, obj, patch, opts)
	return args.Error(0)
}
