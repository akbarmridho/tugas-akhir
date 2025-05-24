# Setup Environments

## Setup Hetzner Cloud

Follow this guide:

[Terraform Kube Hetzner](https://github.com/kube-hetzner/terraform-hcloud-kube-hetzner).

Create two cluster with the following configuration (if possible):

- Backend cluster with 1 control plane node and 1 agent node.
- K6 Agent cluster with 1 control plane node and 2 agent node.

## Setting up DNS

- [ ] Setup DNS for the following entries in the laptop device: `payment.tugas-akhir.local`, `ticket.tugas-akhir.local`, `grafana.tugas-akhir.local`, `prometheus.tugas-akhir.local` `grafana.k6-agent.local` `prometheus.k6-agent.local`. Ensure that each entry corresponds to the Kubernetes Hetzner load balancer node address.
- [ ] Remember the IP address of backend cluster load balancer for `HOST_FORWARD` alias that will be used in k6 script in k6 agent cluster.

## Setup Docker Registry

Build and push the following images:

- [x] Custom K6 Build (`implementation/agent`) under `haiakbar/ta-agent`.
- [x] Payment Backend (`implementation/payment`) under `haiakbar/ta-payment`.
- [x] Ticket Backend (`implementation/backend`) under `haiakbar/ta-ticket`.
- [x] Citus Database (`implementation/infrastructure/local/citus.Dockerfile`) under `haiakbar/ta-citus`.
- [x] Postgres Database (`implementation/infrastructure/local/postgres.Dockerfile`) under `haiakbar/ta-postgres`.

When building the image, follow these steps:

**Note: ensure you have logged in from the Docker Desktop application**.

```bash
docker build -f Dockerfile -t haiakbar/<repo-name>:latest . &&
docker push haiakbar/<repo-name>:latest
```
