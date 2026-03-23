Base deploy, no AlloyDB proxy:

```bash
export PROJECT_ID=your-project-id
export REGION=us-central1
export REPOSITORY=containers
export IMAGE=$REGION-docker.pkg.dev/$PROJECT_ID/$REPOSITORY/go-gke-alloydb

gcloud auth configure-docker $REGION-docker.pkg.dev
docker build -t $IMAGE:latest .
docker push $IMAGE:latest
```

Set the image in [kustomization.yaml](/Users/gioboa/Desktop/Egen/go-gke-alloydb/deploy/gke/kustomization.yaml), then:

```bash
kubectl apply -k deploy/gke
```

Optional DB secret:

```bash
kubectl apply -f deploy/gke/secret.example.yaml
```

Get external IP:

```bash
kubectl get svc go-gke-alloydb
```

Notes:
- service is `LoadBalancer`, so it creates a public GKE load balancer
- if `DATABASE_URL` is unset, only `/ping` and health endpoints work
- if using AlloyDB auth proxy/private IP, point `database-url` at that reachable address

AlloyDB + Workload Identity overlay:

```bash
kubectl apply -k deploy/gke/alloydb-auth-proxy
```

See [alloydb-auth-proxy/README.md](/Users/gioboa/Desktop/Egen/go-gke-alloydb/deploy/gke/alloydb-auth-proxy/README.md).
