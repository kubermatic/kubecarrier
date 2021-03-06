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
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8c.io/kubecarrier/pkg/apiserver/auth"
)

type Authentication []AuthenticationConfig

func (a Authentication) Validate() error {
	if len(a) == 0 {
		return errors.New("at least one authentication configuration should be specified")
	}
	authEnabled := map[string]bool{}
	for _, config := range a {
		if err := config.Validate(); err != nil {
			return err
		}
		provider := config.GetEnabledProvider()
		if authEnabled[provider] {
			return fmt.Errorf("Duplicate %s configuration", provider)
		}
		authEnabled[provider] = true
	}
	return nil
}

func (a *Authentication) Default() bool {
	if len(*a) == 0 {
		*a = append(*a, AuthenticationConfig{ServiceAccount: &ServiceAccount{}})
		*a = append(*a, AuthenticationConfig{Anonymous: &Anonymous{}})
		return true
	}
	return false
}

// APIServerSpec defines the desired state of APIServer
type APIServerSpec struct {
	// TLSSecretRef references the TLS certificate and private key for serving the KubeCarrier API.
	// +optional
	TLSSecretRef *ObjectReference `json:"tlsSecretRef,omitempty"`
	// +optional
	// Authentication configuration
	Authentication Authentication `json:"authentication,omitempty"`
	// Paused tell controller to pause reconciliation process and assume that APIServer is ready
	// +optional
	Paused PausedFlagType `json:"paused,omitempty"`
	// LogLevel
	// +optional
	LogLevel *int `json:"logLevel,omitempty"`
}

func (a *APIServerSpec) SetLogLevel(logLevel int) {
	a.LogLevel = &logLevel
}

func (a APIServerSpec) Validate() error {
	return a.Authentication.Validate()
}

func (a *APIServerSpec) Default() bool {
	return a.Authentication.Default()
}

type AuthenticationConfig struct {
	// OIDC specifies OpenID Connect configuration for API Server authentication
	// +optional
	OIDC *APIServerOIDCConfig `json:"oidc,omitempty"`
	// StaticUsers specifies static users configuration for API Server authentication
	// +optional
	StaticUsers *StaticUsers `json:"staticUsers,omitempty"`
	// ServiceAccount specifies whether service account auth provider enabled
	// +optional
	ServiceAccount *ServiceAccount `json:"serviceAccount,omitempty"`
	// Anonymous specifies whether anonymous auth provider enabled
	// +optional
	Anonymous *Anonymous `json:"anonymous,omitempty"`
}

type ServiceAccount struct{}
type Anonymous struct{}

func (s AuthenticationConfig) GetEnabledProvider() string {
	if s.OIDC != nil {
		return auth.ProviderOIDC
	}
	if s.StaticUsers != nil {
		return auth.ProviderHtpasswd
	}
	if s.ServiceAccount != nil {
		return auth.ProviderToken
	}
	if s.Anonymous != nil {
		return auth.ProviderAnynymous
	}
	return ""
}

func (s AuthenticationConfig) Validate() error {
	var enabled int
	if s.OIDC != nil {
		enabled++
	}
	if s.StaticUsers != nil {
		enabled++
	}
	if s.ServiceAccount != nil {
		enabled++
	}
	if s.Anonymous != nil {
		enabled++
	}
	if enabled != 1 {
		return errors.New("Authentication should have one and only one configuration")
	}
	return nil
}

