# Infrastructure Local

Assumption:

Run on Windows 11 with Ubuntu 22.04 on WSL2 and kernel version 5.15.167.4-microsoft-standard-WSL2.

## Prerequisites on WSL2

Recompile your kernel. Follow [this tutorial](https://kind.sigs.k8s.io/docs/user/using-wsl2/#kubernetes-service-with-session-affinity).

## K3d

Prerequisites:

- [Install Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
- [Install k3d](https://k3d.io/stable/#install-script)
- [Install Helm](https://helm.sh/docs/intro/install/)
- [Install Kubectx](https://github.com/ahmetb/kubectx)

Setup cluster:

```bash
k3d cluster create --config ./k3d.yaml
kubectl apply -f claim-standard-alias.yaml
```

Delete cluster:

```bash
k3d cluster delete tugas-akhir
```

## TLS

```bash
kubectl create secret tls service-tls \
  --cert=../../cert/cert.pem \
  --key=../../cert/key.pem

kubectl create namespace payment

kubectl create secret tls -n payment service-tls \
  --cert=../../cert/cert.pem \
  --key=../../cert/key.pem
```

## DNS

Create a script

```ps1
# Save this as: add-host-alias.ps1

# Path to the Windows hosts file
$hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"

$hostEntry = "127.0.0.1 registry.localhost registry2.localhost payment.tugas-akhir.local ticket.tugas-akhir.local"

# Check if the entry already exists
if ((Get-Content $hostsPath) -notmatch "registry.localhost registry2.localhost payment.tugas-akhir.local ticket.tugas-akhir.local") {
    Add-Content -Path $hostsPath -Value "`n$hostEntry"
    Write-Host "Host alias added: $hostEntry"
} else {
    Write-Host "Host alias already exists."
}
```

Run it

```ps1
Set-ExecutionPolicy Bypass -Scope Process -Force
.\add-host-alias.ps1
```

## Debugging

### Check Pod Status

```bash
kubectl describe pod <pod_name> -n default
```

### Get Logs

```bash
kubectl logs <pod_name> -c <container_name>

### Get Node Status

```bash
kubectl get nodes -o wide
kubectl describe node <node-name>
```

### Check PVC Status

```bash
kubectl get pvc -n default
```

### Get Summary of Resource Limits

```bash
kubectl get pods --all-namespaces -o custom-columns='NAME:.metadata.name,CPU_REQ:spec.containers[].resources.requests.cpu,CPU_LIM:spec.containers[].resources.limits.cpu,MEMORY_REQ:spec.containers[].resources.requests.memory,MEM_LIM:spec.containers[].resources.limits.memory'
```
