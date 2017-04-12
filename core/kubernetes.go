package core

import (
	"regexp"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClientImpl struct {
	clientset *kubernetes.Clientset
}

type KubernetesClient interface {
	ListAllPods(namespace []*string) ([]*v1.Pod, error)
}

func NewKubernetesClient(kubeconfig string) (*KubernetesClientImpl, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubernetesClientImpl{
		clientset: clientset,
	}, nil
}

func (c *KubernetesClientImpl) ListAllPods(namespace []*string) ([]*v1.Pod, error) {
	opts := v1.ListOptions{}
	pods := []*v1.Pod{}

	for _, ns := range namespace {
		podList, err := c.clientset.Core().Pods(*ns).List(opts)
		if err != nil {
			return nil, err
		}

		for i := range podList.Items {
			pods = append(pods, &podList.Items[i])
		}
	}

	return pods, nil
}

func ECRImagesFromPods(region string, pods []*v1.Pod) map[string][]string {
	imagesPerRepo := map[string][]string{}
	encountered := map[string]bool{}

	re := regexp.MustCompile(`^.*\.dkr\.ecr\.[^\.]+\.amazonaws\.com/([^:/]+):(.*)$`)

	for _, pod := range pods {
		podContainers := append(pod.Spec.InitContainers, pod.Spec.Containers...)

		for _, container := range podContainers {
			if !encountered[container.Image] {
				imageData := re.FindStringSubmatch(container.Image)

				// Not an image hosted on ECR, or image is not tagged
				if imageData == nil {
					continue
				}

				repoName := imageData[1]
				imageTag := imageData[2]

				_, ok := imagesPerRepo[repoName]
				if ok {
					imagesPerRepo[repoName] = append(imagesPerRepo[repoName], imageTag)
				} else {
					imagesPerRepo[repoName] = []string{imageTag}
				}

				encountered[container.Image] = true
			}
		}
	}

	return imagesPerRepo
}
