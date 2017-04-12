package core

type CleanupTask struct {

	// Interval in which the clean-up process will happen, in minutes.
	Interval int

	// Number of images to keep in each ECR repository.
	MaxImages int

	// AWS region in which the repositories live.
	AwsRegion string

	// ECR repositories to clean up.
	EcrRepositories []*string

	// Path to the kubeconfig file used to access the Kubernetes cluster.
	// This is used to find out which images are in use, so they don't get
	// deleted by accident.
	KubeConfig string

	// Images used by pods running in these namespaces will not get deleted.
	KubeNamespaces []*string
}

func NewCleanupTask() *CleanupTask {
	return &CleanupTask{
		Interval:  30,
		MaxImages: 900,
		AwsRegion: "us-east-1",
	}
}
