# K8S ECR Cleanup Controller

[![Build Status](https://travis-ci.org/danielfm/kube-ecr-cleanup-controller.svg?branch=master)](https://travis-ci.org/danielfm/kube-ecr-cleanup-controller)
[![codecov](https://codecov.io/gh/danielfm/kube-ecr-cleanup-controller/branch/master/graph/badge.svg)](https://codecov.io/gh/danielfm/kube-ecr-cleanup-controller)

AWS limits the number of images per ECR repository to 1000. This is not a
problem for low-activity projects, but if you have a full-fledged continuous
delivery pipeline in place that pushes images to a repository at every new
commit, sooner or later this limit will require you to periodically remove old
images in order to create room for new ones.

This controller handles this task of automatically keeping the number of images
in a ECR repository under some specified threshold.

## How it Works

First, the controller will query the Kubernetes API server to get the list of
currently running pods from the specified namespaces in order to see which ECR
images are currently in use.

Then, it will load the contents of the specified ECR repositories, sort those
images by push date, and remove from this list the images currently in use.
This step is very important as it ensures images in use _are not accidentally
deleted_. Also, this controller will not touch images tagged with the `latest`
tag.

Finally, it will remove the oldest images from this list.

### AWS Credentials

For the controller to work, it must have access to AWS credentials in
`~/.aws/credentials`, or via `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
environment variables.

The following IAM policy describes which actions the user must be able to
perform in order for the controller to work:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecr:BatchDeleteImage",
                "ecr:DescribeRepositories",
                "ecr:DescribeImages"
            ],
            "Resource": [
                "arn:aws:ecr:us-east-1:<id>:*"
            ]
        }
    ]
}
```

Make sure to set the `Resources` correctly for all ECR repos you intend to
clean up with this controller.

## Flags

```
$ ./kube-ecr-cleanup-controller -h
Usage of ./bin/kube-ecr-cleanup-controller:
  -alsologtostderr
    	log to standard error as well as files
  -dry-run
    	just log, don't delete any images.
  -interval int
    	check interval, in minutes. (default 30)
  -keep-filters string
        comma-separated list of filters or regexes that when matched will preserve the matching images.
  -kubeconfig string
    	path to a kubeconfig file.
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -logtostderr
    	log to standard error instead of files
  -max-images int
    	maximum number of images to keep in each repository. (default 900)
  -namespaces string
    	do not remove images used by pods in this comma-separated list of namespaces. (default "default")
  -region string
    	region to use when talking to AWS. (default "us-east-1")
  -registry-id string
    	specify a registry account ID. If not specified, uses the account ID of the credentials passed.
  -repos string
    	comma-separated list of repository names to watch.
  -stderrthreshold value
    	logs at or above this threshold go to stderr
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging
```

## Build Locally

Assuming you have your Go environment already configured:

1. Clone this repository in `$GOPATH/src/github.com/danielfm/kube-ecr-cleanup-controller`
2. Run `glide i --strip-vendor` and then `make` to build the Linux binary (or `make image` to build the Docker image)
3. Then, just re-tag the image to `<your-name>/kube-ecr-cleanup-controller:<tag>` and push to your own registry, if you feel like it

Or, you can use a pre-built image hosted in Docker Hub:
https://hub.docker.com/r/danielfm/kube-ecr-cleanup-controller/

Also, you can run `make test` to run the unit tests (or `make cover`, to run the tests with coverage reporting).

## Donate

If this project is useful for you, buy me a beer!

Bitcoin: `bc1qtwyfcj7pssk0krn5wyfaca47caar6nk9yyc4mu`

## License

Copyright (C) Daniel Fernandes Martins

Distributed under the New BSD License. See LICENSE for further details.