type StaticUsers struct {
	// HtpassswdSecret specifies the htpasswd secret to use for static user authentication.
	HtpasswdSecret ObjectReference `json:"htpasswdSecret"`
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

// APIServerStatus defines the observed state of APIServer
type APIServerStatus struct {
	// ObservedGeneration is the most recent generation observed for this APIServer by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a APIServer's current state.
	Conditions []APIServerCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase APIServerPhaseType `json:"phase,omitempty"`
}

// APIServerPhaseType represents all conditions as a single string for printing by using kubectl commands.
type APIServerPhaseType string

// Values of APIServerPhaseType.
const (
	APIServerPhaseReady       APIServerPhaseType = "Ready"
	APIServerPhasePaused      APIServerPhaseType = "Paused"
	APIServerPhaseNotReady    APIServerPhaseType = "NotReady"
	APIServerPhaseUnknown     APIServerPhaseType = "Unknown"
	APIServerPhaseTerminating APIServerPhaseType = "Terminating"
)

const (
	APIServerTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated.
func (s *APIServerStatus) updatePhase() {

	if paused, ok := s.GetCondition(APIServerPaused); ok && paused.True() {
		s.Phase = APIServerPhasePaused
		return
	}

	for _, condition := range s.Conditions {
		if condition.Type == APIServerReady {
			switch condition.Status {
			case ConditionTrue:
				s.Phase = APIServerPhaseReady
			case ConditionFalse:
				if condition.Reason == APIServerTerminatingReason {
					s.Phase = APIServerPhaseTerminating
				} else {
					s.Phase = APIServerPhaseNotReady
				}
			case ConditionUnknown:
				s.Phase = APIServerPhaseUnknown
			}
			return
		}
	}

	s.Phase = APIServerPhaseUnknown
}

// APIServerConditionType represents a APIServerCondition value.
type APIServerConditionType string

const (
	// APIServerReady represents an APIServer condition is in ready state.
	APIServerReady APIServerConditionType = "Ready"
	// APIServerPaused represents an APIServer condition is in paused state.
	APIServerPaused APIServerConditionType = "Paused"
)

// APIServerCondition contains details for the current condition of this APIServer.
type APIServerCondition struct {
	// Type is the type of the APIServer condition, currently ('Ready').
	Type APIServerConditionType `json:"type"`
	// Status is the status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// LastTransitionTime is the last time the condition transits from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
}

// True returns whether .Status == "True"
func (c APIServerCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *APIServerStatus) GetCondition(t APIServerConditionType) (condition APIServerCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// SetCondition replaces or adds the given condition.
func (s *APIServerStatus) SetCondition(condition APIServerCondition) {
	defer s.updatePhase()

	if condition.LastTransitionTime.IsZero() {
		condition.LastTransitionTime = metav1.Now()
	}

	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {

			// Only update the LastTransitionTime when the Status is changed.
			if s.Conditions[i].Status != condition.Status {
				s.Conditions[i].LastTransitionTime = condition.LastTransitionTime
			}

			s.Conditions[i].Status = condition.Status
			s.Conditions[i].Reason = condition.Reason
			s.Conditions[i].Message = condition.Message

			return
		}
	}

	s.Conditions = append(s.Conditions, condition)
}

// APIServer manages the deployment of the KubeCarrier central API server.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=all
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type APIServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIServerSpec   `json:"spec,omitempty"`
	Status APIServerStatus `json:"status,omitempty"`
}

// IsReady returns if the APIServer is ready.
func (s *APIServer) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == APIServerReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *APIServer) SetReadyCondition() bool {
	if !s.IsReady() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(APIServerCondition{
			Type:    APIServerReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the deployment of the APIServer controller manager is ready",
		})
		return true
	}
	return false
}
func (s *APIServer) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(APIServerReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(APIServerCondition{
			Type:    APIServerReady,
			Status:  ConditionFalse,
			Reason:  "DeploymentUnready",
			Message: "the deployment of the APIServer controller manager is not ready",
		})
		return true
	}
	return false
}

func (s *APIServer) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(APIServerReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != APIServerTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(APIServerCondition{
			Type:    APIServerReady,
			Status:  ConditionFalse,
			Reason:  APIServerTerminatingReason,
			Message: "APIServer is being deleted",
		})
		return true
	}
	return false
}

// IsPaused returns if the APIServer is paused.
func (s *APIServer) IsPaused() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == APIServerPaused &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *APIServer) SetPausedCondition() bool {
	var changed bool
	if !s.IsReady() {
		changed = true
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(APIServerCondition{
			Type:    APIServerReady,
			Status:  ConditionTrue,
			Reason:  "Paused",
			Message: "Reconcilation is paused, assuming component is ready.",
		})
	}
	if !s.IsPaused() {
		changed = true
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(APIServerCondition{
			Type:    APIServerPaused,
			Status:  ConditionTrue,
			Reason:  "Paused",
			Message: "Reconcilation is paused",
		})
	}
	return changed
}

func (s *APIServer) SetUnPausedCondition() bool {
	if s.IsPaused() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(APIServerCondition{
			Type:    APIServerPaused,
			Status:  ConditionFalse,
			Reason:  "UnPaused",
			Message: "Reconcilation is resumed",
		})
		return true
	}
	return false
}

// +kubebuilder:object:root=true

// APIServerList contains a list of APIServer
type APIServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIServer{}, &APIServerList{})
}
