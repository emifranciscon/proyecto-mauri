# Asientos — full-stack monorepo (Go + React + MySQL + Docker + Kubernetes)

Web app for **Asientos** (home), **Tanques** and **Balanzas** (master–detail with historial), with **JWT** auth and **bcrypt** passwords. MySQL runs **inside Kubernetes** as a **StatefulSet** with **persistent volumes**; the backend connects via **internal DNS** (for example `mysql.asientos.svc.cluster.local`).

Project layout:

- `frontend/` — React (Vite) SPA
- `backend/` — Go API (Gin, GORM, MySQL)
- `k8s/` — Kubernetes manifests
- `docs/` — data model notes

Authoritative product and architecture rules live in **`agent.md`**.

---

## Local development (without Docker)

### MySQL

Run MySQL 8 with database `asientos` and user/password matching `.env`. Mount `backend/migrations/docker-init` to `/docker-entrypoint-initdb.d` on first boot, or let GORM `AutoMigrate` create tables.

### Backend

```bash
cd backend
cp .env.example .env   # edit values
go run ./cmd/server
```

Health: `GET http://localhost:8080/healthz`

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Vite proxies `/api` and `/healthz` to `http://127.0.0.1:8080` by default (`VITE_DEV_PROXY_TARGET`).

---

## Local development (Docker Compose)

```bash
cp .env.example .env   # set JWT_SECRET and DB passwords
docker compose up --build
```

- Frontend (nginx + SPA + `/api` proxy): **http://localhost:3000**
- Backend (direct): **http://localhost:8080**
- MySQL: **localhost:3306** (dev only; do not expose MySQL in production)

MySQL data persists in the `mysql_data` volume. Init SQL is loaded from `backend/migrations/docker-init/` on first initialization.

Register a user from the login screen, then use Asientos / Tanques / Balanzas.

---

## Kubernetes deployment

### 1. Secrets and config

Edit and apply:

- `k8s/mysql-secret.yaml` — MySQL users/passwords (never commit real values).
- `k8s/backend-secret.yaml` — long random `JWT_SECRET`.
- `k8s/backend-configmap.yaml` — set `CORS_ORIGINS` to the browser origin users use (for example `https://asientos.example.com` or `http://<node-ip>:30080` if you only expose NodePort).

The backend uses:

- `DB_HOST=mysql.asientos.svc.cluster.local` (FQDN is explicit; short name `mysql` also works in the same namespace).

### 2. Storage

- **Managed cloud (EKS, GKE, AKS, …):** the default `mysql-statefulset.yaml` requests a PVC via the cluster default `StorageClass`. No extra PV manifest is required.
- **Lab without a provisioner:** apply `k8s/mysql-pv.yaml`, then add under `volumeClaimTemplates[].spec`:

  ```yaml
  storageClassName: manual
  ```

  so the PVC binds to that PV.

### 3. Images

Build and push (replace registry/path):

```bash
docker build -t <registry>/asientos-backend:latest ./backend
docker build -t <registry>/asientos-frontend:latest ./frontend
docker push <registry>/asientos-backend:latest
docker push <registry>/asientos-frontend:latest
```

Update `image:` in `k8s/backend-deployment.yaml` and `k8s/frontend-deployment.yaml`.

For **kind** / **minikube**:

```bash
docker build -t asientos-backend:latest ./backend
docker build -t asientos-frontend:latest ./frontend
kind load docker-image asientos-backend:latest
kind load docker-image asientos-frontend:latest
```

### 4. Apply manifests

Using kustomize:

```bash
kubectl apply -k k8s/
```

Or apply files in order: namespace → secrets → configmaps → MySQL service + StatefulSet → backend → frontend → optional Ingress.

### 5. Ingress (recommended for production)

`k8s/ingress.yaml` routes `/api` and `/healthz` to the backend and `/` to the frontend so the browser can call **same-origin** `/api` (build frontend with empty `VITE_API_URL`). Uncomment `ingress.yaml` in `k8s/kustomization.yaml` or apply it separately; set `host` and `ingressClassName` for your cluster.

The MySQL Service is **headless** (`clusterIP: None`); there is **no** public Service for the database.

---

## Troubleshooting (database connection issues)

| Symptom | Things to check |
|--------|------------------|
| Backend `CrashLoopBackOff` right after deploy | MySQL pod not ready yet — the app **retries** DB connect with backoff; if it still fails, verify `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` match `mysql-secret`. |
| `Access denied for user` | Wrong `MYSQL_USER` / `MYSQL_PASSWORD` vs backend env; re-create secret and restart backend. |
| `Unknown database` | `MYSQL_DATABASE` in secret must match `DB_NAME`; StatefulSet env must reference the same secret keys. |
| Works locally, fails in cluster | Use the in-cluster Service DNS: `mysql.<namespace>.svc.cluster.local` and port `3306`. Do not point to `localhost` from inside the backend pod. |
| PVC `Pending` | No default StorageClass or quota issues — `kubectl describe pvc -n asientos`; add a StorageClass or use manual PV + `storageClassName: manual`. |
| Empty schema | Init ConfigMap only runs on **first** MySQL data dir; wipe PVC only in non-prod or run migrations manually. Backend **AutoMigrate** still aligns tables if the DB is empty and credentials work. |

Useful commands:

```bash
kubectl logs -n asientos deploy/backend
kubectl logs -n asientos statefulset/mysql
kubectl get pvc -n asientos
kubectl exec -n asientos mysql-0 -- mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "SHOW DATABASES;"
```

---

## API overview (all JSON except `GET /healthz`)

| Method | Path | Auth |
|--------|------|------|
| POST | `/api/auth/register` | No |
| POST | `/api/auth/login` | No |
| GET | `/api/auth/me` | JWT |
| CRUD | `/api/asientos` | JWT |
| CRUD | `/api/tanques`, `/api/tanques/:id/historial` | JWT |
| CRUD | `/api/balanzas`, `/api/balanzas/:id/historial` | JWT |

Dates in JSON: `YYYY-MM-DD`.

---

## License

Proprietary / internal — adjust as needed.
