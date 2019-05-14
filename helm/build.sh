#!/bin/bash
rm -f ./dex-base/charts/*
helm dependency update ./dex-base
rm -f ./dex-app/charts/*
helm dependency update ./dex-app
helm package dex-base dex-app
mv dex-base*.tgz dex-app*.tgz ../eks-charts/
cd ..
helm repo index ./eks-charts --url https://arunmahadevan.github.io/eks-charts
git add eks-charts
cat<<EOF
Please check the changes and:
 git commit -m "changes"
 git push
EOF
