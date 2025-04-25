# Infrastructure Local

Assumption:

Run on Windows 11 with Ubuntu 22.04 on WSL2.

## K3d

Prerequisites:

- [Install Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
- [Install k3d](https://k3d.io/stable/#install-script)
- [Install Helm](https://helm.sh/docs/intro/install/)
<!-- - [Install Helmfile (preferably via brew)](https://helmfile.readthedocs.io/en/latest/#installation) -->

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

## DNS

Create a script

```ps1
# Save this as: add-host-alias.ps1

# Path to the Windows hosts file
$hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"

$hostEntry = "127.0.0.1 registry.localhost"

# Check if the entry already exists
if ((Get-Content $hostsPath) -notmatch "registry.localhost") {
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
