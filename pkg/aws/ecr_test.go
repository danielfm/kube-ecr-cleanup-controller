package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

// mockAWSECRClient is used to verify that the ECR client is being called with the
// correct arguments, and that the return values are being handled correctly by
// its consumers.
type mockAWSECRClient struct {
	t *testing.T
	ecriface.ECRAPI

	expectedRepositoryNames []string
	expectedImageDigests    []string
	expectedRegistryID      *string

	outputError error
}

func (m *mockAWSECRClient) DescribeRepositoriesPages(input *ecr.DescribeRepositoriesInput, fn func(*ecr.DescribeRepositoriesOutput, bool) bool) error {
	if input == nil {
		m.t.Errorf("Unexpected nil input")
	}

	if len(input.RepositoryNames) != len(m.expectedRepositoryNames) {
		m.t.Errorf("Expected number of repository names was %d, but the actual value was %d", len(m.expectedRepositoryNames), len(input.RepositoryNames))
	}

	if input.RegistryId != m.expectedRegistryID {
		m.t.Errorf("Expected registry id of %v, but got %v", m.expectedRegistryID, input.RegistryId)
	}

	for i := range input.RepositoryNames {
		if *input.RepositoryNames[i] != m.expectedRepositoryNames[i] {
			m.t.Errorf("Expected repository name in idx %d to be %v, but was %v", i, m.expectedRepositoryNames[i], *input.RepositoryNames[i])
		}
	}

	repositoryName := "repo-name"
	page := &ecr.DescribeRepositoriesOutput{
		Repositories: []*ecr.Repository{
			{
				RepositoryName: &repositoryName,
			},
		},
	}

	// There's two pages, so the function must return true
	if fn(page, false) != true {
		m.t.Errorf("Expected callback to return true for first page, but returned false")
	}

	// This is the last page, so the function must return false
	if fn(page, true) != false {
		m.t.Errorf("Expected callback to return true for last page, but returned true")
	}

	return m.outputError
}

func (m *mockAWSECRClient) DescribeImagesPages(input *ecr.DescribeImagesInput, fn func(*ecr.DescribeImagesOutput, bool) bool) error {
	if input == nil {
		m.t.Errorf("Unexpected nil input")
	}

	if *input.RepositoryName != m.expectedRepositoryNames[0] {
		m.t.Errorf("Expected repository name to be %s, but was %s", m.expectedRepositoryNames[0], *input.RepositoryName)
	}

	if input.RegistryId != m.expectedRegistryID {
		m.t.Errorf("Expected registry id of %v, but got %v", m.expectedRegistryID, input.RegistryId)
	}

	imageDigest := "image-digest"
	page := &ecr.DescribeImagesOutput{
		ImageDetails: []*ecr.ImageDetail{
			{
				ImageDigest: &imageDigest,
			},
		},
	}

	// There's two pages, so the function must return true
	if fn(page, false) != true {
		m.t.Errorf("Expected callback to return true for first page, but returned false")
	}

	// This is the last page, so the function must return false
	if fn(page, true) != false {
		m.t.Errorf("Expected callback to return true for last page, but returned true")
	}

	return m.outputError
}

func (m *mockAWSECRClient) BatchDeleteImage(input *ecr.BatchDeleteImageInput) (*ecr.BatchDeleteImageOutput, error) {
	if input == nil {
		m.t.Errorf("Unexpected nil input")
	}

	if *input.RepositoryName != m.expectedRepositoryNames[0] {
		m.t.Errorf("Expected repository name to be %s, but was %s", m.expectedRepositoryNames[0], *input.RepositoryName)
	}

	if len(input.ImageIds) != len(m.expectedImageDigests) {
		m.t.Errorf("Expected delete with %d images, but got %d", len(m.expectedImageDigests), len(input.ImageIds))
	}

	for i := range input.ImageIds {
		if *input.ImageIds[i].ImageDigest != m.expectedImageDigests[i] {
			m.t.Errorf("Expected image digest of image in idx %d to be %v, but was %v", i, m.expectedImageDigests, *input.ImageIds[i].ImageDigest)
		}
	}

	return nil, m.outputError
}

