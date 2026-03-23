This overlay adds:
- AlloyDB Auth Proxy sidecar
- Kubernetes ServiceAccount for Workload Identity

Use this overlay instead of the base when connecting to AlloyDB:

```bash
kubectl apply -k deploy/gke/alloydb-auth-proxy
```

Edit before applying:
- [configmap.yaml](/Users/gioboa/Desktop/Egen/go-gke-alloydb/deploy/gke/alloydb-auth-proxy/configmap.yaml): set `ALLOYDB_INSTANCE_URI`
- [serviceaccount.yaml](/Users/gioboa/Desktop/Egen/go-gke-alloydb/deploy/gke/alloydb-auth-proxy/serviceaccount.yaml): set your IAM service account email
- [secret.example.yaml](/Users/gioboa/Desktop/Egen/go-gke-alloydb/deploy/gke/secret.example.yaml): set `database-url` to use `127.0.0.1:5432`

Typical `DATABASE_URL`:

```text
postgres://DB_USER:DB_PASSWORD@127.0.0.1:5432/DB_NAME?sslmode=disable
```

One-time GKE / IAM setup:

```bash
export PROJECT_ID=your-project-id
export PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')
export CLUSTER_NAME=your-gke-cluster
export LOCATION=us-central1
export NAMESPACE=default
export KSA_NAME=go-gke-alloydb
export GSA_NAME=go-gke-alloydb

gcloud container clusters update "$CLUSTER_NAME" \
  --location="$LOCATION" \
  --workload-pool="$PROJECT_ID".svc.id.goog

gcloud iam service-accounts create "$GSA_NAME" \
  --project="$PROJECT_ID"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$GSA_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/alloydb.client"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$GSA_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/serviceusage.serviceUsageConsumer"

gcloud iam service-accounts add-iam-policy-binding \
  "$GSA_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/iam.workloadIdentityUser" \
  --member="serviceAccount:$PROJECT_ID.svc.id.goog[$NAMESPACE/$KSA_NAME]"
```

Notes:
- on GKE Autopilot, Workload Identity Federation is already enabled
- on GKE Standard, existing node pools may also need GKE metadata server enabled
- if you later add NetworkPolicy, allow egress to the GKE metadata server
- the proxy still needs network reachability to AlloyDB and egress to `443` and AlloyDB `5433`
- if you need public IP or PSC, adjust proxy args in [deployment-patch.yaml](/Users/gioboa/Desktop/Egen/go-gke-alloydb/deploy/gke/alloydb-auth-proxy/deployment-patch.yaml)
- review and bump the pinned proxy image tag periodically
