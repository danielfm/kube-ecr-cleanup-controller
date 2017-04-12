package core

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type ECRClientImpl struct {
	ECRClient *ecr.ECR
}

type ECRClient interface {
	ListRepositories(repositoryNames []*string) ([]*ecr.Repository, error)
	ListImages(repositoryName *string) ([]*ecr.ImageDetail, error)
	BatchRemoveImages(repositoryName *string, images []*ecr.ImageDetail) error
}

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

func NewECRClient(region string) (*ECRClientImpl, error) {
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
	}, nil
}

func (c *ECRClientImpl) ListRepositories(repositoryNames []*string) ([]*ecr.Repository, error) {
	repos := []*ecr.Repository{}

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

func (c *ECRClientImpl) ListImages(repositoryName *string) ([]*ecr.ImageDetail, error) {
	images := []*ecr.ImageDetail{}

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

	if len(output.Failures) > 0 {
		msg := "Failed to remove the following images: \n\n"

		for _, failure := range output.Failures {
			msg += fmt.Sprintf("Error %d (%s): %s\n", failure.FailureCode, failure.ImageId.ImageDigest, failure.FailureReason)
		}

		return fmt.Errorf(msg)
	}

	return nil
}

func SortImagesByPushDate(images []*ecr.ImageDetail) {
	var imagesByDate ImagesByPushDate
	imagesByDate = images

	sort.Sort(imagesByDate)
}

func FilterOldUnusedImages(keepMax int, repoImages []*ecr.ImageDetail, tagsInUse []string) []*ecr.ImageDetail {
	unusedImages := []*ecr.ImageDetail{}

	// There's no need to remove any images for now
	if keepMax > len(repoImages) {
		return []*ecr.ImageDetail{}
	}

repoImagesLoop:
	for _, repoImage := range repoImages {
		for _, tag := range repoImage.ImageTags {
			for _, tagInUse := range tagsInUse {
				if tagInUse == *tag {
					continue repoImagesLoop
				}
			}
		}

		unusedImages = append(unusedImages, repoImage)
	}

	SortImagesByPushDate(unusedImages)

	// Old unused images still within the maximum, so don't remove anything
	if keepMax > len(unusedImages) {
		return []*ecr.ImageDetail{}
	}

	// Returns the oldest images that are not in use that, when deleted,
	// will bring the number of images down to the specified theshold
	return unusedImages[:(len(unusedImages) - keepMax + 1)]
}