func TestSortImagesByPushDate(t *testing.T) {
	orderedTime := []time.Time{
		time.Unix(0, 0),
		time.Unix(1, 0),
		time.Unix(2, 0),
	}

	ecrImages := []*ecr.ImageDetail{
		{
			ImagePushedAt: &orderedTime[2],
		},
		{
			ImagePushedAt: &orderedTime[1],
		},
		{
			ImagePushedAt: &orderedTime[0],
		},
	}

	SortImagesByPushDate(ecrImages)

	if len(ecrImages) != 3 {
		t.Errorf("Expected image list to remain with 3 elements, but the size is now %d", len(ecrImages))
	}

	for i := range ecrImages {
		actualDate := *ecrImages[i].ImagePushedAt
		expectedDate := time.Unix(int64(i), 0)

		if *ecrImages[i].ImagePushedAt != expectedDate {
			t.Errorf("Expected ercImages[%d] timestamp to be %+v, but was %+v", i, expectedDate, actualDate)
		}
	}
}

func TestListRepositoriesWithEmptyRepos(t *testing.T) {
	client := ECRClientImpl{
		ECRClient: nil, // Should not interact with the ECR client
	}

	repos, err := client.ListRepositories([]*string{}, nil)

	if len(repos) != 0 {
		t.Errorf("Expected repos to be empty, but was not: %q", repos)
	}

	if err != nil {
		t.Errorf("Expected error to be nil, but was %v", err)
	}
}

func TestListRepositoriesError(t *testing.T) {
	repoNames := []string{"repo-1"}

	client := ECRClientImpl{
		ECRClient: &mockAWSECRClient{
			t: t,

			expectedRepositoryNames: repoNames,
			outputError:             fmt.Errorf(""),
		},
	}

	repos, err := client.ListRepositories([]*string{&repoNames[0]}, nil)

	if repos != nil {
		t.Errorf("Expected repos to be nil, but was %v", repos)
	}

	if err == nil {
		t.Errorf("Expected error not to be nil, but it was")
	}
}

func TestListRepositories(t *testing.T) {
	repoNames := []string{"repo-1"}

	client := ECRClientImpl{
		ECRClient: &mockAWSECRClient{
			t: t,

			expectedRepositoryNames: repoNames,
		},
	}

	repos, err := client.ListRepositories([]*string{&repoNames[0]}, nil)

	if err != nil {
		t.Errorf("Expected error to be nil, but it was: %v", err)
	}

	if repos == nil {
		t.Errorf("Expected repos not to be nil, but it was")
	}

	if len(repos) != 2 {
		t.Errorf("Expected repos to contain 2 items, but it contains: %q", repos)
	}
}

func TestListRepositoriesWithRegistryId(t *testing.T) {
	repoNames := []string{"repo-1"}
	registryID := "123456789012"

	client := ECRClientImpl{
		ECRClient: &mockAWSECRClient{
			t: t,

			expectedRepositoryNames: repoNames,
			expectedRegistryID:      &registryID,
		},
	}

	repos, err := client.ListRepositories([]*string{&repoNames[0]}, &registryID)

	if err != nil {
		t.Errorf("Expected error to be nil, but it was: %v", err)
	}

	if repos == nil {
		t.Errorf("Expected repos not to be nil, but it was")
	}

	if len(repos) != 2 {
		t.Errorf("Expected repos to contain 2 items, but it contains: %q", repos)
	}
}

func TestListImagesWithNilRepositoryName(t *testing.T) {
	client := ECRClientImpl{
		ECRClient: nil, // Should not interact with the ECR client
	}

	images, err := client.ListImages(nil, nil)

	if len(images) != 0 {
		t.Errorf("Expected images to be empty, but was not: %q", images)
	}

	if err != nil {
		t.Errorf("Expected error to be nil, but was %v", err)
	}
}

func TestListImagesError(t *testing.T) {
	repoName := "repo-1"

	client := ECRClientImpl{
		ECRClient: &mockAWSECRClient{
			t: t,

			expectedRepositoryNames: []string{repoName},

			outputError: fmt.Errorf(""),
		},
	}

	images, err := client.ListImages(&repoName, nil)

	if images != nil {
		t.Errorf("Expected images to be nil, but was %v", images)
	}

	if err == nil {
		t.Errorf("Expected error not to be nil, but it was")
	}
}

