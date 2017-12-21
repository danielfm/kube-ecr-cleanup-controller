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

## Donate

If this project is useful for you, buy me a beer!

Bitcoin: `bc1qtwyfcj7pssk0krn5wyfaca47caar6nk9yyc4mu`

## License

Copyright (C) Daniel Fernandes Martins

Distributed under the New BSD License. See LICENSE for further details.
