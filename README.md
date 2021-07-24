# kubectl-grep

Find the right peace in `kubectl get -o yaml` output. Pass substring of key to `kgrep` and it will find all matching keys.

## Examples

Show podAffinity

```sh
$ kubectl get pod my-pod -o yaml | kgrep podAff
.spec.affinity.podAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
  - labelSelector:
      matchLabels:
        app: my-app
        release: my-app
    topologyKey: kubernetes.io/hostname
```

Show container images for pod

```sh
$ kubectl get pods my-app -o yaml | kgrep --exact image # or kgrep -e image
.spec.containers[0].image: my-company/my-app:1.0.0
.spec.containers[1].image: hashicorp/vault-sidecar:1.4.5
```

Show kernel version on nodes

```sh
$ kubectl get node -o yaml | kgrep --show-status ker  # or kgrep -s ker
.items[0].status.nodeInfo.kernelVersion: 4.19.0-11-amd64
.items[1].status.nodeInfo.kernelVersion: 4.19.0-11-amd64
.items[2].status.nodeInfo.kernelVersion: 4.19.0-11-amd64
```
