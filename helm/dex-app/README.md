# Helm Chart for Spark on Kubernetes cluster with Apache Livy and Spark History Server

#### Configurations

The configurable parameters for the Spark cluster components shold be found in the appropriate repos:
- [livy](https://github.com/arunmahadevan/arunmahadevan.github.io/tree/eks/helm/livy)
- [spark-history-server](https://github.com/arunmahadevan/arunmahadevan.github.io/tree/eks/helm/spark-history-server)

Review [values.yaml](values.yaml) file to see the defaults overrides.

#### Installing the Chart

To install or upgrade the chart execute:
```bash
helm repo add arunmahadevan https://arunmahadevan.github.io/charts
helm repo update
helm upgrade --install spark-test --namespace spark-test arunmahadevan/spark-test
```
