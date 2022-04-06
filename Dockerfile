FROM centurylink/ca-certs
LABEL org.opencontainers.image.authors="Daniel Martins <daniel.martins@descomplica.com.br>"

COPY ./bin/kube-ecr-cleanup-controller /kube-ecr-cleanup-controller
ENTRYPOINT ["/kube-ecr-cleanup-controller"]

USER 65534:65534
