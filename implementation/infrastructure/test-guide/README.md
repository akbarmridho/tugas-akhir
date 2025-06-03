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

**Cluster:**

```bash
terraform init --upgrade
terraform validate
terraform apply -auto-approve
```

**Helm:**

```bash
wget https://get.helm.sh/helm-v3.18.0-linux-amd64.tar.gz
tar -zxvf helm-v3.18.0-linux-amd64.tar.gz
mv linux-amd64/helm /usr/local/bin/helm

helm plugin install https://github.com/databus23/helm-diff

wget https://github.com/helmfile/helmfile/releases/download/v1.1.0/helmfile_1.1.0_linux_amd64.tar.gz
tar -zxvf helmfile_1.1.0_linux_amd64.tar.gz
mv helmfile /usr/local/bin/helmfile

kubectl config view --raw > ~/.kube/config
```

### Storageclass

```bash
# kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

kubectl get storageclass

kubectl patch storageclass hcloud-volumes -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'

kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

### Remove Control Plane Taint

```bash
kubectl get nodes

kubectl taint nodes <control-plane-node-name> node-role.kubernetes.io/control-plane-
```

### ENVSUBT

```bash
wget https://download.opensuse.org/repositories/openSUSE:/Factory/standard/x86_64/gettext-runtime-0.22.5-8.2.x86_64.rpm
transactional-update pkg install gettext-runtime-0.22.5-8.2.x86_64.rpm
```

### Other

- Update IP Adress in hostfile
- Set KUBECONFIG sesuai di notes.
