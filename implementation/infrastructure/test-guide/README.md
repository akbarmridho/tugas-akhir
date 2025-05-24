# Test Guide

Steps:

- [Setting up environments](./setup-environments.md).
- [Setting up backend cluster](./setup-backend-cluster.md).
- [Setting up k6 cluster](./setup-k6-cluster.md).
- [Testing](./testing.md).
- [Backup data](./backup.md).

## Quick Notes

SSH to backend cluster.

```bash
ssh root@138.199.153.132 -i ~/.ssh/id_ed25519 -o StrictHostKeyChecking=no
```

### Setup Nodes

- Download binary for helm, helmfile.
- `kubectl config view --raw > ~/.kube/config`.
- Helm plugin diff `helm plugin install https://github.com/databus23/helm-diff`.

### Storageclass

```bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

kubectl get storageclass

kubectl patch storageclass hcloud-volumes -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'

kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

### Remove Control Plane Taint

```bash
kubectl get nodes

kubectl taint nodes <control-plane-node-name> node-role.kubernetes.io/control-plane-
```
