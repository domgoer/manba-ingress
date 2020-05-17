# Manba Ingress (Still In Progress)

![Go](https://github.com/domgoer/manba-ingress/workflows/Go/badge.svg)

Use [Manba](https://github.com/fagongzi/manba) For Kubernetes Ingress.

Configure `api`, `routing`, `cluster`, `server` by using Custom Resource Definitions(CRDs)

## Features

- Fine-grained management

    Use crd to achieve configuration detailedly

- Automatic service registration

    Automatically register service to Manba when the it starts basing on Kubernetes

- Service consistency

    Health check consistent with Kubernetes
    

## Get Started

Setting up Manba for Kubernetes:

```shell script
# deploy custom resource definitions
kubectl apply -f deploy/crd.yml
# create namespace, serviceaccount, deployment
kubectl apply -f deploy/all-in-one.yml
```

## Documentation

All documentation is inside the [docs](./docs) directory.

## Build Locally

Start from code

```shell script
POD_NAME=test POD_NAMESPACE=default go run cmd/* --publish-status-address=:4443 --kubeconfig=~/.kube/config --manba-api-server-addr=localhost:32275 --update-status=false
```

Build docker image

```shell script
docker build -f Dockerfile -t domgoer/manba-ingress .
```

## Roadmap

- [ ] Doc 
- [ ] Test Case
- [ ] API TLS
