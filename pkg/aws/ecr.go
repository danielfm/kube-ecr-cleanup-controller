package aws

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

const (
	batchRemoveMaxImages = 100
)

// ECRClientImpl provides an interface for mocking.
type ECRClientImpl struct {
	ECRClient ecriface.ECRAPI
}

// ECRClient defines the expected interface of any object capable of
// listing and removing images from a ECR repository.
type ECRClient interface {
	ListRepositories(repositoryNames []*string, registryID *string) ([]*ecr.Repository, error)
	ListImages(repositoryName *string, registryID *string) ([]*ecr.ImageDetail, error)
	BatchRemoveImages(images []*ecr.ImageDetail) error
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
	awsConfig := aws.NewConfig()
	awsConfig.WithRegion(region)

	sess := session.Must(
		session.NewSession(awsConfig),
	)

	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&stscreds.WebIdentityRoleProvider{},
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		})

	awsConfig.WithCredentials(creds)

	return &ECRClientImpl{
		ECRClient: ecr.New(sess),
	}
}

// ListRepositories returns the data belonging to the given repository names.
func (c *ECRClientImpl) ListRepositories(repositoryNames []*string, registryID *string) ([]*ecr.Repository, error) {
	repos := []*ecr.Repository{}

	if len(repositoryNames) == 0 {
		return repos, nil
	}

	// If the user has specified a registryID (account ID), then use it here.  If not
	// then don't set it so that the default will be assumed.
	input := &ecr.DescribeRepositoriesInput{}
	if registryID == nil {
		input = &ecr.DescribeRepositoriesInput{
			RepositoryNames: repositoryNames,
		}
	} else {
		input = &ecr.DescribeRepositoriesInput{
			RepositoryNames: repositoryNames,
			RegistryId:      registryID,
		}
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
func (c *ECRClientImpl) ListImages(repositoryName *string, registryID *string) ([]*ecr.ImageDetail, error) {
	images := []*ecr.ImageDetail{}

	if repositoryName == nil {
		return images, nil
	}

	// If the user has specified a registryID (account ID), then use it here.  If not
	// then don't set it so that the default will be assumed.
	input := &ecr.DescribeImagesInput{}
	if registryID == nil {
		input = &ecr.DescribeImagesInput{
			RepositoryName: repositoryName,
		}
	} else {
		input = &ecr.DescribeImagesInput{
			RepositoryName: repositoryName,
			RegistryId:     registryID,
		}
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

// BatchRemoveImages deletes all the given images in one go. All images must
// be stored in the same repository for this to work.
func (c *ECRClientImpl) BatchRemoveImages(images []*ecr.ImageDetail) error {

	// No images to be removed
	if len(images) == 0 {
		return nil
	}

	// Too many images to delete
	if len(images) > batchRemoveMaxImages {
		return fmt.Errorf("Only allows to remove %d images in a single call", batchRemoveMaxImages)
	}

	repositoryName := images[0].RepositoryName
	for i := range images {
		if *images[i].RepositoryName != *repositoryName {
			return fmt.Errorf("All images must belong to the same ECR repo")
		}
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

	_, err := c.ECRClient.BatchDeleteImage(input)
	if err != nil {
		return err
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
// This list will contain at most 100 images, which is the maximum number of
// images we are allowed to delete in a single API call to AWS.
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
			if *tag == "latest" {
				continue repoImagesLoop
			}

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

	// Only returns the 100 oldest unused images, which is the number of
	// images we are allowed to delete in a single API call
	if lastImageIdx > batchRemoveMaxImages {
		lastImageIdx = batchRemoveMaxImages
	}

	return unusedImages[:lastImageIdx]
}
