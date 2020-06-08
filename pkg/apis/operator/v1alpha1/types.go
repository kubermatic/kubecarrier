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

package v1alpha1

import (
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

// ConditionStatus represents a condition's status.
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in
// the condition; "ConditionFalse" means a resource is not in the condition;
// "ConditionUnknown" means Kubernetes can't decide if a resource is in the
// condition or not.
const (
	// ConditionTrue represents the fact that a given condition is true
	ConditionTrue ConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false
	ConditionFalse ConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)

// ObjectReference describes the link to another object in the same namespace
type ObjectReference struct {
	Name string `json:"name"`
}

// CRDReference references a CustomResourceDefitition.
type CRDReference struct {
	Kind    string `json:"kind"`
	Version string `json:"version"`
	Group   string `json:"group"`
	Plural  string `json:"plural"`
}

type APIServerConfig struct {
	// TLSSecretRef references the TLS certificate and private key for serving the KubeCarrier API.
	// +optional
	TLSSecretRef *ObjectReference `json:"tlsSecretRef,omitempty"`
	// OIDC specifies OpenID Connect configuration for API Server authentication
	// +optional
	OIDC *APIServerOIDCConfig `json:"oidc,omitempty"`
}

type APIServerOIDCConfig struct {
	// IssuerURL is the URL the provider signs ID Tokens as. This will be the "iss"
	// field of all tokens produced by the provider and is used for configuration
	// discovery.
	//
	// The URL is usually the provider's URL without a path, for example
	// "https://accounts.google.com" or "https://login.salesforce.com".
	//
	// The provider must implement configuration discovery.
	// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfig
	IssuerURL string `json:"issuerURL"`

	// ClientID the JWT must be issued for, the "sub" field. This plugin only trusts a single
	// client to ensure the plugin can be used with public providers.
	//
	// The plugin supports the "authorized party" OpenID Connect claim, which allows
	// specialized providers to issue tokens to a client for a different client.
	// See: https://openid.net/specs/openid-connect-core-1_0.html#IDToken
	ClientID string `json:"clientID"`

	// APIAudiences are the audiences that the API server identitifes as. The
	// (API audiences unioned with the ClientIDs) should have a non-empty
	// intersection with the request's target audience. This preserves the
	// behavior of the OIDC authenticator pre-introduction of API audiences.
	// +optional
	APIAudiences authenticator.Audiences `json:"apiAudiences,omitempty"`

	// CertificateAuthority references the secret containing issuer's CA in a PEM encoded root certificate of the provider.
	CertificateAuthority ObjectReference `json:"certificateAuthority"`

	// UsernameClaim is the JWT field to use as the user's username.
	// +kubebuilder:default=sub
	// +optional
	UsernameClaim string `json:"usernameClaim"`

	// UsernamePrefix, if specified, causes claims mapping to username to be prefix with
	// the provided value. A value "oidc:" would result in usernames like "oidc:john".
	// +optional
	UsernamePrefix string `json:"usernamePrefix,omitempty"`

	// GroupsClaim, if specified, causes the OIDCAuthenticator to try to populate the user's
	// groups with an ID Token field. If the GroupsClaim field is present in an ID Token the value
	// must be a string or list of strings.
	// +optional
	GroupsClaim string `json:"groupsClaim,omitempty"`

	// GroupsPrefix, if specified, causes claims mapping to group names to be prefixed with the
	// value. A value "oidc:" would result in groups like "oidc:engineering" and "oidc:marketing".
	// +optional
	GroupsPrefix string `json:"groupsPrefix,omitempty"`

	// SupportedSigningAlgs sets the accepted set of JOSE signing algorithms that
	// can be used by the provider to sign tokens.
	//
	// https://tools.ietf.org/html/rfc7518#section-3.1
	//
	// This value defaults to RS256, the value recommended by the OpenID Connect
	// spec:
	//
	// https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation
	// +kubebuilder:default=RS256;
	SupportedSigningAlgs []string `json:"supportedSigningAlgs,omitempty"`

	// RequiredClaims, if specified, causes the OIDCAuthenticator to verify that all the
	// required claims key value pairs are present in the ID Token.
	// +optional
	RequiredClaims map[string]string `json:"requiredClaims,omitempty"`
}
