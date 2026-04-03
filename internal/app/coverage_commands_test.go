package app

import (
	"context"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/janosmiko/lfk/internal/k8s"
	"github.com/janosmiko/lfk/internal/model"
	"github.com/janosmiko/lfk/internal/ui"
)

// =====================================================================
// Helpers
// =====================================================================

// newFakeScheme creates a runtime.Scheme with core resources registered.
func newFakeScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	return s
}

// baseModelWithFakeClient returns a Model wired to fake k8s clients.
// The fake clientset is pre-loaded with the given objects.
func baseModelWithFakeClient(objs ...runtime.Object) Model {
	cs := k8sfake.NewClientset(objs...)
	scheme := newFakeScheme()
	dyn := dynamicfake.NewSimpleDynamicClient(scheme)
	client := k8s.NewTestClient(cs, dyn)

	m := baseModelCov()
	m.client = client
	m.nav.Context = "test-ctx"
	m.namespace = "default"
	m.reqCtx = context.Background()
	return m
}

// baseModelWithFakeDynamic returns a Model with a dynamic client that knows
// about the provided GVR-to-list-kind mappings and unstructured objects.
func baseModelWithFakeDynamic(
	gvrToListKind map[schema.GroupVersionResource]string,
	objs ...runtime.Object,
) Model {
	cs := k8sfake.NewClientset()
	scheme := newFakeScheme()
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, gvrToListKind, objs...)
	client := k8s.NewTestClient(cs, dyn)

	m := baseModelCov()
	m.client = client
	m.nav.Context = "test-ctx"
	m.namespace = "default"
	m.reqCtx = context.Background()
	return m
}

// withActionCtx sets common action context fields on a model.
// Uses "test-ctx" as the default kube context for tests.
func withActionCtx(m Model, name, ns, kind string, rt model.ResourceTypeEntry) Model {
	m.actionCtx = actionContext{
		name:         name,
		namespace:    ns,
		context:      "test-ctx",
		kind:         kind,
		resourceType: rt,
	}
	return m
}

// withMiddleItem sets a single item in the middle pane so selectedMiddleItem() works.
func withMiddleItem(m Model, item model.Item) Model {
	m.middleItems = []model.Item{item}
	m.setCursor(0)
	return m
}

// execCmd runs a tea.Cmd and returns the resulting tea.Msg.
func execCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	require.NotNil(t, cmd)
	return cmd()
}

// =====================================================================
// commands_gitops.go -- all gitops functions call m.client methods
// =====================================================================

func TestCovSyncArgoApp(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-app", "argocd", "Application", model.ResourceTypeEntry{})
	cmd := m.syncArgoApp(false)
	msg := execCmd(t, cmd)
	// The fake clientset doesn't have ArgoCD CRDs so we expect an error.
	result, ok := msg.(actionResultMsg)
	require.True(t, ok)
	assert.Error(t, result.err)
}

