# K8S ECR Cleanup Controller

AWS limits the number of images per ECR repository to 1000. This is not a
problem for low-activity projects, but if you have a full-fledged continuous
delivery pipeline in place that pushes images to a repository at every new
commit, sooner or later this limit will require you to periodically remove old
images in order to create room for new ones.

This controller handles this task of automatically keeping the number of images
in a ECR repository under some specified threshold.

**Notice:** This project is still WIP, do not use in production.

## How it Works

First, the controller will query the Kubernetes API server to get the list of
currently running pods from the specified namespaces in order to see which ECR
images are currently in use.

Then, it will load the contents of the specified ECR repositories, sort those
images by push date, and remove from this list the images currently in use.
This step is very important as it ensures images in use _are not accidentally
deleted_.

Finally, it will remove the oldest images from this list.

## Flags

```
$ ./kube-ecr-cleanup-controller -h
Usage of ./kube-ecr-cleanup-controller:
  -alsologtostderr
    	log to standard error as well as files
  -interval int
    	Check interval in minutes. (default 30)
  -kubeconfig string
    	Path to a kubeconfig file.
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -logtostderr
    	log to standard error instead of files
  -max-images int
    	Maximum number of images to keep in each repository. (default 900)
  -namespaces string
    	Do not remove images used by pods in this comma-separated list of namespaces. (default "default")
  -region string
    	AWS Region to use when talking to AWS. (default "us-east-1")
  -repos string
    	Comma-separated list of repository names to watch.
  -stderrthreshold value
    	logs at or above this threshold go to stderr
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging
```

## License

Copyright (C) Daniel Fernandes Martins

Distributed under the New BSD License. See LICENSE for further details.
