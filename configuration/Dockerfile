FROM public.ecr.aws/amazonlinux/amazonlinux:2 as build

FROM scratch

# copy ca-bundle.crt from AmazonLinux2...
COPY --from=build /etc/ssl/certs/ca-bundle.crt /etc/ssl/certs/
COPY bin/amd64/linux/eks-connector /var/eks/connector

ENTRYPOINT ["/var/eks/connector"]