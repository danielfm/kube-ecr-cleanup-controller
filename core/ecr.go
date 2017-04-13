package core

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

type ECRClientImpl struct {
	ECRClient ecriface.ECRAPI
}

// ECRClient defines the expected interface of any object capable of
// listing and removing images from a ECR repository.
type ECRClient interface {
	ListRepositories(repositoryNames []*string) ([]*ecr.Repository, error)
	ListImages(repositoryName *string) ([]*ecr.ImageDetail, error)
	BatchRemoveImages(repositoryName *string, images []*ecr.ImageDetail) error
}

// ImagesByPushDate lets us sort ECR images by push date so that we can
// delete old images.
type ImagesByPushDate []*ecr.ImageDetail

func (slice ImagesByPushDate) Len() int {
	return len(slice)
}

func (slice ImagesByPushDate) Less(i, j int) bool {
	ti := *slice[i].ImagePushedAt
	tj := *slice[j].ImagePushedAt
	return ti.Before(tj)
}

func (slice ImagesByPushDate) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// NewECRClient returns a new client for interacting with the ECR API. The
// credentials are retrieved from environment variables or from the
// `~/.aws/credentials` file.
func NewECRClient(region string) *ECRClientImpl {
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
		})

	awsConfig := aws.NewConfig()
	awsConfig.WithCredentials(creds)
	awsConfig.WithRegion(region)

	sess := session.New(awsConfig)

	return &ECRClientImpl{
		ECRClient: ecr.New(sess),
	}
}

// ListRepositories returns the data belonging to the given repository names.
func (c *ECRClientImpl) ListRepositories(repositoryNames []*string) ([]*ecr.Repository, error) {
	repos := []*ecr.Repository{}

	if len(repositoryNames) == 0 {
		return repos, nil
	}

	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: repositoryNames,
	}

	callback := func(page *ecr.DescribeRepositoriesOutput, lastPage bool) bool {
		repos = append(repos, page.Repositories...)
		return !lastPage
	}

	err := c.ECRClient.DescribeRepositoriesPages(input, callback)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// ListImages returns data from all images stored in the repository identified
// by the given repository name.
func (c *ECRClientImpl) ListImages(repositoryName *string) ([]*ecr.ImageDetail, error) {
	images := []*ecr.ImageDetail{}

	if repositoryName == nil {
		return images, nil
	}

	input := &ecr.DescribeImagesInput{
		RepositoryName: repositoryName,
	}

	callback := func(page *ecr.DescribeImagesOutput, lastPage bool) bool {
		images = append(images, page.ImageDetails...)
		return !lastPage
	}

	err := c.ECRClient.DescribeImagesPages(input, callback)
	if err != nil {
		return nil, err
	}

	return images, nil
}

// BatchRemoveImages deletes all the given images from the repository identified
// by the given repository name in one go.
func (c *ECRClientImpl) BatchRemoveImages(repositoryName *string, images []*ecr.ImageDetail) error {

	// No images to be removed
	if len(images) == 0 {
		return nil
	}

	imageIds := make([]*ecr.ImageIdentifier, len(images))

	for i := range images {
		imageIds[i] = &ecr.ImageIdentifier{
			ImageDigest: images[i].ImageDigest,
		}
	}

	input := &ecr.BatchDeleteImageInput{
		RepositoryName: repositoryName,
		ImageIds:       imageIds,
	}

	output, err := c.ECRClient.BatchDeleteImage(input)
	if err != nil {
		return err
	}

	// Aggregates all failures in a single error message for convenience
	if len(output.Failures) > 0 {
		msg := "Failed to remove the following images: \n\n"

		for _, failure := range output.Failures {
			msg += fmt.Sprintf("Error %d (%s): %s\n", failure.FailureCode, failure.ImageId.ImageDigest, failure.FailureReason)
		}

		return fmt.Errorf(msg)
	}

	return nil
}

// SortImagesByPushDate uses the `ImagesByPushDate` type to sort the given slice
// of ECR image objects.
func SortImagesByPushDate(images []*ecr.ImageDetail) {
	var imagesByDate ImagesByPushDate
	imagesByDate = images

	sort.Sort(imagesByDate)
}

// FilterOldUnusedImages goes through the given list of ECR images and returns
// another list of images (giving priority to older images) that are not in use.
// The filtered images, when removed, will bring the number of images stored in
// the repository down to a value as close to the number specified in `keepMax`
// as possible.
func FilterOldUnusedImages(keepMax int, repoImages []*ecr.ImageDetail, tagsInUse []string) []*ecr.ImageDetail {
	usedImagesFound := 0
	unusedImages := []*ecr.ImageDetail{}

	// There's no need to remove any images for now
	if keepMax >= len(repoImages) {
		return []*ecr.ImageDetail{}
	}

repoImagesLoop:
	for _, repoImage := range repoImages {
		for _, tag := range repoImage.ImageTags {
			for _, tagInUse := range tagsInUse {
				if tagInUse == *tag {
					usedImagesFound++
					continue repoImagesLoop
				}
			}
		}

		unusedImages = append(unusedImages, repoImage)
	}

	SortImagesByPushDate(unusedImages)

	lastImageIdx := len(unusedImages) - keepMax + usedImagesFound
	if lastImageIdx > len(unusedImages) {
		lastImageIdx = len(unusedImages)
	}

	// Returns the oldest images that are not in use that, when deleted,
	// will bring the number of images down to the specified theshold
	return unusedImages[:lastImageIdx]
}