func TestListImages(t *testing.T) {
	repoName := "repo-1"

	client := ECRClientImpl{
		ECRClient: &mockAWSECRClient{
			t: t,

			expectedRepositoryNames: []string{repoName},
		},
	}

	images, err := client.ListImages(&repoName, nil)

	if err != nil {
		t.Errorf("Expected error to be nil, but it was: %v", err)
	}

	if images == nil {
		t.Errorf("Expected images not to be nil, but it was")
	}

	if len(images) != 2 {
		t.Errorf("Expected images to contain 2 items, but it contains: %q", images)
	}
}

func TestListImagesWithRegistryId(t *testing.T) {
	repoName := "repo-1"
	registryID := "123456789012"

	client := ECRClientImpl{
		ECRClient: &mockAWSECRClient{
			t: t,

			expectedRepositoryNames: []string{repoName},
			expectedRegistryID:      &registryID,
		},
	}

	images, err := client.ListImages(&repoName, &registryID)

	if err != nil {
		t.Errorf("Expected error to be nil, but it was: %v", err)
	}

	if images == nil {
		t.Errorf("Expected images not to be nil, but it was")
	}

	if len(images) != 2 {
		t.Errorf("Expected images to contain 2 items, but it contains: %q", images)
	}
}

func TestBatchRemoveImagesWithEmptyImages(t *testing.T) {
	client := ECRClientImpl{
		ECRClient: nil, // Should not interact with the ECR client
	}

	err := client.BatchRemoveImages([]*ecr.ImageDetail{})

	if err != nil {
		t.Errorf("Expected error to be nil, but was %v", err)
	}
}

func TestBatchRemoveImagesWithTooManyImages(t *testing.T) {
	client := ECRClientImpl{
		ECRClient: nil, // Should not interact with the ECR client
	}

	err := client.BatchRemoveImages(make([]*ecr.ImageDetail, 101))

	if err == nil {
		t.Errorf("Expected error not to be nil, but it was")
	}
}

func TestBatchRemoveImagesFromDifferentRepos(t *testing.T) {
	repoNames := []string{"repo-1", "repo-2"}

	client := ECRClientImpl{
		ECRClient: nil, // Should not interact with the ECR client
	}

	err := client.BatchRemoveImages([]*ecr.ImageDetail{
		{
			RepositoryName: &repoNames[0],
		},
		{
			RepositoryName: &repoNames[1],
		},
	})

	if err == nil {
		t.Errorf("Expected error not to be nil, but it was")
	}
}

func TestBatchRemoveImagesError(t *testing.T) {
	repoName, digest := "repo-1", "digest-1"

	images := []*ecr.ImageDetail{
		{
			ImageDigest:    &digest,
			RepositoryName: &repoName,
		},
	}

	client := ECRClientImpl{
		ECRClient: &mockAWSECRClient{
			t: t,

			expectedRepositoryNames: []string{repoName},
			expectedImageDigests:    []string{digest},

			outputError: fmt.Errorf(""),
		},
	}

	err := client.BatchRemoveImages(images)

	if err == nil {
		t.Errorf("Expected error not to be nil, but it was")
	}
}

func TestBatchRemoveImages(t *testing.T) {
	repoName, digest := "repo-1", "digest-1"

	images := []*ecr.ImageDetail{
		{
			ImageDigest:    &digest,
			RepositoryName: &repoName,
		},
	}

	client := ECRClientImpl{
		ECRClient: &mockAWSECRClient{
			t: t,

			expectedRepositoryNames: []string{repoName},
			expectedImageDigests:    []string{digest},
		},
	}

	err := client.BatchRemoveImages(images)

	if err != nil {
		t.Errorf("Expected error to be nil, but was %v", err)
	}
}

