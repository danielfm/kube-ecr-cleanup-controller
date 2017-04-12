package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"flag"

	"github.com/golang/glog"

	"github.com/danielfm/kube-ecr-cleanup-controller/core"
)

var task *core.CleanupTask

func init() {
	namespacesStr, reposStr := "default", ""

	task = core.NewCleanupTask()

	flag.StringVar(&task.KubeConfig, "kubeconfig", task.KubeConfig, "Path to a kubeconfig file.")
	flag.StringVar(&namespacesStr, "namespaces", namespacesStr, "Do not remove images used by pods in this comma-separated list of namespaces.")
	flag.IntVar(&task.Interval, "interval", task.Interval, "Check interval in minutes.")
	flag.IntVar(&task.MaxImages, "max-images", task.MaxImages, "Maximum number of images to keep in each repository.")
	flag.StringVar(&reposStr, "repos", reposStr, "Comma-separated list of repository names to watch.")
	flag.StringVar(&task.AwsRegion, "region", task.AwsRegion, "AWS Region to use when talking to AWS.")

	flag.Parse()

	if len(namespacesStr) == 0 {
		log.Fatalf("Must specify at least one namespace, exiting.")
	}
	if len(reposStr) == 0 {
		log.Fatalf("Must specify at least one ECR repository to watch, exiting.")
	}

	namespaces := core.ParseCommaSeparatedList(namespacesStr)
	repositories := core.ParseCommaSeparatedList(reposStr)

	if len(namespaces) == 0 {
		glog.Fatalf("Must specify at least one namespace, exiting.")
	}
	if len(repositories) == 0 {
		glog.Fatalf("Must specify at least one repository to watch, exiting.")
	}

	task.KubeNamespaces = namespaces
	task.EcrRepositories = repositories
}

func main() {
	glog.Infof("Kubernetes ECR Image Cleanup Controller started, will run every %d minute(s).", task.Interval)

	doneChan := make(chan struct{})
	var wg sync.WaitGroup

	for _, repo := range task.EcrRepositories {
		glog.Infof("Will clean up '%s' repo in '%s' region.", *repo, task.AwsRegion)
	}

	for _, namespace := range task.KubeNamespaces {
		glog.Infof("Images currently used by pods in '%s' namespace *will not* be removed.", *namespace)
	}

	wg.Add(1)
	task.ImageCleanupLoop(doneChan, &wg)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signalChan:
			glog.Info("Shutdown signal received, exiting...")
			close(doneChan)
			wg.Wait()
			os.Exit(0)
		}
	}
}
