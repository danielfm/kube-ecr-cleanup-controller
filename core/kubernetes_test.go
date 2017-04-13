package core

import (
	"reflect"
	"testing"

	"k8s.io/client-go/pkg/api/v1"
)

func TestECRImagesFromPods(t *testing.T) {
	testCases := []struct {
		pods     []*v1.Pod
		expected map[string][]string
	}{
		// Different tagged images from different repos in the same pod
		{
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						InitContainers: []v1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-1",
							},
						},
						Containers: []v1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-2:tag-2",
							},
						},
					},
				},
			},
			expected: map[string][]string{
				"repo-1": []string{"tag-1"},
				"repo-2": []string{"tag-2"},
			},
		},

		// Same tagged image in the same pod
		{
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						InitContainers: []v1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-1",
							},
						},
						Containers: []v1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-1",
							},
						},
					},
				},
			},
			expected: map[string][]string{
				"repo-1": []string{"tag-1"},
			},
		},

		// Two tagged images from the same repo
		{
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-1",
							},
						},
					},
				},
				{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-2",
							},
						},
					},
				},
			},
			expected: map[string][]string{
				"repo-1": []string{"tag-1", "tag-2"},
			},
		},

		// Ignore non-ECR image
		{
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "other-registry.com/repo-1:tag-1",
							},
						},
					},
				},
				{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-2",
							},
						},
					},
				},
			},
			expected: map[string][]string{
				"repo-1": []string{"tag-2"},
			},
		},

		// Ignore untagged ECR images
		{
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						InitContainers: []v1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1",
							},
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-2",
							},
						},
					},
				},
			},
			expected: map[string][]string{
				"repo-1": []string{"tag-2"},
			},
		},

		// Ignore 'latest' tag
		{
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:latest",
							},
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-2",
							},
						},
					},
				},
			},
			expected: map[string][]string{
				"repo-1": []string{"tag-2"},
			},
		},
	}

	for _, testCase := range testCases {
		actual := ECRImagesFromPods(testCase.pods)

		if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("Expected result to be %+v, but was %+v", testCase.expected, actual)
		}
	}
}
