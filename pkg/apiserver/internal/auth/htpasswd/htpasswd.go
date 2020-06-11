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

package htpasswd

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/foomo/htpasswd"
	"github.com/go-logr/logr"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/user"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"github.com/kubermatic/kubecarrier/pkg/apiserver/auth"
	"github.com/kubermatic/kubecarrier/pkg/internal/constants"
)

type HtpasswdAuthenticator struct {
	logr.Logger
	client             client.Client
	htpasswdSecretName string
}

var _ auth.Provider = (*HtpasswdAuthenticator)(nil)
var _ inject.Logger = (*HtpasswdAuthenticator)(nil)
var _ inject.Client = (*HtpasswdAuthenticator)(nil)

func (h *HtpasswdAuthenticator) InjectLogger(l logr.Logger) error {
	h.Logger = l
	return nil
}

func (h *HtpasswdAuthenticator) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

func init() {
	auth.RegisterAuthProvider("Htpasswd", &HtpasswdAuthenticator{})
}

func (h *HtpasswdAuthenticator) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&h.htpasswdSecretName, "htpasswd-secret-name", "ht-secret", "the secret of htpasswd file.")
	fmt.Printf("Add %s", h.htpasswdSecretName)
}

func (h *HtpasswdAuthenticator) Init() error {
	return nil
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (h *HtpasswdAuthenticator) Authenticate(ctx context.Context) (user.Info, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "Basic")
	fmt.Println(token)
	if err != nil {
		h.Error(err, "cannot extract token")
		return nil, err
	}
	username, password, err := h.parseUserInfo(token)
	if err != nil {
		h.Error(err, "cannot parse username and password from authentication token")
		return nil, err
	}
	secret := &corev1.Secret{}
	fmt.Printf("Auth %s", h.htpasswdSecretName)
	if err := h.client.Get(ctx, types.NamespacedName{
		Name:      h.htpasswdSecretName,
		Namespace: constants.KubeCarrierDefaultNamespace,
	}, secret); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "getting htpasswd secret: %s", err.Error())
	}
	htpasswdBytes, ok := secret.Data["auth"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "cannot find auth data in htpasswd secret")
	}
	passwordMap, err := htpasswd.ParseHtpasswd(htpasswdBytes)
	fmt.Println(passwordMap)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "cannot parse htpasswd in bytes")
	}
	fmt.Println(password)
	fmt.Println(passwordMap[username])
	if passwordMap[username] != password {
		return nil, status.Error(codes.Unauthenticated, "username or password doesn't match")
	}
	userInfo := &user.DefaultInfo{
		Name:   username,
		Groups: []string{"kubecarrier:htpasswd"},
	}
	return userInfo, nil
}

func (h *HtpasswdAuthenticator) parseUserInfo(token string) (username string, password string, err error) {
	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return
	}
	fmt.Println(string(data))
	userInfo := strings.Split(string(data), ":")
	if len(userInfo) != 2 {
		err = fmt.Errorf("can not parse username and password")
		return
	}
	fmt.Println(userInfo)
	return userInfo[0], userInfo[1], nil
}
