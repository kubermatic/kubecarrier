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

package deletecmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8c.io/utils/pkg/util"

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
)

func newAccountCommand(log logr.Logger, cl *util.ClientWatcher) *cobra.Command {
	var (
		force bool
		all   bool
	)
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(1),
		Use:   "account",
		Short: "account NAME [--all] [--force]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if all && len(args) > 0 {
				return fmt.Errorf("either --all or specific account must be specified")
			}
			if !all && len(args) == 0 {
				return fmt.Errorf("one of --all or specific account must be specified")
			}

			ctx := context.Background()

			if all {
				list := &catalogv1alpha1.AccountList{}
				if err := cl.List(ctx, list); err != nil {
					return err
				}
				for _, it := range list.Items {
					args = append(args, it.Name)
				}
			}
			for _, it := range args {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "deleting account", "name", it)
				account := &catalogv1alpha1.Account{}
				if err := client.IgnoreNotFound(cl.Get(ctx, types.NamespacedName{Name: it}, account)); err != nil {
					return fmt.Errorf("getting account %s: %w", it, err)
				}
				err := client.IgnoreNotFound(cl.Delete(ctx, account))
				if err != nil {
					if !force || !strings.HasPrefix(err.Error(), "admission webhook \"vaccount.kubecarrier.io\" denied the request: deletion blocking objects found") {
						return fmt.Errorf("deleting account %s : %w", account.Name, err)
					}

					for _, objKind := range []runtime.Object{
						&catalogv1alpha1.CatalogEntrySetList{},
						&catalogv1alpha1.CatalogEntryList{},
						&catalogv1alpha1.DerivedCustomResourceList{},
						&corev1alpha1.CustomResourceDiscoverySetList{},
						&corev1alpha1.CustomResourceDiscoveryList{},
						&corev1alpha1.ServiceClusterAssignmentList{},
					} {
						if err := cl.List(ctx, objKind, client.InNamespace(account.Status.Namespace.Name)); err != nil {
							return err
						}
						objs, err := meta.ExtractList(objKind)
						if err != nil {
							return err
						}
						for _, obj := range objs {
							_, _ = fmt.Fprintln(cmd.OutOrStdout(), "deleting", util.MustLogLine(obj, scheme))
							if err := client.IgnoreNotFound(cl.Delete(ctx, obj)); err != nil {
								return err
							}
						}
						for _, obj := range objs {
							if err := cl.WaitUntilNotFound(ctx, obj); err != nil {
								return err
							}
						}
					}
					if err := client.IgnoreNotFound(cl.Delete(ctx, account)); err != nil {
						return err
					}
				}
				if err := cl.WaitUntilNotFound(ctx, account); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "force delete the account")
	cmd.Flags().BoolVar(&all, "all", false, "delete all accounts")
	return cmd
}
