FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.authors="Daniel Martins <daniel.martins@jusbrasil.com.br>"

COPY ./bin/kube-ecr-cleanup-controller /kube-ecr-cleanup-controller
ENTRYPOINT ["/kube-ecr-cleanup-controller"]

USER 65534:65534
