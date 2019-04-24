package processor

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/danielfm/kube-ecr-cleanup-controller/pkg/core"
	apiv1 "k8s.io/api/core/v1"
)

// mockKubeClient is used to verify that the Kubernetes client is being called
// with the correct arguments, and that the return values are being handled
// correctly by its consumers.
type mockKubeClient struct {
	t *testing.T

	expectedNamespace []string

	listAllPodsResult []*apiv1.Pod
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

	expectedImagesToRemove []*ecr.ImageDetail
	batchRemoveImagesError error
}

func (m *mockKubeClient) ListAllPods(namespace []*string) ([]*apiv1.Pod, error) {
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

func (m *mockECRClient) ListRepositories(repositoryNames []*string, registryID *string) ([]*ecr.Repository, error) {
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

func (m *mockECRClient) ListImages(repositoryName *string, registryID *string) ([]*ecr.ImageDetail, error) {
	if m.expectedImagesRepositoryName != *repositoryName {
		m.t.Errorf("Expected repository name to be %v, but was %v", m.expectedImagesRepositoryName, *repositoryName)
	}

	return m.listImagesResult, m.listImagesError
}

func (m *mockECRClient) BatchRemoveImages(images []*ecr.ImageDetail) error {
	if len(images) != len(m.expectedImagesToRemove) {
		m.t.Errorf("Expected images to contain %d elements, but it contains %d", len(m.expectedImagesToRemove), len(images))
	}

	for i := range images {
		if *images[i].ImageDigest != *m.expectedImagesToRemove[i].ImageDigest {
			m.t.Errorf("Expected image digest at index %d to be %v, but was %v", i, m.expectedImagesToRemove[i].ImageDigest, *images[i].ImageDigest)
		}
	}

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

	task := &core.CleanupTask{
		KubeNamespaces: []*string{&namespace},
	}

	errs := RemoveOldImages(task, kubeClient, nil)

	if len(errs) != 1 {
		t.Errorf("Expected errors to contain 1 element, but it contains %d", len(errs))
	}
}

func TestRemoveOldImagesWithECRListRepositoriesError(t *testing.T) {
	namespace, repoName := "namespace", "repo"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},
		listAllPodsResult: []*apiv1.Pod{
			{},
		},
	}

	ecrClient := &mockECRClient{
		t: t,

		expectedRepositoryNames: []string{repoName},
		listRepositoriesResult:  nil,
		listRepositoriesError:   fmt.Errorf(""),
	}

	task := &core.CleanupTask{
		KubeNamespaces:  []*string{&namespace},
		EcrRepositories: []*string{&repoName},
	}

	errs := RemoveOldImages(task, kubeClient, ecrClient)

	if len(errs) != 1 {
		t.Errorf("Expected errors to contain 1 element, but it contains %d", len(errs))
	}
}

func TestRemoveOldImagesWithECRListImagesError(t *testing.T) {
	namespace, repoName := "namespace", "repo"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},
		listAllPodsResult: []*apiv1.Pod{
			{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
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

	task := &core.CleanupTask{
		KubeNamespaces:  []*string{&namespace},
		EcrRepositories: []*string{&repoName},
		MaxImages:       1,
	}

	errs := RemoveOldImages(task, kubeClient, ecrClient)

	if len(errs) != 1 {
		t.Errorf("Expected errors to contain 1 element, but it contains %d", len(errs))
	}
}

func TestRemoveOldImagesWithoutOldImagesToRemove(t *testing.T) {
	namespace, repoName, imageDigest := "namespace", "repo", "image-digest"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},
		listAllPodsResult: []*apiv1.Pod{
			{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
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
		listImagesResult: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},
	}

	task := &core.CleanupTask{
		KubeNamespaces:  []*string{&namespace},
		EcrRepositories: []*string{&repoName},

		// No need to clean up any images
		MaxImages: 1000,
	}

	errs := RemoveOldImages(task, kubeClient, ecrClient)

	if len(errs) != 0 {
		t.Errorf("Expected errors to be empty, but is %q", errs)
	}
}

func TestRemoveOldImagesWithECRBatchRemoveImagesError(t *testing.T) {
	namespace, repoName, imageDigest := "namespace", "repo", "image-digest"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},
		listAllPodsResult: []*apiv1.Pod{
			{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
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
		listImagesResult: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},

		expectedImagesToRemove: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},
		batchRemoveImagesError: fmt.Errorf(""),
	}

	task := &core.CleanupTask{
		KubeNamespaces:  []*string{&namespace},
		EcrRepositories: []*string{&repoName},

		// Will cause the image to be deleted
		MaxImages: 0,
	}

	errs := RemoveOldImages(task, kubeClient, ecrClient)

	if len(errs) == 0 {
		t.Errorf("Expected errors to contain 1 element, but it contains %d", len(errs))
	}
}

func TestRemoveOldImagesWithKeepFilter(t *testing.T) {
	namespace, repoName, imageDigest := "namespace", "repo", "image-digest"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},
		listAllPodsResult: []*apiv1.Pod{
			{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Image: "id.dkr.ecr.region.amazonaws.com/repo:tag-1",
						},
					},
				},
			},
			{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Image: "id.dkr.ecr.region.amazonaws.com/repo:keep-1",
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
		listImagesResult: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},

		expectedImagesToRemove: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},
	}

	keep := "keep"
	task := &core.CleanupTask{
		KubeNamespaces:  []*string{&namespace},
		EcrRepositories: []*string{&repoName},

		// Will cause the image to be deleted
		MaxImages:   0,
		KeepFilters: []*string{&keep},
	}

	errs := RemoveOldImages(task, kubeClient, ecrClient)

	if len(errs) != 0 {
		t.Errorf("Expected errors to be empty, but is %q", errs)
	}
}

func TestRemoveOldImages(t *testing.T) {
	namespace, repoName, imageDigest := "namespace", "repo", "image-digest"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},
		listAllPodsResult: []*apiv1.Pod{
			{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
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
		listImagesResult: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},

		expectedImagesToRemove: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},
	}

	task := &core.CleanupTask{
		KubeNamespaces:  []*string{&namespace},
		EcrRepositories: []*string{&repoName},

		// Will cause the image to be deleted
		MaxImages: 0,
	}

	errs := RemoveOldImages(task, kubeClient, ecrClient)

	if len(errs) != 0 {
		t.Errorf("Expected errors to be empty, but is %q", errs)
	}
}

func TestRemoveOldImagesWithDryRun(t *testing.T) {
	namespace, repoName, imageDigest := "namespace", "repo", "image-digest"
	kubeClient := &mockKubeClient{
		t: t,

		expectedNamespace: []string{namespace},
		listAllPodsResult: []*apiv1.Pod{
			{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
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
		listImagesResult: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},

		expectedImagesToRemove: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},
	}

	task := &core.CleanupTask{
		KubeNamespaces:  []*string{&namespace},
		EcrRepositories: []*string{&repoName},
		DryRun:          true,

		// Will cause the image to be deleted
		MaxImages: 0,
	}

	errs := RemoveOldImages(task, kubeClient, ecrClient)

	if len(errs) != 0 {
		t.Errorf("Expected errors to be empty, but is %q", errs)
	}
}
