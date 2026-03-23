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

Optional DB secret. Keep real values out of git:

```bash
kubectl create secret generic go-gke-alloydb \
  --from-literal=db-user=YOUR_DB_USER \
  --from-literal=db-password='YOUR_DB_PASSWORD' \
  --from-literal=db-name=YOUR_DB_NAME
```

Get external IP:

```bash
kubectl get svc go-gke-alloydb
```

Notes:
- service is `LoadBalancer`, so it creates a public GKE load balancer
- if DB env vars are unset, only `/ping` and health endpoints work
- app supports either `DATABASE_URL` or split vars `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME`
- if using AlloyDB auth proxy, app talks to `127.0.0.1:5432`
- [secret.example.yaml](/Users/gioboa/Desktop/Egen/go-gke-alloydb/deploy/gke/secret.example.yaml) is placeholders only

AlloyDB + Workload Identity overlay:

```bash
kubectl apply -k deploy/gke/alloydb-auth-proxy
```

See [alloydb-auth-proxy/README.md](/Users/gioboa/Desktop/Egen/go-gke-alloydb/deploy/gke/alloydb-auth-proxy/README.md).
