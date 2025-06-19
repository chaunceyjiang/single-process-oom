# Single Process OOM In Pod

In Kubernetes, when using cgroup v2, if a single process within a Pod experiences an OOM (Out-Of-Memory) event, the entire Pod may be killed as a result.

This project allows a single process inside a Pod to be killed by the OOM killer **without causing the entire Pod to restart**, thereby improving the stability of the Pod.


## Requirements
- containerd >= 1.7.0


## Usage

``` shell
helm repo add single-process-oom-charts https://chaunceyjiang.github.io/single-process-oom
helm repo update
helm install single-process-oom single-process-oom-charts/single-process-oom -n kube-system
```