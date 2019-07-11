#!/bin/bash -e

rm -f find-failing-notifications.linux64
echo Compiling...
GOOS=linux GOARCH=amd64 go build -o find-failing-notifications.linux64

if [ "$VERSION" == "" ]; then
  VERSION=DEV
fi

echo Building docker image...
docker build . -t andyg42/find-failing-notifications:$VERSION

echo Uploading...
docker push andyg42/find-failing-notifications:$VERSION
