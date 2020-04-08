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

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

func NewOIDCCommand(log logr.Logger) *cobra.Command {
	var (
		vaultKey = "dev/e2e-dex"
	)

	cmd := &cobra.Command{
		Use:   "oidc",
		Short: "oidc helper to get the right barer token for account",
		RunE: func(cmd *cobra.Command, args []string) error {
			//vCmd := exec.Command("vault", "kv", "get", "-format=json", vaultKey)
			//out, err := vCmd.CombinedOutput()
			//if err != nil {
			//	log.Error(err, string(out))
			//	return err
			//}
			//data := make(map[string]interface{})
			//if err := json.Unmarshal(out, &data); err != nil {
			//	return fmt.Errorf("cannot unmarshall: %w", err)
			//}
			//creds := data["data"].(map[string]interface{})["data"].(map[string]interface{})
			//
			//username := creds["username"].(string)
			//password := creds["password"].(string)
			username := "admin@example.com"
			password := "password"
			redirectURI := "http://192.168.42.219:31850/oauth2/callback"
			providerURL := "http://127.0.0.1:8080/auth"
			clientID := "e2e-client-id"

			// found these constants by digging through kubermatic repo
			// api/hack/ci/testdata/oauth_values.yaml
			// redirectURI :=  "http://localhost:8080"
			// providerURL := "https://dev.kubermatic.io/dex/auth"
			// clientID := kubermatic

			ctx := context.Background()

			log.Info("successfully fetched data from vault", "username", username)
			cl := NewClient(clientID, redirectURI, providerURL, log)
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
