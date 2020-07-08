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
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kubermatic/utils/pkg/testutil"
)

type testUser struct {
	username, password string
}

func TestVerifyPassword(t *testing.T) {
	testScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(testScheme))

	users := []testUser{
		{"user1", "mickey5"},
		{"user2", "alexandrew"},
		{"user3", "hawaiicats78"},
		{"user4", "DIENOW"},
		{"user5", "e8f685"},
	}

	tests := []struct {
		name         string
		htpasswdData string
	}{
		{
			name: "Plain",
			htpasswdData: `user1:mickey5
user2:alexandrew
user3:hawaiicats78
user4:DIENOW
user5:e8f685`,
		},
		{
			name: "MD5",
			htpasswdData: `user1:$apr1$gxNb79DX$6wi9QaGNM5TA0kBKiC4710
user2:$apr1$kv1uUfCO$iEwrWojf92uZ/9uhTQmMo.
user3:$apr1$UQ6GxE7V$OrIqWONGuSV9RfS3B2dfO1
user4:$apr1$OZ.RwYJH$AwfW2h0gJnu2fQi0GegVe1
user5:$apr1$9r9GyMpL$3IiaLNos/tbouLJwsW8ey/`,
		},
		{
			name: "SHA",
			htpasswdData: `user1:{SHA}D9rQ8iK6feNAniulHNKdr5V38ok=
user2:{SHA}KS7VQqgAnMUfXgWmFCCa6DVhY+M=
user3:{SHA}mzD9ouM0P06arY0Obdb2KojkFeY=
user4:{SHA}2HrOk971ockoAr1Ct1o7GpvFLdU=
user5:{SHA}IyrjpSzIjrlLT7KjVh1q1LBDCFA=`,
		},
		{
			name: "Bcrypt",
			htpasswdData: `user1:$2y$05$fpu.jNd5fPlx3ggfZ2BWR.Wc3/hc7ke7LsIpwZM6/e0B6VniqFRIW
user2:$2y$05$4QmbRfzXERVFyLbUdtCd8ekz1pAfNB5ZsmXevnKgSMc3XHqDYm2wa
user3:$2y$05$.V03HbzL5HAdwq8DYbt/JOVi/crBiqSXvsgNLHucGLLBpApHjK0Di
user4:$2y$05$/jwDvqAoKjNWwRpUzyLvcuhcSloP9tjxAlPfAUlVvVtmMpBPEC9s2
user5:$2y$05$yVjPeTy8/FIUZAJWSSmnAO7GsWHFA2jVeBWFF6Y6RoWEpoxGxtFzS`,
		},
		{
			name: "Ssha",
			htpasswdData: `user1:{SSHA}KHQzbbDgqjRkfd7li1NBL7kI0D5oMzBM
user2:{SSHA}OFxiAyw1TSNiyGybLyjVg+yewdhoMzBM
user3:{SSHA}FeUKUVnpp9IlolmuqfIMUYaa3/doMzBM
user4:{SSHA}mfDeU9QRfvED1gfBExrJgDsi74xoMzBM
user5:{SSHA}Y3eY1xbgHUOFKOPoiLwluYlsd3FoMzBM`,
		},
		{
			name: "Md5Crypt",
			htpasswdData: `user1:$1$D89ubl/e$dJ8XW4DfrJHTrnwCdx3Ji1
user2:$1$D89ubl/e$xuQ74IxhM3J10sv0QHVgA/
user3:$1$D89ubl/e$Y07COBJSUbNDlYlFyRYUp.
user4:$1$D89ubl/e$4IZ.tBiqvtxt7Dpt1MkgE1
user5:$1$D89ubl/e$mLrBtDw8UTdAX7jDZLQIB0`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			secret := &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      "htpasswd",
					Namespace: "htpasswd",
				},
				Data: map[string][]byte{
					"auth": []byte(test.htpasswdData),
				},
			}
			client := fakeclient.NewFakeClientWithScheme(testScheme, secret)
			htpasswdAuthenticator := HtpasswdAuthenticator{
				htpasswdSecretName: secret.Name,
				Logger:             testutil.NewLogger(t),
				client:             client,
			}
			for _, user := range users {
				assert.NoError(t, htpasswdAuthenticator.verifyPassword(user.username, user.password, secret))
				assert.Error(t, htpasswdAuthenticator.verifyPassword(user.username, user.password+"test", secret))
			}
		})
	}
}
