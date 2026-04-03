package k8s

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// NewTestClient creates a Client with injected fake clients for testing.
// cs should be a kubernetes.Interface (e.g. k8sfake.NewClientset()),
// dyn should be a dynamic.Interface (e.g. dynamicfake.NewSimpleDynamicClient()).
// Both may be nil if the test does not exercise those code paths.
func NewTestClient(cs, dyn interface{}) *Client {
	return &Client{
		rawConfig: api.Config{
			Contexts: map[string]*api.Context{
				"test-ctx": {Namespace: "default", Cluster: "test-cluster", AuthInfo: "test-user"},
			},
			CurrentContext: "test-ctx",
		},
		loadingRules: &clientcmd.ClientConfigLoadingRules{
			Precedence: []string{"/dev/null"},
		},
		testClientset: cs,
		testDynClient: dyn,
	}
}