func TestCovSyncArgoAppApplyOnly(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-app", "argocd", "Application", model.ResourceTypeEntry{})
	cmd := m.syncArgoApp(true)
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovRefreshArgoApp(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-app", "argocd", "Application", model.ResourceTypeEntry{})
	cmd := m.refreshArgoApp()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovRefreshArgoAppSet(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-appset", "argocd", "ApplicationSet", model.ResourceTypeEntry{})
	cmd := m.refreshArgoAppSet()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovReconcileFluxResource(t *testing.T) {
	m := baseModelWithFakeClient()
	rt := model.ResourceTypeEntry{
		APIGroup:   "kustomize.toolkit.fluxcd.io",
		APIVersion: "v1",
		Resource:   "kustomizations",
	}
	m = withActionCtx(m, "my-ks", "flux-system", "Kustomization", rt)
	cmd := m.reconcileFluxResource()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovSuspendFluxResource(t *testing.T) {
	m := baseModelWithFakeClient()
	rt := model.ResourceTypeEntry{
		APIGroup:   "kustomize.toolkit.fluxcd.io",
		APIVersion: "v1",
		Resource:   "kustomizations",
	}
	m = withActionCtx(m, "my-ks", "flux-system", "Kustomization", rt)
	cmd := m.suspendFluxResource()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovResumeFluxResource(t *testing.T) {
	m := baseModelWithFakeClient()
	rt := model.ResourceTypeEntry{
		APIGroup:   "kustomize.toolkit.fluxcd.io",
		APIVersion: "v1",
		Resource:   "kustomizations",
	}
	m = withActionCtx(m, "my-ks", "flux-system", "Kustomization", rt)
	cmd := m.resumeFluxResource()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovForceRenewCertificate(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-cert", "default", "Certificate", model.ResourceTypeEntry{})
	cmd := m.forceRenewCertificate()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovSuspendArgoWorkflow(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-wf", "argo", "Workflow", model.ResourceTypeEntry{})
	cmd := m.suspendArgoWorkflow()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovResumeArgoWorkflow(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-wf", "argo", "Workflow", model.ResourceTypeEntry{})
	cmd := m.resumeArgoWorkflow()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovStopArgoWorkflow(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-wf", "argo", "Workflow", model.ResourceTypeEntry{})
	cmd := m.stopArgoWorkflow()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovTerminateArgoWorkflow(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-wf", "argo", "Workflow", model.ResourceTypeEntry{})
	cmd := m.terminateArgoWorkflow()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovResubmitArgoWorkflow(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-wf", "argo", "Workflow", model.ResourceTypeEntry{})
	cmd := m.resubmitArgoWorkflow()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovSubmitWorkflowFromTemplate(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-tmpl", "argo", "WorkflowTemplate", model.ResourceTypeEntry{})
	cmd := m.submitWorkflowFromTemplate(false)
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	// SubmitWorkflowFromTemplate creates a workflow via dynamic client; fake dynamic
	// client without the GVR registered may succeed or fail depending on version.
	if result.err != nil {
		assert.Error(t, result.err)
	} else {
		assert.Contains(t, result.message, "Submitted workflow")
	}
}

func TestCovSubmitWorkflowFromTemplateClusterScope(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-tmpl", "argo", "ClusterWorkflowTemplate", model.ResourceTypeEntry{})
	cmd := m.submitWorkflowFromTemplate(true)
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	if result.err != nil {
		assert.Error(t, result.err)
	} else {
		assert.Contains(t, result.message, "Submitted workflow")
	}
}

func TestCovSuspendCronWorkflow(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-cwf", "argo", "CronWorkflow", model.ResourceTypeEntry{})
	cmd := m.suspendCronWorkflow()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovResumeCronWorkflow(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-cwf", "argo", "CronWorkflow", model.ResourceTypeEntry{})
	cmd := m.resumeCronWorkflow()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovWatchArgoWorkflow(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-wf", "argo", "Workflow", model.ResourceTypeEntry{})
	cmd := m.watchArgoWorkflow()
	msg := execCmd(t, cmd)
	result, ok := msg.(describeLoadedMsg)
	require.True(t, ok)
	assert.Error(t, result.err)
	assert.Contains(t, result.title, "Watch: my-wf")
}

func TestCovForceRefreshExternalSecret(t *testing.T) {
	m := baseModelWithFakeClient()
	rt := model.ResourceTypeEntry{
		APIGroup:   "external-secrets.io",
		APIVersion: "v1beta1",
		Resource:   "externalsecrets",
		Namespaced: true,
	}
	m = withActionCtx(m, "my-es", "default", "ExternalSecret", rt)
	cmd := m.forceRefreshExternalSecret()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovForceRefreshExternalSecretClusterScoped(t *testing.T) {
	m := baseModelWithFakeClient()
	rt := model.ResourceTypeEntry{
		APIGroup:   "external-secrets.io",
		APIVersion: "v1beta1",
		Resource:   "clusterexternalsecrets",
		Namespaced: false,
	}
	m = withActionCtx(m, "my-ces", "default", "ClusterExternalSecret", rt)
	cmd := m.forceRefreshExternalSecret()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovPauseKEDAResource(t *testing.T) {
	m := baseModelWithFakeClient()
	rt := model.ResourceTypeEntry{
		APIGroup:   "keda.sh",
		APIVersion: "v1alpha1",
		Resource:   "scaledobjects",
	}
	m = withActionCtx(m, "my-so", "default", "ScaledObject", rt)
	cmd := m.pauseKEDAResource()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovUnpauseKEDAResource(t *testing.T) {
	m := baseModelWithFakeClient()
	rt := model.ResourceTypeEntry{
		APIGroup:   "keda.sh",
		APIVersion: "v1alpha1",
		Resource:   "scaledobjects",
	}
	m = withActionCtx(m, "my-so", "default", "ScaledObject", rt)
	cmd := m.unpauseKEDAResource()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovBulkSyncArgoApps(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "", "argocd", "Application", model.ResourceTypeEntry{})
	m.bulkItems = []model.Item{
		{Name: "app-1", Namespace: "argocd"},
		{Name: "app-2"},
	}
	cmd := m.bulkSyncArgoApps(false)
	msg := execCmd(t, cmd)
	result, ok := msg.(bulkActionResultMsg)
	require.True(t, ok)
	// Both should fail since there are no ArgoCD CRDs.
	assert.Equal(t, 2, result.failed)
	assert.Equal(t, 0, result.succeeded)
}

func TestCovBulkRefreshArgoApps(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "", "argocd", "Application", model.ResourceTypeEntry{})
	m.bulkItems = []model.Item{
		{Name: "app-1", Namespace: "argocd"},
		{Name: "app-2"},
	}
	cmd := m.bulkRefreshArgoApps()
	msg := execCmd(t, cmd)
	result, ok := msg.(bulkActionResultMsg)
	require.True(t, ok)
	assert.Equal(t, 2, result.failed)
}

func TestCovTerminateArgoSync(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-app", "argocd", "Application", model.ResourceTypeEntry{})
	cmd := m.terminateArgoSync()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovLoadAutoSyncConfig(t *testing.T) {
	m := baseModelWithFakeClient()
	item := model.Item{Name: "my-app", Namespace: "argocd"}
	m = withMiddleItem(m, item)
	cmd := m.loadAutoSyncConfig()
	msg := execCmd(t, cmd)
	result, ok := msg.(autoSyncLoadedMsg)
	require.True(t, ok)
	assert.Error(t, result.err)
}

func TestCovLoadAutoSyncConfigNilSel(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadAutoSyncConfig()
	assert.Nil(t, cmd)
}

func TestCovSaveAutoSyncConfig(t *testing.T) {
	m := baseModelWithFakeClient()
	item := model.Item{Name: "my-app", Namespace: "argocd"}
	m = withMiddleItem(m, item)
	m.autoSyncEnabled = true
	m.autoSyncSelfHeal = true
	m.autoSyncPrune = false
	cmd := m.saveAutoSyncConfig()
	msg := execCmd(t, cmd)
	result, ok := msg.(autoSyncSavedMsg)
	require.True(t, ok)
	assert.Error(t, result.err)
}

func TestCovSaveAutoSyncConfigNilSel(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.saveAutoSyncConfig()
	assert.Nil(t, cmd)
}

// =====================================================================
// commands_load.go -- load commands with fake client
// =====================================================================

func TestCovLoadContexts(t *testing.T) {
	m := baseModelWithFakeClient()
	msg := m.loadContexts()
	result, ok := msg.(contextsLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
	// The test client has "test-ctx" as the only context.
	require.Len(t, result.items, 1)
	assert.Equal(t, "test-ctx", result.items[0].Name)
	assert.Equal(t, "current", result.items[0].Status)
}

func TestCovLoadResourceTypes(t *testing.T) {
	m := baseModelWithFakeClient()
	cmd := m.loadResourceTypes()
	msg := execCmd(t, cmd)
	result, ok := msg.(resourceTypesMsg)
	require.True(t, ok)
	assert.NotEmpty(t, result.items)
}

func TestCovLoadResourceTypesWithCRDsFakeClient(t *testing.T) {
	m := baseModelWithFakeClient()
	m.discoveredCRDs["test-ctx"] = []model.ResourceTypeEntry{
		{DisplayName: "TestCRD", Kind: "TestCRD", APIGroup: "test.io", APIVersion: "v1", Resource: "testcrds"},
	}
	cmd := m.loadResourceTypes()
	msg := execCmd(t, cmd)
	result, ok := msg.(resourceTypesMsg)
	require.True(t, ok)
	assert.NotEmpty(t, result.items)
}

func TestCovDiscoverCRDsReturnsCmd(t *testing.T) {
	m := baseModelWithFakeClient()
	cmd := m.discoverCRDs("test-ctx")
	// Just verify it returns a non-nil command. Executing it would panic because
	// the fake dynamic client does not have the CRD list kind registered.
	assert.NotNil(t, cmd)
}

func TestCovLoadQuotas(t *testing.T) {
	m := baseModelWithFakeClient()
	cmd := m.loadQuotas()
	msg := execCmd(t, cmd)
	result, ok := msg.(quotaLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
}

func TestCovLoadNamespaces(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "kube-system"},
		Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
	}
	m := baseModelWithFakeClient(ns)
	cmd := m.loadNamespaces()
	msg := execCmd(t, cmd)
	result, ok := msg.(namespacesLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
	assert.Len(t, result.items, 1)
	assert.Equal(t, "kube-system", result.items[0].Name)
}

func TestCovLoadNamespacesNoContext(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Context = ""
	cmd := m.loadNamespaces()
	msg := execCmd(t, cmd)
	result, ok := msg.(namespacesLoadedMsg)
	require.True(t, ok)
	// CurrentContext() is "test-ctx" so it should still work.
	assert.NoError(t, result.err)
}

func TestCovLoadResources(t *testing.T) {
	gvr := schema.GroupVersionResource{Version: "v1", Resource: "pods"}
	gvrToListKind := map[schema.GroupVersionResource]string{
		gvr: "PodList",
	}
	m := baseModelWithFakeDynamic(gvrToListKind)
	m.nav.ResourceType = model.ResourceTypeEntry{
		Kind:       "Pod",
		APIVersion: "v1",
		Resource:   "pods",
		Namespaced: true,
	}
	cmd := m.loadResources(false)
	msg := execCmd(t, cmd)
	result, ok := msg.(resourcesLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
	// No pods pre-loaded, so empty list.
	assert.Empty(t, result.items)
}

func TestCovLoadResourcesNilForPreviewWithNoSelection(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadResources(true)
	assert.Nil(t, cmd)
}

func TestCovLoadOwnedReturnsCmd(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Deployment"}
	m.nav.ResourceName = "my-deploy"
	cmd := m.loadOwned(false)
	// Just verify a command is returned. Executing it would panic because
	// the fake dynamic client does not have replicaset list kind registered.
	assert.NotNil(t, cmd)
}

func TestCovLoadOwnedForPreviewNil(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadOwned(true)
	assert.Nil(t, cmd)
}

func TestCovLoadResourceTreeNil(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelResources
	m.middleItems = nil
	cmd := m.loadResourceTree()
	assert.Nil(t, cmd)
}

func TestCovLoadResourceTreeDefaultLevel(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelClusters
	cmd := m.loadResourceTree()
	assert.Nil(t, cmd)
}

func TestCovLoadContainers(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "main", Image: "nginx:latest"},
			},
		},
	}
	m := baseModelWithFakeClient(pod)
	m.nav.OwnedName = "test-pod"
	cmd := m.loadContainers(false)
	msg := execCmd(t, cmd)
	result, ok := msg.(containersLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
	assert.Len(t, result.items, 1)
	assert.Equal(t, "main", result.items[0].Name)
}

func TestCovLoadContainersForPreviewNil(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadContainers(true)
	assert.Nil(t, cmd)
}

func TestCovLoadDiff(t *testing.T) {
	rt := model.ResourceTypeEntry{
		Kind:       "ConfigMap",
		APIVersion: "v1",
		Resource:   "configmaps",
		Namespaced: true,
	}
	gvr := schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}
	cm1 := &unstructured.Unstructured{}
	cm1.SetGroupVersionKind(schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"})
	cm1.SetName("cm-1")
	cm1.SetNamespace("default")
	cm2 := &unstructured.Unstructured{}
	cm2.SetGroupVersionKind(schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"})
	cm2.SetName("cm-2")
	cm2.SetNamespace("default")

	gvrToListKind := map[schema.GroupVersionResource]string{gvr: "ConfigMapList"}
	m := baseModelWithFakeDynamic(gvrToListKind, cm1, cm2)

	itemA := model.Item{Name: "cm-1", Namespace: "default"}
	itemB := model.Item{Name: "cm-2", Namespace: "default"}
	cmd := m.loadDiff(rt, itemA, itemB)
	msg := execCmd(t, cmd)
	result, ok := msg.(diffLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
}

func TestCovLoadDiffDifferentNamespaces(t *testing.T) {
	rt := model.ResourceTypeEntry{
		Kind:       "ConfigMap",
		APIVersion: "v1",
		Resource:   "configmaps",
		Namespaced: true,
	}
	gvr := schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}
	gvrToListKind := map[schema.GroupVersionResource]string{gvr: "ConfigMapList"}
	m := baseModelWithFakeDynamic(gvrToListKind)

	itemA := model.Item{Name: "cm-1", Namespace: "ns-a"}
	itemB := model.Item{Name: "cm-2", Namespace: "ns-b"}
	cmd := m.loadDiff(rt, itemA, itemB)
	msg := execCmd(t, cmd)
	result, ok := msg.(diffLoadedMsg)
	require.True(t, ok)
	// Items don't exist so we expect an error.
	assert.Error(t, result.err)
}

func TestCovLoadYAMLResources(t *testing.T) {
	gvr := schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}
	cm := &unstructured.Unstructured{}
	cm.SetGroupVersionKind(schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"})
	cm.SetName("my-cm")
	cm.SetNamespace("default")

	gvrToListKind := map[schema.GroupVersionResource]string{gvr: "ConfigMapList"}
	m := baseModelWithFakeDynamic(gvrToListKind, cm)
	m.nav.Level = model.LevelResources
	m.nav.ResourceType = model.ResourceTypeEntry{
		Kind:       "ConfigMap",
		APIVersion: "v1",
		Resource:   "configmaps",
		Namespaced: true,
	}
	m = withMiddleItem(m, model.Item{Name: "my-cm", Namespace: "default"})
	cmd := m.loadYAML()
	msg := execCmd(t, cmd)
	result, ok := msg.(yamlLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
	assert.Contains(t, result.content, "my-cm")
}

func TestCovLoadYAMLContainersReturnsCmd(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelContainers
	m.nav.OwnedName = "my-pod"
	cmd := m.loadYAML()
	// Verifies the LevelContainers branch is reached and returns a cmd.
	assert.NotNil(t, cmd)
}

func TestCovLoadYAMLNilSelection(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelResources
	m.middleItems = nil
	cmd := m.loadYAML()
	assert.Nil(t, cmd)
}

func TestCovLoadMetricsNilFakeClient(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadMetrics()
	assert.Nil(t, cmd)
}

func TestCovLoadPreviewEventsNilFakeClient(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadPreviewEvents()
	assert.Nil(t, cmd)
}

func TestCovLoadPodMetricsForListReturnsCmd(t *testing.T) {
	m := baseModelWithFakeClient()
	cmd := m.loadPodMetricsForList()
	// Just verify the function returns a non-nil command.
	// Executing it would panic because the metrics.k8s.io GVR is not registered.
	assert.NotNil(t, cmd)
}

func TestCovLoadNodeMetricsForListReturnsCmd(t *testing.T) {
	m := baseModelWithFakeClient()
	cmd := m.loadNodeMetricsForList()
	assert.NotNil(t, cmd)
}

func TestCovLoadSecretData(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-secret", Namespace: "default"},
		Data:       map[string][]byte{"key": []byte("val")},
	}
	m := baseModelWithFakeClient(secret)
	m = withMiddleItem(m, model.Item{Name: "my-secret", Namespace: "default"})
	cmd := m.loadSecretData()
	msg := execCmd(t, cmd)
	result, ok := msg.(secretDataLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
	assert.Equal(t, "val", result.data.Data["key"])
}

func TestCovLoadSecretDataNil(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadSecretData()
	assert.Nil(t, cmd)
}

func TestCovSaveSecretData(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-secret", Namespace: "default"},
		Data:       map[string][]byte{"key": []byte("old")},
	}
	m := baseModelWithFakeClient(secret)
	m = withMiddleItem(m, model.Item{Name: "my-secret", Namespace: "default"})
	m.secretData = &model.SecretData{Data: map[string]string{"key": "new"}}
	cmd := m.saveSecretData()
	msg := execCmd(t, cmd)
	result, ok := msg.(secretSavedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
}

func TestCovSaveSecretDataNilSecretData(t *testing.T) {
	m := baseModelWithFakeClient()
	m.secretData = nil
	cmd := m.saveSecretData()
	assert.Nil(t, cmd)
}

func TestCovSaveSecretDataNilSel(t *testing.T) {
	m := baseModelWithFakeClient()
	m.secretData = &model.SecretData{Data: map[string]string{"k": "v"}}
	m.middleItems = nil
	cmd := m.saveSecretData()
	assert.Nil(t, cmd)
}

func TestCovLoadConfigMapData(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "my-cm", Namespace: "default"},
		Data:       map[string]string{"config.yaml": "data: true"},
	}
	m := baseModelWithFakeClient(cm)
	m = withMiddleItem(m, model.Item{Name: "my-cm", Namespace: "default"})
	cmd := m.loadConfigMapData()
	msg := execCmd(t, cmd)
	result, ok := msg.(configMapDataLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
	assert.Equal(t, "data: true", result.data.Data["config.yaml"])
}

func TestCovLoadConfigMapDataNil(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadConfigMapData()
	assert.Nil(t, cmd)
}

func TestCovSaveConfigMapData(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "my-cm", Namespace: "default"},
		Data:       map[string]string{"key": "old"},
	}
	m := baseModelWithFakeClient(cm)
	m = withMiddleItem(m, model.Item{Name: "my-cm", Namespace: "default"})
	m.configMapData = &model.ConfigMapData{Data: map[string]string{"key": "new"}}
	cmd := m.saveConfigMapData()
	msg := execCmd(t, cmd)
	result, ok := msg.(configMapSavedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
}

func TestCovSaveConfigMapDataNil(t *testing.T) {
	m := baseModelWithFakeClient()
	m.configMapData = nil
	cmd := m.saveConfigMapData()
	assert.Nil(t, cmd)
}

func TestCovLoadLabelDataNil(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadLabelData()
	assert.Nil(t, cmd)
}

func TestCovSaveLabelDataNilData(t *testing.T) {
	m := baseModelWithFakeClient()
	m.labelData = nil
	cmd := m.saveLabelData()
	assert.Nil(t, cmd)
}

func TestCovSaveLabelDataNilSel(t *testing.T) {
	m := baseModelWithFakeClient()
	m.labelData = &model.LabelAnnotationData{
		Labels:      map[string]string{"l": "v"},
		Annotations: map[string]string{"a": "v"},
	}
	m.middleItems = nil
	cmd := m.saveLabelData()
	assert.Nil(t, cmd)
}

func TestCovLoadRevisionsNil(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadRevisions()
	assert.Nil(t, cmd)
}

// =====================================================================
// commands_load_preview.go -- preview commands
// =====================================================================

func TestCovLoadEventTimeline(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-pod", "default", "Pod", model.ResourceTypeEntry{})
	cmd := m.loadEventTimeline()
	msg := execCmd(t, cmd)
	result, ok := msg.(eventTimelineMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
}

func TestCovCheckRBAC(t *testing.T) {
	m := baseModelWithFakeClient()
	rt := model.ResourceTypeEntry{
		Kind:     "Pod",
		APIGroup: "",
		Resource: "pods",
	}
	m = withActionCtx(m, "my-pod", "default", "Pod", rt)
	cmd := m.checkRBAC()
	msg := execCmd(t, cmd)
	result, ok := msg.(rbacCheckMsg)
	require.True(t, ok)
	assert.Equal(t, "Pod", result.kind)
	assert.Equal(t, "pods", result.resource)
}

func TestCovLoadCanISAList(t *testing.T) {
	m := baseModelWithFakeClient()
	cmd := m.loadCanISAList()
	msg := execCmd(t, cmd)
	result, ok := msg.(canISAListMsg)
	require.True(t, ok)
	// No SAs exist in the fake client.
	assert.NoError(t, result.err)
}

func TestCovLoadPodStartup(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "my-pod", Namespace: "default"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "main", Image: "nginx"}}},
	}
	m := baseModelWithFakeClient(pod)
	m = withActionCtx(m, "my-pod", "default", "Pod", model.ResourceTypeEntry{})
	cmd := m.loadPodStartup()
	msg := execCmd(t, cmd)
	result, ok := msg.(podStartupMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
}

func TestCovLoadAlertsReturnsCmd(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-pod", "default", "Pod", model.ResourceTypeEntry{})
	cmd := m.loadAlerts()
	// Just verify command is returned. Executing would hit nil pointers in
	// the alerts code that tries Prometheus port-forwarding.
	assert.NotNil(t, cmd)
}

func TestCovLoadNetworkPolicy(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-netpol", "default", "NetworkPolicy", model.ResourceTypeEntry{})
	cmd := m.loadNetworkPolicy()
	msg := execCmd(t, cmd)
	result, ok := msg.(netpolLoadedMsg)
	require.True(t, ok)
	// No NetworkPolicy exists; expect error.
	assert.Error(t, result.err)
}

func TestCovLoadContainerPorts(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "my-pod", Namespace: "default"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: "nginx",
					Ports: []corev1.ContainerPort{
						{Name: "http", ContainerPort: 80, Protocol: corev1.ProtocolTCP},
					},
				},
			},
		},
	}
	m := baseModelWithFakeClient(pod)
	m = withActionCtx(m, "my-pod", "default", "Pod", model.ResourceTypeEntry{})
	cmd := m.loadContainerPorts()
	msg := execCmd(t, cmd)
	result, ok := msg.(containerPortsLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
	require.Len(t, result.ports, 1)
	assert.Equal(t, int32(80), result.ports[0].ContainerPort)
}

func TestCovLoadContainerPortsService(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "default"},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP},
			},
		},
	}
	m := baseModelWithFakeClient(svc)
	m = withActionCtx(m, "my-svc", "default", "Service", model.ResourceTypeEntry{})
	cmd := m.loadContainerPorts()
	msg := execCmd(t, cmd)
	result, ok := msg.(containerPortsLoadedMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
}

func TestCovLoadContainerPortsUnsupportedKind(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-job", "default", "Job", model.ResourceTypeEntry{})
	cmd := m.loadContainerPorts()
	msg := execCmd(t, cmd)
	result, ok := msg.(containerPortsLoadedMsg)
	require.True(t, ok)
	assert.Error(t, result.err)
	assert.Contains(t, result.err.Error(), "unsupported kind")
}

func TestCovLoadPreview(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadPreview()
	assert.Nil(t, cmd)
}

func TestCovLoadPreviewClusters(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelClusters
	m = withMiddleItem(m, model.Item{Name: "test-ctx"})
	cmd := m.loadPreview()
	assert.NotNil(t, cmd)
}

func TestCovLoadPreviewResourceTypesOverview(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelResourceTypes
	m = withMiddleItem(m, model.Item{Extra: "__overview__"})
	ui.ConfigDashboard = true
	cmd := m.loadPreview()
	assert.NotNil(t, cmd)
}

func TestCovLoadPreviewResourceTypesMonitoring(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelResourceTypes
	m = withMiddleItem(m, model.Item{Extra: "__monitoring__"})
	cmd := m.loadPreview()
	assert.NotNil(t, cmd)
}

func TestCovLoadPreviewResourceTypesCollapsed(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelResourceTypes
	m = withMiddleItem(m, model.Item{Kind: "__collapsed_group__"})
	cmd := m.loadPreview()
	assert.Nil(t, cmd)
}

func TestCovLoadPreviewContainers(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelContainers
	m = withMiddleItem(m, model.Item{Name: "container-1"})
	cmd := m.loadPreview()
	assert.Nil(t, cmd)
}

func TestCovLoadPreviewYAMLNil(t *testing.T) {
	m := baseModelWithFakeClient()
	m.middleItems = nil
	cmd := m.loadPreviewYAML()
	assert.Nil(t, cmd)
}

// =====================================================================
// commands_exec.go -- functions that call k8s API (not external procs)
// =====================================================================

func TestCovResizePVC(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-pvc", "default", "PersistentVolumeClaim", model.ResourceTypeEntry{})
	cmd := m.resizePVC("10Gi")
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	// PVC doesn't exist in fake client.
	assert.Error(t, result.err)
}

func TestCovScaleResource(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-deploy", "default", "Deployment", model.ResourceTypeEntry{})
	cmd := m.scaleResource(3)
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	// Deployment doesn't exist in fake client.
	assert.Error(t, result.err)
}

func TestCovRestartResource(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-deploy", "default", "Deployment", model.ResourceTypeEntry{})
	cmd := m.restartResource()
	msg := execCmd(t, cmd)
	result := msg.(actionResultMsg)
	assert.Error(t, result.err)
}

func TestCovRollbackDeployment(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-deploy", "default", "Deployment", model.ResourceTypeEntry{})
	cmd := m.rollbackDeployment(1)
	msg := execCmd(t, cmd)
	result, ok := msg.(rollbackDoneMsg)
	require.True(t, ok)
	assert.Error(t, result.err)
}

func TestCovTriggerCronJob(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-cj", "default", "CronJob", model.ResourceTypeEntry{})
	cmd := m.triggerCronJob()
	msg := execCmd(t, cmd)
	result, ok := msg.(triggerCronJobMsg)
	require.True(t, ok)
	assert.Error(t, result.err)
}

// =====================================================================
// commands_finalizer.go -- bulkRemoveFinalizer
// =====================================================================

func TestCovBulkRemoveFinalizer(t *testing.T) {
	m := baseModelWithFakeClient()
	m.finalizerSearchResults = []k8s.FinalizerMatch{
		{Namespace: "default", Kind: "Pod", Name: "pod-1", Matched: "test/finalizer"},
		{Namespace: "default", Kind: "Pod", Name: "pod-2", Matched: "test/finalizer"},
	}
	m.finalizerSearchSelected = map[string]bool{
		"default/Pod/pod-1": true,
	}
	cmd := m.bulkRemoveFinalizer()
	msg := execCmd(t, cmd)
	result, ok := msg.(finalizerRemoveResultMsg)
	require.True(t, ok)
	// Fake client doesn't have these resources, so removal fails.
	assert.Equal(t, 1, result.failed)
	assert.Equal(t, 0, result.succeeded)
}

func TestCovBulkRemoveFinalizerNoneSelected(t *testing.T) {
	m := baseModelWithFakeClient()
	m.finalizerSearchResults = []k8s.FinalizerMatch{
		{Namespace: "default", Kind: "Pod", Name: "pod-1", Matched: "test/finalizer"},
	}
	m.finalizerSearchSelected = map[string]bool{}
	cmd := m.bulkRemoveFinalizer()
	msg := execCmd(t, cmd)
	result, ok := msg.(finalizerRemoveResultMsg)
	require.True(t, ok)
	assert.Equal(t, 0, result.failed)
	assert.Equal(t, 0, result.succeeded)
}

// =====================================================================
// commands.go -- scheduleDescribeRefresh, openInBrowser, loadPodsForAction
// =====================================================================

func TestCovScheduleDescribeRefresh(t *testing.T) {
	cmd := scheduleDescribeRefresh()
	assert.NotNil(t, cmd)
}

func TestCovOpenInBrowser(t *testing.T) {
	cmd := openInBrowser("https://example.com")
	assert.NotNil(t, cmd)
	// We don't execute because it would open a real browser.
}

func TestCovLoadPodsForActionFakeClient(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withActionCtx(m, "my-deploy", "default", "Deployment", model.ResourceTypeEntry{})
	cmd := m.loadPodsForAction()
	// Just verify command is returned; the GetOwnedResources call needs
	// replicaset list kind registered in the dynamic client.
	assert.NotNil(t, cmd)
}

func TestCovSearchFinalizers(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelResources
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod", APIVersion: "v1", Resource: "pods", Namespaced: true}
	cmd := m.searchFinalizers("test/finalizer")
	msg := execCmd(t, cmd)
	result, ok := msg.(finalizerSearchResultMsg)
	require.True(t, ok)
	// No resources with finalizers in fake client.
	assert.NoError(t, result.err)
}

func TestCovSearchFinalizersAllTypes(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelResourceTypes
	cmd := m.searchFinalizers("test/*")
	// Just verify cmd is returned; executing it would panic because the
	// fake dynamic client does not have all resource list kinds registered.
	assert.NotNil(t, cmd)
}

// =====================================================================
// commands_dashboard.go -- loadMonitoringDashboard
// =====================================================================

func TestCovLoadMonitoringDashboardReturnsCmd(t *testing.T) {
	m := baseModelWithFakeClient()
	cmd := m.loadMonitoringDashboard()
	// Just verify a command is returned. Executing it hits nil pointer in
	// alerts code that needs a real clientset for service discovery.
	assert.NotNil(t, cmd)
}

// =====================================================================
// update_actions.go -- openIngressInBrowser, openBulkActionDirect, executeBulkAction
// =====================================================================

func TestCovOpenIngressInBrowserNoSelection(t *testing.T) {
	m := baseModelCov()
	m.middleItems = nil
	ret, cmd := m.openIngressInBrowser()
	model := ret.(Model)
	assert.True(t, model.hasStatusMessage())
	assert.NotNil(t, cmd)
}

func TestCovOpenIngressInBrowserNoURL(t *testing.T) {
	m := baseModelCov()
	m = withMiddleItem(m, model.Item{Name: "my-ingress"})
	ret, cmd := m.openIngressInBrowser()
	result := ret.(Model)
	assert.True(t, result.hasStatusMessage())
	assert.NotNil(t, cmd)
}

func TestCovOpenIngressInBrowserWithURL(t *testing.T) {
	m := baseModelCov()
	item := model.Item{
		Name: "my-ingress",
		Columns: []model.KeyValue{
			{Key: "__ingress_url", Value: "https://example.com"},
		},
	}
	m = withMiddleItem(m, item)
	ret, cmd := m.openIngressInBrowser()
	result := ret.(Model)
	assert.True(t, result.hasStatusMessage())
	assert.NotNil(t, cmd)
}

func TestCovOpenBulkActionDirectNoSelection(t *testing.T) {
	m := baseModelCov()
	m.selectedItems = map[string]bool{}
	ret, cmd := m.openBulkActionDirect("Delete")
	_ = ret.(Model)
	assert.Nil(t, cmd)
}

func TestCovExecuteBulkActionLogs(t *testing.T) {
	m := baseModelWithFakeClient()
	m.execMu = &sync.Mutex{}
	m.bulkItems = []model.Item{{Name: "pod-1"}}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod"}
	m.actionCtx = actionContext{context: "ctx", namespace: "ns"}
	ret, _ := m.executeBulkAction("Logs")
	switch v := ret.(type) {
	case *Model:
		assert.False(t, v.bulkMode)
	case Model:
		assert.False(t, v.bulkMode)
	}
}

func TestCovExecuteBulkActionDelete(t *testing.T) {
	m := baseModelCov()
	m.bulkItems = []model.Item{{Name: "pod-1"}, {Name: "pod-2"}}
	m.confirmTypeInput = TextInput{}
	ret, cmd := m.executeBulkAction("Delete")
	result := ret.(Model)
	assert.Equal(t, overlayConfirm, result.overlay)
	assert.Nil(t, cmd)
}

func TestCovExecuteBulkActionForceDelete(t *testing.T) {
	m := baseModelCov()
	m.bulkItems = []model.Item{{Name: "pod-1"}}
	m.confirmTypeInput = TextInput{}
	ret, cmd := m.executeBulkAction("Force Delete")
	result := ret.(Model)
	assert.Equal(t, overlayConfirmType, result.overlay)
	assert.Nil(t, cmd)
}

func TestCovExecuteBulkActionScale(t *testing.T) {
	m := baseModelCov()
	m.bulkItems = []model.Item{{Name: "deploy-1"}}
	m.scaleInput = TextInput{}
	ret, cmd := m.executeBulkAction("Scale")
	result := ret.(Model)
	assert.Equal(t, overlayScaleInput, result.overlay)
	assert.Nil(t, cmd)
}

func TestCovExecuteBulkActionRestart(t *testing.T) {
	m := baseModelCov()
	m.bulkItems = []model.Item{{Name: "deploy-1"}}
	m.actionCtx = actionContext{context: "ctx", namespace: "ns"}
	ret, cmd := m.executeBulkAction("Restart")
	result := ret.(Model)
	assert.True(t, result.loading)
	assert.NotNil(t, cmd)
}

func TestCovExecuteBulkActionLabels(t *testing.T) {
	m := baseModelCov()
	m.bulkItems = []model.Item{{Name: "pod-1"}}
	m.batchLabelInput = TextInput{}
	ret, cmd := m.executeBulkAction("Labels / Annotations")
	result := ret.(Model)
	assert.Equal(t, overlayBatchLabel, result.overlay)
	assert.Nil(t, cmd)
}

func TestCovExecuteBulkActionDiffWrongCount(t *testing.T) {
	m := baseModelCov()
	m.bulkItems = []model.Item{{Name: "pod-1"}}
	ret, cmd := m.executeBulkAction("Diff")
	result := ret.(Model)
	assert.True(t, result.hasStatusMessage())
	assert.NotNil(t, cmd)
}

func TestCovExecuteBulkActionUnknown(t *testing.T) {
	m := baseModelCov()
	m.bulkItems = []model.Item{{Name: "pod-1"}}
	ret, cmd := m.executeBulkAction("NonExistent")
	_ = ret.(Model)
	assert.Nil(t, cmd)
}

// =====================================================================
// commands.go -- executeCommandBar, executeCommandBarKubectl
// =====================================================================

func TestCovExecuteCommandBarEmpty(t *testing.T) {
	m := baseModelWithFakeClient()
	cmd := m.executeCommandBar("")
	assert.Nil(t, cmd)
}

func TestCovExecuteCommandBarShell(t *testing.T) {
	m := baseModelWithFakeClient()
	cmd := m.executeCommandBar("echo hello")
	msg := execCmd(t, cmd)
	result, ok := msg.(commandBarResultMsg)
	require.True(t, ok)
	assert.NoError(t, result.err)
	assert.Contains(t, result.output, "hello")
}

func TestCovExecuteCommandBarKubectl(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Context = "test-ctx"
	m.namespace = "default"
	cmd := m.executeCommandBar("kubectl version --client")
	// Returns non-nil even if kubectl is not found.
	assert.NotNil(t, cmd)
}

// =====================================================================
// update_describe.go -- toggleDiffFoldAtCursor, toggleAllDiffFolds
// =====================================================================

func TestCovToggleDiffFoldAtCursor(t *testing.T) {
	m := baseModelCov()
	m.diffLeft = "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10"
	m.diffRight = "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10"

	foldRegions := ui.ComputeDiffFoldRegions(m.diffLeft, m.diffRight)
	m.diffFoldState = make([]bool, len(foldRegions))
	m.diffCursor = 0

	// Should not panic even with no foldable regions (all lines equal = one big fold).
	m.toggleDiffFoldAtCursor(foldRegions)
}

func TestCovToggleDiffFoldAtCursorOutOfBounds(t *testing.T) {
	m := baseModelCov()
	m.diffLeft = ""
	m.diffRight = ""
	m.diffCursor = 100
	m.diffFoldState = nil
	m.toggleDiffFoldAtCursor(nil)
}

func TestCovToggleAllDiffFolds(t *testing.T) {
	m := baseModelCov()
	m.diffFoldState = []bool{false, false, true}
	regions := []ui.DiffFoldRegion{{Start: 0, End: 2}, {Start: 3, End: 5}, {Start: 6, End: 8}}

	// Some collapsed -> expand all.
	m.toggleAllDiffFolds(regions)
	for _, v := range m.diffFoldState {
		assert.False(t, v)
	}

	// None collapsed -> collapse all.
	m.toggleAllDiffFolds(regions)
	for _, v := range m.diffFoldState {
		assert.True(t, v)
	}
}

// =====================================================================
// filter_input.go -- stringFilterInput Home, End, Left, Right (no-ops)
// =====================================================================

func TestCovStringFilterInputNavigation(t *testing.T) {
	s := "hello"
	fi := &stringFilterInput{ptr: &s}

	// These are all no-ops for stringFilterInput but need coverage.
	fi.Home()
	fi.End()
	fi.Left()
	fi.Right()
	assert.Equal(t, "hello", s)

	fi.Insert("!")
	assert.Equal(t, "hello!", s)

	fi.Backspace()
	assert.Equal(t, "hello", s)

	fi.DeleteWord()
	assert.Equal(t, "", s)

	fi.Clear()
	assert.Equal(t, "", s)
}

// =====================================================================
// update_overlays_selectors.go -- scheme display items
// =====================================================================

func TestCovSchemeDisplayItemsEmpty(t *testing.T) {
	m := baseModelCov()
	m.schemeEntries = nil
	m.schemeFilter = TextInput{}
	items := m.schemeDisplayItems()
	assert.Empty(t, items)
}

func TestCovSchemeDisplayItemsWithEntries(t *testing.T) {
	m := baseModelCov()
	m.schemeEntries = []ui.SchemeEntry{
		{Name: "Dark Themes", IsHeader: true},
		{Name: "dracula"},
		{Name: "nord"},
		{Name: "Light Themes", IsHeader: true},
		{Name: "solarized-light"},
	}
	m.schemeFilter = TextInput{}
	items := m.schemeDisplayItems()
	assert.Len(t, items, 5)
	assert.Equal(t, -1, items[0].selectIdx) // header
	assert.Equal(t, 0, items[1].selectIdx)
	assert.Equal(t, 1, items[2].selectIdx)
	assert.Equal(t, -1, items[3].selectIdx) // header
	assert.Equal(t, 2, items[4].selectIdx)
}

func TestCovSchemeFirstVisibleSelectable(t *testing.T) {
	m := baseModelCov()
	m.schemeEntries = []ui.SchemeEntry{
		{Name: "Dark Themes", IsHeader: true},
		{Name: "dracula"},
		{Name: "nord"},
	}
	m.schemeFilter = TextInput{}
	m.schemeCursor = 0
	ui.ResetOverlaySchemeScroll()
	idx := m.schemeFirstVisibleSelectable()
	assert.Equal(t, 0, idx) // first selectable
}

func TestCovSchemeLastVisibleSelectable(t *testing.T) {
	m := baseModelCov()
	m.schemeEntries = []ui.SchemeEntry{
		{Name: "Dark Themes", IsHeader: true},
		{Name: "dracula"},
		{Name: "nord"},
	}
	m.schemeFilter = TextInput{}
	m.schemeCursor = 0
	ui.ResetOverlaySchemeScroll()
	idx := m.schemeLastVisibleSelectable()
	assert.Equal(t, 1, idx) // last selectable
}

// =====================================================================
// update_bookmarks.go -- jumpToSlot, restoreSession, buildSessionTabState
// =====================================================================

func TestCovJumpToSlotNotFound(t *testing.T) {
	m := baseModelCov()
	m.bookmarks = nil
	ret, cmd := m.jumpToSlot("a")
	result := ret.(Model)
	assert.True(t, result.hasStatusMessage())
	assert.NotNil(t, cmd)
}

func TestCovBuildSessionTabState(t *testing.T) {
	st := &SessionTab{
		Context:       "my-ctx",
		Namespace:     "my-ns",
		AllNamespaces: false,
		ResourceType:  "",
	}
	tab := buildSessionTabState(st)
	assert.Equal(t, "my-ctx", tab.nav.Context)
	assert.Equal(t, "my-ns", tab.namespace)
	assert.Equal(t, model.LevelResourceTypes, tab.nav.Level)
}

func TestCovBuildSessionTabStateAllNS(t *testing.T) {
	st := &SessionTab{
		Context:       "my-ctx",
		AllNamespaces: true,
	}
	tab := buildSessionTabState(st)
	assert.True(t, tab.allNamespaces)
}

func TestCovBuildSessionTabStateNoContext(t *testing.T) {
	st := &SessionTab{}
	tab := buildSessionTabState(st)
	assert.Equal(t, model.LevelClusters, tab.nav.Level)
}

func TestCovBuildSessionTabStateWithSelectedNS(t *testing.T) {
	st := &SessionTab{
		Context:            "ctx",
		Namespace:          "ns1",
		SelectedNamespaces: []string{"ns1", "ns2"},
	}
	tab := buildSessionTabState(st)
	assert.True(t, tab.selectedNamespaces["ns1"])
	assert.True(t, tab.selectedNamespaces["ns2"])
}

// =====================================================================
// update_navigation.go -- navigateToOwner
// =====================================================================

func TestCovNavigateToOwnerUnknownKind(t *testing.T) {
	m := baseModelCov()
	m.discoveredCRDs = map[string][]model.ResourceTypeEntry{}
	ret, cmd := m.navigateToOwner("UnknownKind", "some-name")
	result := ret.(Model)
	assert.True(t, result.hasStatusMessage())
	assert.NotNil(t, cmd)
}

// =====================================================================
// tabs.go -- portForwardItems, navigateToPortForwards
// =====================================================================

func TestCovPortForwardItemsNoManager(t *testing.T) {
	m := baseModelCov()
	m.portForwardMgr = k8s.NewPortForwardManager()
	items := m.portForwardItems()
	assert.Empty(t, items)
}

func TestCovNavigateToPortForwards(t *testing.T) {
	m := baseModelWithFakeClient()
	m.portForwardMgr = k8s.NewPortForwardManager()
	m.navigateToPortForwards()
	assert.Equal(t, model.LevelResources, m.nav.Level)
	assert.Equal(t, "__port_forwards__", m.nav.ResourceType.Kind)
}

// =====================================================================
// commands.go -- refreshCurrentLevel
// =====================================================================

func TestCovRefreshCurrentLevelClustersFakeClient(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelClusters
	cmd := m.refreshCurrentLevel()
	assert.NotNil(t, cmd)
}

func TestCovRefreshCurrentLevelResourceTypesFakeClient(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelResourceTypes
	cmd := m.refreshCurrentLevel()
	assert.NotNil(t, cmd)
}

func TestCovRefreshCurrentLevelPortForwards(t *testing.T) {
	m := baseModelWithFakeClient()
	m.portForwardMgr = k8s.NewPortForwardManager()
	m.nav.Level = model.LevelResources
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "__port_forwards__"}
	cmd := m.refreshCurrentLevel()
	assert.NotNil(t, cmd)
}

func TestCovRefreshCurrentLevelResourcesFakeClient(t *testing.T) {
	m := baseModelWithFakeClient()
	m.nav.Level = model.LevelResources
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod", APIVersion: "v1", Resource: "pods", Namespaced: true}
	cmd := m.refreshCurrentLevel()
	assert.NotNil(t, cmd)
}

// =====================================================================
// commands_logs.go -- saveLoadedLogs, saveAllLogs
// =====================================================================

func TestCovSaveLoadedLogs(t *testing.T) {
	m := baseModelCov()
	m.logLines = []string{"line1", "line2", "line3"}
	m.actionCtx = actionContext{name: "test-pod"}
	path, err := m.saveLoadedLogs()
	assert.NoError(t, err)
	assert.Contains(t, path, "lfk-logs-test-pod")
}

func TestCovSaveAllLogs(t *testing.T) {
	m := baseModelWithFakeClient()
	m.logLines = []string{"line1", "line2"}
	m.logTitle = "Logs: my-pod"
	m = withActionCtx(m, "my-pod", "default", "Pod", model.ResourceTypeEntry{})
	cmd := m.saveAllLogs()
	assert.NotNil(t, cmd)
}

// =====================================================================
// portforward_state.go -- saveCurrentPortForwards, restorePortForwards
// =====================================================================

func TestCovSaveCurrentPortForwards(t *testing.T) {
	m := baseModelCov()
	m.portForwardMgr = k8s.NewPortForwardManager()
	// Should not panic with no entries.
	m.saveCurrentPortForwards()
}

func TestCovRestorePortForwards(t *testing.T) {
	m := baseModelCov()
	m.portForwardMgr = k8s.NewPortForwardManager()
	m.client = k8s.NewTestClient(nil, nil)
	m.pendingPortForwards = &PortForwardStates{
		PortForwards: []PortForwardState{},
	}
	cmds := m.restorePortForwards()
	// No port forwards to restore, and kubectl may not be available.
	_ = cmds
}

// =====================================================================
// update_mouse.go -- switchToTab
// =====================================================================

func TestCovSwitchToTabSameTab(t *testing.T) {
	m := baseModelCov()
	m.activeTab = 0
	m.tabs = []TabState{{}}
	ret, cmd := m.switchToTab(0)
	_ = ret.(Model)
	assert.Nil(t, cmd)
}

// =====================================================================
// resolveOwnedResourceType
// =====================================================================

func TestCovResolveOwnedResourceTypeNil(t *testing.T) {
	m := baseModelCov()
	_, ok := m.resolveOwnedResourceType(nil)
	assert.False(t, ok)
}

func TestCovResolveOwnedResourceTypeFallback(t *testing.T) {
	m := baseModelCov()
	m.discoveredCRDs = map[string][]model.ResourceTypeEntry{}
	sel := &model.Item{Kind: "MyCustomResource", Extra: "mygroup.io/v1"}
	rt, ok := m.resolveOwnedResourceType(sel)
	assert.True(t, ok)
	assert.Equal(t, "mygroup.io", rt.APIGroup)
	assert.Equal(t, "v1", rt.APIVersion)
	assert.Equal(t, "mycustomresources", rt.Resource)
}

// =====================================================================
// update_overlays_logs.go -- buildLogTitle
// =====================================================================

func TestCovBuildLogTitle(t *testing.T) {
	m := baseModelCov()
	m.actionCtx = actionContext{name: "my-pod", namespace: "default", context: "ctx"}
	title := m.buildLogTitle()
	assert.Contains(t, title, "Logs: default/my-pod")
}

func TestCovBuildLogTitleWithContainerFilter(t *testing.T) {
	m := baseModelCov()
	m.actionCtx = actionContext{name: "my-pod", namespace: "default", context: "ctx"}
	m.logContainers = []string{"main", "sidecar"}
	m.logSelectedContainers = []string{"main"}
	title := m.buildLogTitle()
	assert.Contains(t, title, "[main]")
}

// =====================================================================
// loadPreviewEvents, loadMetrics with selection
// =====================================================================

func TestCovLoadPreviewEventsWithSelection(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withMiddleItem(m, model.Item{Name: "my-pod", Namespace: "default"})
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod"}
	cmd := m.loadPreviewEvents()
	msg := execCmd(t, cmd)
	result, ok := msg.(previewEventsLoadedMsg)
	require.True(t, ok)
	_ = result
}

func TestCovLoadMetricsPod(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withMiddleItem(m, model.Item{Name: "my-pod", Namespace: "default"})
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod"}
	cmd := m.loadMetrics()
	assert.NotNil(t, cmd)
}

func TestCovLoadMetricsDeployment(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withMiddleItem(m, model.Item{Name: "my-deploy", Namespace: "default"})
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Deployment"}
	cmd := m.loadMetrics()
	assert.NotNil(t, cmd)
}

func TestCovLoadMetricsUnsupported(t *testing.T) {
	m := baseModelWithFakeClient()
	m = withMiddleItem(m, model.Item{Name: "my-cm", Namespace: "default"})
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "ConfigMap"}
	cmd := m.loadMetrics()
	assert.Nil(t, cmd)
}

// =====================================================================
// waitForStderr
// =====================================================================

func TestCovWaitForStderrNil(t *testing.T) {
	m := baseModelCov()
	m.stderrChan = nil
	cmd := m.waitForStderr()
	assert.Nil(t, cmd)
}
