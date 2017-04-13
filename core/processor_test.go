package core

import (
	"fmt"
	"testing"

	"k8s.io/client-go/pkg/api/v1"
)

// mockECRClient is used to verify that the Kubernetes client is being called
// with the correct arguments, and that the return values are being handled
// correctly by its consumers.
type mockKubeClient struct {
	t *testing.T

	expectedNamespace []string

	listAllPodsResult []*v1.Pod
	listAllPodsError  error
}

func (m *mockKubeClient) ListAllPods(namespace []*string) ([]*v1.Pod, error) {
	if len(namespace) != len(m.expectedNamespace) {
		m.t.Errorf("Expected namespaces to contain %d elements, but it contains %d", len(m.expectedNamespace), len(namespace))
	}

	for i := range namespace {
		if *namespace[i] != m.expectedNamespace[i] {
			m.t.Errorf("Expected namespace at index %d to be %v, but was %v", i, m.expectedNamespace[i], *namespace[i])
		}
	}

	return m.listAllPodsResult, m.listAllPodsError
}

func TestRemoveOldImagesWithKubeListPodsError(t *testing.T) {
	namespace := "namespace"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},

		listAllPodsResult: nil,
		listAllPodsError:  fmt.Errorf(""),
	}

	task := &CleanupTask{
		KubeNamespaces: []*string{&namespace},
	}

	errs := task.RemoveOldImages(kubeClient, nil)

	if len(errs) != 1 {
		t.Errorf("Expected errors to contain 1 element, but it contains %d", len(errs))
	}
}

func TestRemoveOldImagesWithECRListRepositoriesError(t *testing.T) {
	// TODO
}

func TestRemoveOldImagesWithECRListImagesError(t *testing.T) {
	// TODO
}

func TestRemoveOldImagesWithECRBatchRemoveImagesError(t *testing.T) {
	// TODO
}

func TestRemoveOldImages(t *testing.T) {
	// TODO
}
