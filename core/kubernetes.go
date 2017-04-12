package core

import (
	"regexp"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesClient defines the expected interface of any object capable of
// listing pods from a Kubernetes cluster.
type KubernetesClient interface {
	ListAllPods(namespace []*string) ([]*v1.Pod, error)
}

type KubernetesClientImpl struct {
	clientset *kubernetes.Clientset
}

// NewKubernetesClient returns a client capable of talking to the API server
// of a Kubernetes cluster specified in the given kubeconfig filepath. If no
// kubeconfig filepath is specified, it assumes it's running inside a Kubernetes
// cluster, and will try to connect to it via the exposed service account.
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

// ListAllPods returns all pods from the given namespaces.
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

// ECRImagesFromPods converts the given list of pods to a map where the keys
// are the ECR repository names and their values are a slice of strings
// containing the unique image tags referenced by those pods.
func ECRImagesFromPods(pods []*v1.Pod) map[string][]string {
	imagesPerRepo := map[string][]string{}
	encountered := map[string]bool{}

	// Only matches tagged images hosted on ECR
	re := regexp.MustCompile(`^.*\.dkr\.ecr\.[^\.]+\.amazonaws\.com/([^:/]+):(.*)$`)

	for _, pod := range pods {
		podContainers := append(pod.Spec.InitContainers, pod.Spec.Containers...)

		for _, container := range podContainers {

			// Ignore images we already seen
			if !encountered[container.Image] {
				imageData := re.FindStringSubmatch(container.Image)
				if imageData == nil {
					continue
				}

				encountered[container.Image] = true

				repoName, imageTag := imageData[1], imageData[2]
				_, ok := imagesPerRepo[repoName]
				if ok {
					imagesPerRepo[repoName] = append(imagesPerRepo[repoName], imageTag)
				} else {
					imagesPerRepo[repoName] = []string{imageTag}
				}
			}
		}
	}

	return imagesPerRepo
}