func TestFilterOldUnusedImages(t *testing.T) {
	latestTag := "latest"
	tags := []string{"tag-1", "tag-2", "tag-3", "tag-4", "tag-5"}

	orderedTime := []time.Time{
		time.Unix(0, 0),
		time.Unix(1, 0),
		time.Unix(2, 0),
	}

	tooManyImages := make([]*ecr.ImageDetail, 1000)
	for i := range tooManyImages {
		tooManyImages[i] = &ecr.ImageDetail{
			ImagePushedAt: &orderedTime[0],
		}
	}

	oldImagesCapped := make([]*ecr.ImageDetail, 100)
	for i := range oldImagesCapped {
		oldImagesCapped[i] = &ecr.ImageDetail{
			ImagePushedAt: &orderedTime[0],
		}
	}

	testCases := []struct {
		keepMax   int
		tagsInUse []string
		images    []*ecr.ImageDetail
		oldImages []*ecr.ImageDetail
	}{

		// Should return no images
		{
			keepMax:   3,
			tagsInUse: []string{},
			images: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[2],
				},
				{
					ImagePushedAt: &orderedTime[1],
				},
				{
					ImagePushedAt: &orderedTime[0],
				},
			},
			oldImages: []*ecr.ImageDetail{},
		},

		// Should return the oldest image
		{
			keepMax:   2,
			tagsInUse: []string{},
			images: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[2],
				},
				{
					ImagePushedAt: &orderedTime[1],
				},
				{
					ImagePushedAt: &orderedTime[0],
				},
			},
			oldImages: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[0],
				},
			},
		},

		// Should return all images sorted by date
		{
			keepMax:   0,
			tagsInUse: []string{},
			images: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[2],
				},
				{
					ImagePushedAt: &orderedTime[1],
				},
				{
					ImagePushedAt: &orderedTime[0],
				},
			},
			oldImages: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[0],
				},
				{
					ImagePushedAt: &orderedTime[1],
				},
				{
					ImagePushedAt: &orderedTime[2],
				},
			},
		},

		// Should return all images but the one with 'latest' tag
		{
			keepMax:   0,
			tagsInUse: []string{},
			images: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[2],
				},
				{
					ImagePushedAt: &orderedTime[1],
				},
				{
					ImagePushedAt: &orderedTime[0],
					ImageTags:     []*string{&latestTag},
				},
			},
			oldImages: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[1],
				},
				{
					ImagePushedAt: &orderedTime[2],
				},
			},
		},

		// Should return no images as they're all being used
		{
			keepMax:   0,
			tagsInUse: []string{tags[0], tags[2], tags[4]},
			images: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[2],
					ImageTags:     []*string{&tags[4]},
				},
				{
					ImagePushedAt: &orderedTime[1],
					ImageTags:     []*string{&tags[2]},
				},
				{
					ImagePushedAt: &orderedTime[0],
					ImageTags:     []*string{&tags[0]},
				},
			},
			oldImages: []*ecr.ImageDetail{},
		},

		// Should return all images but the oldest one which is in use
		{
			keepMax:   1,
			tagsInUse: []string{tags[0]},
			images: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[2],
					ImageTags:     []*string{&tags[2]},
				},
				{
					ImagePushedAt: &orderedTime[1],
					ImageTags:     []*string{&tags[1]},
				},
				{
					ImagePushedAt: &orderedTime[0],
					ImageTags:     []*string{&tags[0]},
				},
			},
			oldImages: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[1],
				},
				{
					ImagePushedAt: &orderedTime[2],
				},
			},
		},

		// Should return the newest image as the two oldest ones are in use
		{
			keepMax:   1,
			tagsInUse: []string{tags[0], tags[0], tags[1], tags[1]}, // Duplicate tag must be handled correctly
			images: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[2],
					ImageTags:     []*string{&tags[2], &tags[2]}, // Intentionally inconsistent metadata
				},
				{
					ImagePushedAt: &orderedTime[1],
					ImageTags:     []*string{&tags[1]},
				},
				{
					ImagePushedAt: &orderedTime[0],
					ImageTags:     []*string{&tags[0]},
				},
			},
			oldImages: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[2],
				},
			},
		},

		// Should limit the output to 100 images
		{
			keepMax:   0,
			tagsInUse: []string{},
			images:    tooManyImages,
			oldImages: oldImagesCapped,
		},
	}

	for _, testCase := range testCases {
		filtered := FilterOldUnusedImages(testCase.keepMax, testCase.images, testCase.tagsInUse)

		if len(filtered) != len(testCase.oldImages) {
			t.Errorf("Expected list of old images to have %d items, but it has %d:\n\nExpected: %+v\nActual: %+v", len(testCase.oldImages), len(filtered), testCase.oldImages, filtered)
		}

		for i := range filtered {
			actualDate := *filtered[i].ImagePushedAt
			expectedDate := *testCase.oldImages[i].ImagePushedAt

			if actualDate != expectedDate {
				t.Errorf("Expected filtered[%d] timestamp to be %+v, but was %+v", i, expectedDate, actualDate)
			}
		}
	}
}
