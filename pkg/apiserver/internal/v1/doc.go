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
	"mime"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "k8c.io/kubecarrier/pkg/apiserver/api/v1"
)

type docServer struct{}

var _ v1.DocServer = (*docServer)(nil)

func NewDocServiceServer() v1.DocServer {
	return &docServer{}
}

func (o docServer) OpenAPI(context.Context, *empty.Empty) (*httpbody.HttpBody, error) {
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

func (o docServer) Swagger(ctx context.Context, req *v1.DocStaticRequest) (*httpbody.HttpBody, error) {
	return o.serveStatic(req.Path)
}

func (o docServer) serveStatic(path string) (*httpbody.HttpBody, error) {
	if path == "" {
		path = "index.html"
	}

	r, err := vfs.Open("/" + path)
	if err != nil {
		if err == os.ErrNotExist {
			return nil, status.Error(codes.NotFound, "file not found")
		}
		return nil, err
	}
	defer r.Close()

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return &httpbody.HttpBody{
		ContentType: mime.TypeByExtension(path),
		Data:        contents,
	}, nil
}
