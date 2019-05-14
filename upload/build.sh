#!/bin/bash
docker build -t arunmahadevan/upload-optimized:latest -f Dockerfile .
docker push arunmahadevan/upload-optimized
