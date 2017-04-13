package core

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"

	"k8s.io/client-go/pkg/api/v1"
)

// mockKubeClient is used to verify that the Kubernetes client is being called
// with the correct arguments, and that the return values are being handled
// correctly by its consumers.
type mockKubeClient struct {
	t *testing.T

	expectedNamespace []string

	listAllPodsResult []*v1.Pod
	listAllPodsError  error
}

// mockECRClient is used to verify that the Kubernetes client is being called
// with the correct arguments, and that the return values are being handled
// correctly by its consumers.
type mockECRClient struct {
	t *testing.T

	expectedRepositoryNames []string

	listRepositoriesResult []*ecr.Repository
	listRepositoriesError  error

	expectedImagesRepositoryName string
	listImagesResult             []*ecr.ImageDetail
	listImagesError              error

	batchRemoveImagesError error
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

func (m *mockECRClient) ListRepositories(repositoryNames []*string) ([]*ecr.Repository, error) {
	if len(repositoryNames) != len(m.expectedRepositoryNames) {
		m.t.Errorf("Expected repository names to contain %d elements, but it contains %d", len(m.expectedRepositoryNames), len(repositoryNames))
	}

	for i := range repositoryNames {
		if *repositoryNames[i] != m.expectedRepositoryNames[i] {
			m.t.Errorf("Expected repository name at index %d to be %v, but was %v", i, m.expectedRepositoryNames[i], *repositoryNames[i])
		}
	}

	return m.listRepositoriesResult, m.listRepositoriesError
}

func (m *mockECRClient) ListImages(repositoryName *string) ([]*ecr.ImageDetail, error) {
	if m.expectedImagesRepositoryName != *repositoryName {
		m.t.Errorf("Expected repository name to be %v, but was %v", m.expectedImagesRepositoryName, *repositoryName)
	}

	return m.listImagesResult, m.listImagesError
}

func (m *mockECRClient) BatchRemoveImages(images []*ecr.ImageDetail) error {
	// TODO

	return m.batchRemoveImagesError
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
	namespace, repoName := "namespace", "repo"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},
		listAllPodsResult: []*v1.Pod{
			{},
		},
	}

	ecrClient := &mockECRClient{
		t: t,

		expectedRepositoryNames: []string{repoName},
		listRepositoriesResult:  nil,
		listRepositoriesError:   fmt.Errorf(""),
	}

	task := &CleanupTask{
		KubeNamespaces:  []*string{&namespace},
		EcrRepositories: []*string{&repoName},
	}

	errs := task.RemoveOldImages(kubeClient, ecrClient)

	if len(errs) != 1 {
		t.Errorf("Expected errors to contain 1 element, but it contains %d", len(errs))
	}
}

func TestRemoveOldImagesWithECRListImagesError(t *testing.T) {
	namespace, repoName := "namespace", "repo"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},
		listAllPodsResult: []*v1.Pod{
			{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "id.dkr.ecr.region.amazonaws.com/repo:tag-1",
						},
					},
				},
			},
		},
	}

	ecrClient := &mockECRClient{
		t: t,

		expectedRepositoryNames: []string{repoName},
		listRepositoriesResult: []*ecr.Repository{
			{
				RepositoryName: &repoName,
			},
		},

		expectedImagesRepositoryName: repoName,
		listImagesResult:             nil,
		listImagesError:              fmt.Errorf(""),
	}

	task := &CleanupTask{
		KubeNamespaces:  []*string{&namespace},
		EcrRepositories: []*string{&repoName},
		MaxImages:       1,
	}

	errs := task.RemoveOldImages(kubeClient, ecrClient)

	if len(errs) != 1 {
		t.Errorf("Expected errors to contain 1 element, but it contains %d", len(errs))
	}
}

func TestRemoveOldImagesWithoutOldImagesToRemove(t *testing.T) {
	// TODO
}

func TestRemoveOldImagesWithECRBatchRemoveImagesError(t *testing.T) {
	// TODO
}

func TestRemoveOldImages(t *testing.T) {
	// TODO
}
