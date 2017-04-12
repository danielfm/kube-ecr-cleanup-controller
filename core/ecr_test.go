package core

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ecr"
)

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

func TestFilterOldUnusedImages(t *testing.T) {
	tags := []string{"tag-1", "tag-2", "tag-3", "tag-4", "tag-5"}

	imageTags := [][]*string{
		{&tags[0], &tags[1]},
		{&tags[2], &tags[3]},
		{&tags[4]},
	}

	orderedTime := []time.Time{
		time.Unix(0, 0),
		time.Unix(1, 0),
		time.Unix(2, 0),
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

		// Should return no images as they're all being used
		{
			keepMax:   0,
			tagsInUse: []string{"tag-1", "tag-3", "tag-5"},
			images: []*ecr.ImageDetail{
				{
					ImagePushedAt: &orderedTime[2],
					ImageTags:     imageTags[2],
				},
				{
					ImagePushedAt: &orderedTime[1],
					ImageTags:     imageTags[1],
				},
				{
					ImagePushedAt: &orderedTime[0],
					ImageTags:     imageTags[0],
				},
			},
			oldImages: []*ecr.ImageDetail{},
		},

		// TODO: Add more cases
	}

	for _, testCase := range testCases {
		filtered := FilterOldUnusedImages(testCase.keepMax, testCase.images, testCase.tagsInUse)

		if len(filtered) != len(testCase.oldImages) {
			t.Errorf("Expected list of old images to have %d items, but it has %d", len(testCase.oldImages), len(filtered))
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
