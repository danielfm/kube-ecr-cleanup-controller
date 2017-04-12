FROM centurylink/ca-certs
MAINTAINER Daniel Martins <daniel.martins@descomplica.com.br>

COPY ./bin/kube-ecr-cleanup-controller /kube-ecr-cleanup-controller
ENTRYPOINT ["/kube-ecr-cleanup-controller"]
