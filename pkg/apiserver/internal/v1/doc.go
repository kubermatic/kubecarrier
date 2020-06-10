/*
Copyright 2020 The KubeCarrier Authors.

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

package v1

import (
	"context"
	"io/ioutil"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/genproto/googleapis/api/httpbody"

	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

type docServer struct {
}

var _ v1.DocServer = (*docServer)(nil)

func (o docServer) Swagger(context.Context, *empty.Empty) (*httpbody.HttpBody, error) {
	r, err := vfs.Open("/apidocs.swagger.json")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &httpbody.HttpBody{
		ContentType: "application/json",
		Data:        contents,
	}, nil
}

func (o docServer) OpenAPI(context.Context, *empty.Empty) (*httpbody.HttpBody, error) {
	r, err := vfs.Open("/_index.md")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &httpbody.HttpBody{
		ContentType: "text/markdown",
		Data:        contents,
	}, nil
}

func NewDocServiceServer() v1.DocServer {
	return &docServer{}
}
