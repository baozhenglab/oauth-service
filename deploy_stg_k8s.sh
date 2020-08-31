#!/usr/bin/env bash

APP_NAME=oauth2
TAG=$1

echo "Downloading packages..."
go mod download
echo "Compiling..."
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app

echo "Docker building for k8s ..."
docker build -t ${APP_NAME} -f ./Dockerfile-Local .
echo "Docker tagging..."
docker image tag ${APP_NAME} asia.gcr.io/dofhunt-200lab/${APP_NAME}:${TAG}

echo "Pushing to registry DOF Hunt Hub ..."
docker push asia.gcr.io/dofhunt-200lab/${APP_NAME}:${TAG}

echo "Deploying with k8s ..."
kubectl set image deployment/oauth2 -n staging oauth2=asia.gcr.io/dofhunt-200lab/oauth2:${TAG} --record

echo "Cleaning..."
docker rmi $(docker images -qa -f 'dangling=true')
echo "Done"