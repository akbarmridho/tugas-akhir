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

### TOTAL REQUESTED

```bash
kubectl get pods --all-namespaces -o go-template='{{range .items}}{{range .spec.containers}}{{.resources.requests.cpu}} {{.resources.requests.memory}}{{"\n"}}{{end}}{{end}}' | \
awk '
function parse_cpu(cpu) {
    if (cpu ~ /m$/) {
        return substr(cpu, 1, length(cpu)-1)
    } else if (cpu ~ /^[0-9]+$/) {
        return cpu * 1000
    }
    return 0
}

function parse_memory(mem) {
    if (mem ~ /Ki$/) {
        return substr(mem, 1, length(mem)-2) / 1024
    } else if (mem ~ /Mi$/) {
        return substr(mem, 1, length(mem)-2)
    } else if (mem ~ /Gi$/) {
        return substr(mem, 1, length(mem)-2) * 1024
    } else if (mem ~ /Ti$/) {
        return substr(mem, 1, length(mem)-2) * 1024 * 1024
    } else if (mem ~ /Pi$/) {
        return substr(mem, 1, length(mem)-2) * 1024 * 1024 * 1024
    } else if (mem ~ /Ei$/) {
        return substr(mem, 1, length(mem)-2) * 1024 * 1024 * 1024 * 1024
    } else if (mem ~ /^[0-9]+$/) { # Assuming bytes if no unit
        return mem / 1024 / 1024
    }
    return 0
}

{
    total_cpu += parse_cpu($1)
    total_memory += parse_memory($2)
}

END {
    print "Total CPU Requested (millicores): " total_cpu
    print "Total Memory Requested (MiB): " total_memory
}
'
```

### Backup Stats

```bash
# Postgres Primary
kubectl exec pgcluster-0 -- psql -U postgres -d postgres -c "SELECT query, calls, total_exec_time, mean_exec_time FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 10;" > pg_output_primary.txt

# Postgres Replica
kubectl exec pgcluster-1 -- psql -U postgres -d postgres -c "SELECT query, calls, total_exec_time, mean_exec_time FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 10;" > pg_output_replica.txt

# Citus
kubectl exec cituscluster-0-0 -- psql -U postgres -d citus -c "SELECT query, calls, total_exec_time, mean_exec_time FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 10;" > citus_output.txt

# Yugabyte
kubectl exec pgcluster-0 -- psql "postgresql://yugabyte@yb-tserver-0.yb-tservers.default.svc.cluster.local:5433,yb-tserver-1.yb-tservers.default.svc.cluster.local:5433/yugabyte?sslmode=disable" -c "SELECT query, calls, total_time, mean_time FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;" > yugabyte_output.txt
```
