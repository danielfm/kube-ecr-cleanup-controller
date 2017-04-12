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
