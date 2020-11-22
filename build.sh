#!/bin/bash

# Build app docker image
./gradlew bootBuildImage --imageName=docker.io/sbluemin/kakaopay-app:v1
./gradlew bootBuildImage --imageName=docker.io/sbluemin/kakaopay-app:v2