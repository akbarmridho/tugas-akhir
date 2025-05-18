# Monitoring

```bash
chmod +x helmfile
helmfile apply
```

## Accessing Grafana

```bash
# Get the password
kubectl get secret --namespace monitoring grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo

# Get the pod name
export POD_NAME=$(kubectl get pods --namespace monitoring -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=grafana" -o jsonpath="{.items[0].metadata.name}")

# Port forward
kubectl --namespace monitoring port-forward $POD_NAME 3000
```

The username is `admin` and the password is `tugas-akhir`.

## Accessing Prometheus

```bash
export POD_NAME=$(kubectl get pods --namespace monitoring -l "app.kubernetes.io/name=prometheus,app.kubernetes.io/instance=prometheus" -o jsonpath="{.items[0].metadata.name}")
kubectl --namespace monitoring port-forward $POD_NAME 9090
```

## Todo for Remote

- In alloy config, change cluster name `k3d-tugas-akhir`. Run `kubectl config get-clusters` to get the cluster name.
