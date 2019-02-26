package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"flag"

	"github.com/golang/glog"

	"github.com/danielfm/kube-ecr-cleanup-controller/pkg/core"
	"github.com/danielfm/kube-ecr-cleanup-controller/pkg/processor"
	"github.com/danielfm/kube-ecr-cleanup-controller/pkg/utils"
)

var task *core.CleanupTask

// VERSION set by build script
var VERSION = "UNKNOWN"

func init() {
	namespacesStr, reposStr, registryID := "default", "", ""

	task = core.NewCleanupTask()

	flag.StringVar(&task.KubeConfig, "kubeconfig", task.KubeConfig, "path to a kubeconfig file.")
	flag.StringVar(&namespacesStr, "namespaces", namespacesStr, "do not remove images used by pods in this comma-separated list of namespaces.")
	flag.IntVar(&task.Interval, "interval", task.Interval, "check interval, in minutes.")
	flag.IntVar(&task.MaxImages, "max-images", task.MaxImages, "maximum number of images to keep in each repository.")
	flag.StringVar(&reposStr, "repos", reposStr, "comma-separated list of repository names to watch.")
	flag.StringVar(&task.AwsRegion, "region", task.AwsRegion, "region to use when talking to AWS.")
	flag.BoolVar(&task.DryRun, "dry-run", task.DryRun, "just log, don't delete any images.")
	flag.StringVar(&registryID, "registry-id", registryID, "specify a registry account ID. If not specified, uses the account ID of the credentials passed.")

	flag.Parse()

	if len(namespacesStr) == 0 {
		log.Fatalf("Must specify at least one namespace, exiting.")
	}
	if len(reposStr) == 0 {
		log.Fatalf("Must specify at least one ECR repository to watch, exiting.")
	}

	namespaces := utils.ParseCommaSeparatedList(namespacesStr)
	repositories := utils.ParseCommaSeparatedList(reposStr)

	if len(namespaces) == 0 {
		glog.Fatalf("Must specify at least one namespace, exiting.")
	}
	if len(repositories) == 0 {
		glog.Fatalf("Must specify at least one repository to watch, exiting.")
	}

	if len(registryID) == 0 {
		task.RegistryID = nil
	} else {
		task.RegistryID = &registryID
	}

	task.KubeNamespaces = namespaces
	task.EcrRepositories = repositories
}

func main() {
	glog.Infof("Kubernetes ECR Image Cleanup Controller v%s started, will run every %d minute(s).", VERSION, task.Interval)

	doneChan := make(chan struct{})
	var wg sync.WaitGroup

	for _, repo := range task.EcrRepositories {
		glog.Infof("Will clean up '%s' repo in '%s' region.", *repo, task.AwsRegion)
	}

	for _, namespace := range task.KubeNamespaces {
		glog.Infof("Images currently used by pods in '%s' namespace *will not* be removed.", *namespace)
	}

	wg.Add(1)
	processor.ImageCleanupLoop(task, doneChan, &wg)

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
