# Infrastructure Local

## K3d

Prerequisites:

- [Install Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
- [Install k3d](https://k3d.io/stable/#install-script)
- [Install Helm](https://helm.sh/docs/intro/install/)

Setup cluster:

```bash
k3d cluster create --config ./k3d.yaml
```

Delete cluster:

```bash
k3d cluster delete tugas-akhir
```

## Nginx

Apply Nginx

```bash
kubectl apply -f nginx.yaml
```
