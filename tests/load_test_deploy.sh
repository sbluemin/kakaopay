#!/bin/bash

# v2 버전으로 롤링 업데이트를 실시 한다.
kubectl apply -f ../k8s/app-v2-deployment.yaml

# 각 1초가 걸리는 요청들을 2분간 던진다.
echo "Load testing..."
echo "GET http://localhost/pause/1" | ./vegeta_linux_amd64 attack -duration=2m | ./vegeta_linux_amd64 report -every 1s

# 테스트 후, v1으로 원상 복구
kubectl apply -f ../k8s/app-v1-deployment.yaml