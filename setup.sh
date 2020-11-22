#!/bin/bash

# Clean k8s
sudo rm -rf /kakaopay-db-data
sudo mkdir -m 770 /kakaopay-db-data

sudo rm -rf /logs
sudo mkdir -m 700 /logs | sudo chown 1000:1000 /logs

kubectl delete -f ./k8s/mysql-deployment.yaml
kubectl delete -f ./k8s/app-ingress.yaml
kubectl delete -f ./k8s/app-service.yaml
kubectl delete -f ./k8s/app-v1-deployment.yaml

# Setup k8s
kubectl apply -f ./k8s/mysql-deployment.yaml
kubectl apply -f ./k8s/app-ingress.yaml
kubectl apply -f ./k8s/app-service.yaml
kubectl apply -f ./k8s/app-v1-deployment.yaml