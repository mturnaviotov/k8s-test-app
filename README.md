# Basic Go Test App (Helm + Kubernetes)

A minimal Go web app packaged with a Helm chart for testing Kubernetes/Helm workflows. Useful as a starter for CI, local Kubernetes development (kind/minikube), and Helm chart experimentation.

## Repo layout

- backend/    — main Go application
- charts/     — Helm chart with sub-charts
- frontend.   - basic ui
- README.md

## Prerequisites

- Go (1.22+)
- Docker (or an alternative container builder)
- kubectl
- Helm (v3+)
- Local Kubernetes (kind, minikube, Docker Desktop, etc.)

## Quick start

1. Build the backend and image via commands described in Jenkinsfile
2. Build Helm package

```bash
helm dependency build ./charts
helm package ./charts
```

1. Install with Helm

```bash
helm install todoapp ./charts \
    --namespace todoapp --create-namespace \
    --set image.repository=<registry>/myapp,image.tag=0.1.0
```

1. Verify

```bash

kubectl get pods -n todoapp
kubectl logs -l app.kubernetes.io/name=todoapp -n mytodoappapp
kubectl port-forward svc/todoapp 8080:80 -n todoapp
# then open http://localhost:8080

```

## CURL API USAGE

```bash
# Add todo
curl -v -H "Content-Type: application/json" -X POST -d '{"text":"My test Todo item"}' http://localhost:8080/todos

# Get todos
curl http://127.0.0.1:8080/todos

# Update todo id
curl -v -H "Content-Type: application/json" -X PUT -d '{"text":"My test Todo item 11"}' http://localhost:8080/todos/1

# Delete todo id
curl -v -H "Content-Type: application/json" -X DELETE  http://localhost:8080/todos/1
```

## Helm tips

- Update values in `charts/todoapp/values.yaml` or pass overrides with `--set` or `-f`.
- Upgrade:

```bash
helm upgrade todoapp ./charts/ -n todoapp -f custom-values.yaml
```

- Uninstall:

```bash
helm uninstall todoapp -n todoapp
kubectl delete namespace todoapp --ignore-not-found
```

## Run locally (no container)

```bash
go run ./bin/myapp
# defaults: listens on PORT (env) or 8080
```

## Tests & lint

```bash
go test ./...
# optionally add go vet/gofmt/golint as needed
helm lint charts/todoapp
```

## Configuration

The app/chart expose simple configuration via:

- Container env: PORT, MESSAGE (or similar)
- Helm values: `image.repository`, `image.tag`, `service.type`, `replicaCount`, resource limits

Adjust `charts/todoapp/values.yaml` to match your environment.

## Contributing

Small changes and chart/value improvements welcome. Open PRs for fixes and enhancements.

## License

GNU GPL. you are free :)

<!-- End of README -->