#!/bin/bash
cd ./src
docker build . -t susanhex/reactor:latest
docker push susanhex/reactor:latest