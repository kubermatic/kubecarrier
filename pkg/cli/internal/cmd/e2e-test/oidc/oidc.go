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

package oidc

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/json"
)

func NewOIDCCommand(log logr.Logger) *cobra.Command {
	var (
		vaultKey = "dev/e2e-dex"
	)

	cmd := &cobra.Command{
		Use:   "oidc",
		Short: "oidc helper to get the right barer token for account",
		RunE: func(cmd *cobra.Command, args []string) error {
			vCmd := exec.Command("vault", "kv", "get", "-format=json", vaultKey)
			out, err := vCmd.CombinedOutput()
			if err != nil {
				log.Error(err, string(out))
				return err
			}
			data := make(map[string]interface{})
			if err := json.Unmarshal(out, &data); err != nil {
				return fmt.Errorf("cannot unmarshall: %w", err)
			}
			creds := data["data"].(map[string]interface{})["data"].(map[string]interface{})
			caCert, err := base64.StdEncoding.DecodeString(creds["caCert"].(string))
			if err != nil {
				return err
			}
			clientCert, err := base64.StdEncoding.DecodeString(creds["clientCert"].(string))
			if err != nil {
				return err
			}
			clientKey, err := base64.StdEncoding.DecodeString(creds["clientKey"].(string))
			if err != nil {
				return err
			}
			username := creds["username"].(string)
			password := creds["password"].(string)

			_ = caCert
			_ = clientCert
			_ = clientKey
			_ = username
			_ = password
			ctx := context.Background()

			log.Info("successfully fetched data from vault", "username", username)
			// found these constants by digging through kubermatic repo
			// api/hack/ci/testdata/oauth_values.yaml
			cl := NewClient("kubermatic", "http://localhost:8000", "https://dev.kubermatic.io/dex/auth", log)
			token, err := cl.Login(ctx, username, password)
			if err != nil {
				return err
			}
			log.Info("got token", "token", token)
			return nil
		},
	}
	cmd.Flags().StringVar(&vaultKey, "vault-key", vaultKey, "vaulg key to search for credentials")
	return cmd
}
