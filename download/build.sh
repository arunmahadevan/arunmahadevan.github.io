#!/bin/bash
docker build -t arunmahadevan/download:latest -f Dockerfile .
docker push arunmahadevan/download
