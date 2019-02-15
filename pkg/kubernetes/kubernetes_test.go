package kubernetes

import (
	"reflect"
	"testing"

	apiv1 "k8s.io/api/core/v1"
)

func TestECRImagesFromPods(t *testing.T) {
	testCases := []struct {
		pods     []*apiv1.Pod
		expected map[string][]string
	}{
		// Different tagged images from different repos in the same pod
		{
			pods: []*apiv1.Pod{
				{
					Spec: apiv1.PodSpec{
						InitContainers: []apiv1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-1",
							},
						},
						Containers: []apiv1.Container{
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
			pods: []*apiv1.Pod{
				{
					Spec: apiv1.PodSpec{
						InitContainers: []apiv1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-1",
							},
						},
						Containers: []apiv1.Container{
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
			pods: []*apiv1.Pod{
				{
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Image: "id.dkr.ecr.region.amazonaws.com/repo-1:tag-1",
							},
						},
					},
				},
				{
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
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
			pods: []*apiv1.Pod{
				{
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Image: "other-registry.com/repo-1:tag-1",
							},
						},
					},
				},
				{
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
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
			pods: []*apiv1.Pod{
				{
					Spec: apiv1.PodSpec{
						InitContainers: []apiv1.Container{
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
			pods: []*apiv1.Pod{
				{
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
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
