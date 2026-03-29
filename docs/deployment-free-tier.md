# Deploy to the internet (free tier): Vercel + Render + Aiven MySQL

This guide deploys the monorepo **without Kubernetes**, using only services with a **usable free tier**, optimized for **simplicity** and a **working** end-to-end URL.

---

## Platform choices (and why)

| Layer | Service | Reason |
|--------|---------|--------|
| **Frontend** | **Vercel** | Free static/SSR hosting, first-class Vite/React, automatic HTTPS, preview URLs, env vars for build time. |
| **Backend** | **Render** | Free **Web Service** with native support for Docker; Go runs from your existing `backend/Dockerfile`. Public HTTPS URL out of the box. |
| **Database** | **Aiven for MySQL** (free tier) | Real **MySQL** (MariaDB-compatible wire protocol) with a documented free tier—**no code change** from your current GORM MySQL driver. |

**Why not only Railway / PlanetScale / Neon?**

- **Railway** is mostly **usage-based credits**; fine for demos, not guaranteed “always $0”.
- **PlanetScale** has largely **removed** the old free hobby tier for new projects.
- **Neon** is **PostgreSQL**, not MySQL—you would need to swap the driver, DSN, and retest migrations.

**Trade-offs you accept on free tiers**

- **Render** free web apps **sleep after ~15 minutes** of no traffic; the first request after sleep can take **30–60+ seconds**.
- **Aiven** free MySQL has **resource limits** (fine for a small app); you must use **TLS** to connect (`DB_TLS=skip-verify` or `true`—see below).
- **Vercel** hobby is generous for static/frontend; API routes are not required here.

---

## Prerequisites

- GitHub repo with this project (root contains `frontend/`, `backend/`, etc.).
- GitHub account authorized on **Render** and **Vercel** when those dashboards ask for it.

---

## Step 1 — MySQL on Aiven (free)

1. Open **https://aiven.io** and sign up / log in.
2. Create a new **MySQL** service (choose the **free** plan if offered for your account/region).
3. Pick a **cloud region** close to your Render region (e.g. both in `eu` or both in `us`) to reduce latency.
4. Wait until the service status is **Running**.
5. In the Aiven console, open **Connection information** (or similar). Note:
   - **Host** → `DB_HOST`
   - **Port** → `DB_PORT` (usually `3306` or a non-default port—use exactly what Aiven shows)
   - **User** → `DB_USER`
   - **Password** → `DB_PASSWORD`
   - **Database name** → `DB_NAME` (often `defaultdb` unless you created another)
6. Aiven requires **TLS**. You will set **`DB_TLS=skip-verify`** on Render unless you configure proper CA verification (for a first deployment, `skip-verify` is common on free tiers; tighten later with the CA they provide).

**Connection string shape (for your notes only):**

```text
DB_USER:DB_PASSWORD@tcp(DB_HOST:DB_PORT)/DB_NAME?parseTime=true&tls=skip-verify
```

The app builds this from env vars; you only set the variables in the next step.

---

## Step 2 — Backend on Render (free Web Service)

1. Open **https://render.com** → **Dashboard** → **New +** → **Web Service**.
2. Connect **GitHub** and select your repository.
3. Configure the service:

   | Setting | Value |
   |--------|--------|
   | **Name** | e.g. `asientos-api` (becomes part of the URL) |
   | **Region** | Same broad area as Aiven if possible |
   | **Branch** | `main` (or your default branch) |
   | **Root Directory** | `backend` |
   | **Runtime** | **Docker** |
   | **Dockerfile Path** | `./Dockerfile` (default when root is `backend`) |
   | **Instance type** | **Free** |

4. **Environment variables** (Render → Environment):

   | Key | Value |
   |-----|--------|
   | `PORT` | **Do not set manually** — Render injects `PORT`. Your app already uses `PORT` from the environment (default `8080` locally). |
   | `DB_HOST` | From Aiven |
   | `DB_PORT` | From Aiven |
   | `DB_USER` | From Aiven |
   | `DB_PASSWORD` | From Aiven |
   | `DB_NAME` | From Aiven |
   | `DB_TLS` | `skip-verify` (recommended first deploy to Aiven) |
   | `JWT_SECRET` | Long random string (≥ 32 characters). Example generation: `openssl rand -hex 32` |
   | `CORS_ORIGINS` | Your **Vercel** URL(s), comma-separated, **no trailing slashes**. After Step 3, set e.g. `https://asientos.vercel.app` or `https://asientos-git-main-yourteam.vercel.app` for previews. You can add both: `https://asientos.vercel.app,https://asientos-xxx.vercel.app` |

5. **Health check path** (if Render asks): `/healthz`

6. Create the service and wait for the first **deploy** to finish.

7. Copy the **public URL** Render shows, e.g. `https://asientos-api.onrender.com`.  
   - Test: `curl -sS https://asientos-api.onrender.com/healthz` → should return JSON `{"status":"ok"}`.

8. **Cold starts:** If the service was idle, the first `curl` or browser load may take **up to ~1 minute** on the free tier.

9. **Update CORS after you know the frontend URL:** If the SPA is rejected by CORS, add the exact browser origin (scheme + host, no path) to `CORS_ORIGINS` and **redeploy** or restart the service.

---

## Step 3 — Frontend on Vercel (free)

1. Open **https://vercel.com** → **Add New…** → **Project** → import the **same GitHub repo**.
2. Configure:

   | Setting | Value |
   |--------|--------|
   | **Framework Preset** | **Vite** (auto-detected often) |
   | **Root Directory** | `frontend` |
   | **Build Command** | `npm run build` (default) |
   | **Output Directory** | `dist` (default for Vite) |
   | **Install Command** | `npm install` or `npm ci` |

3. **Environment Variables** (Vercel → Project → Settings → Environment Variables):

   | Name | Value | Environments |
   |------|--------|--------------|
   | `VITE_API_URL` | **Empty** if you use the rewrites below; **or** full backend origin `https://asientos-api.onrender.com` (no trailing slash) if you call the API directly from the browser |

   **Recommended (one public origin, no CORS for `/api`):** use Vercel rewrites so the browser talks to **same origin** and Vercel proxies to Render:

   Copy **`frontend/vercel.json.example`** to **`frontend/vercel.json`**, replace `YOUR-BACKEND` with your Render hostname (no `https://` inside the placeholder—keep the full URL as in the example), commit and push. Leave **`VITE_API_URL` empty** so the SPA uses relative `/api/...`.

   Equivalent inline snippet:

   ```json
   {
     "rewrites": [
       { "source": "/api/:path*", "destination": "https://YOUR-SERVICE.onrender.com/api/:path*" },
       { "source": "/healthz", "destination": "https://YOUR-SERVICE.onrender.com/healthz" }
     ]
   }
   ```

   With rewrites, you still need **CORS** only if something calls the Render URL **directly** from another origin; same-origin `/api` from the Vercel site does not need CORS. Optionally set `CORS_ORIGINS` on Render to your Vercel URL anyway for direct API testing.

4. Deploy. Copy the production URL, e.g. `https://asientos.vercel.app`.

5. **Order fix:** If you added rewrites, ensure Render URL is final before committing `vercel.json`, or update the file once Render is live.

---

## Step 4 — Smoke test

1. Open `https://<your-vercel-app>/` in a browser.
2. Use **Crear cuenta** / register (email + password ≥ 8 chars).
3. Log in and open **Asientos**; create a tanque/balanza if the lists are empty, then create an asiento.

If login fails with **network / CORS**:

- If **not** using Vercel rewrites: set `VITE_API_URL=https://<render-host>` on Vercel, **redeploy**, and set Render `CORS_ORIGINS` to `https://<vercel-host>` exactly.
- If using **rewrites**: confirm `vercel.json` destinations match the live Render URL and redeploy Vercel.

---

## Environment variables (summary)

### Render (backend)

| Variable | Required | Example / note |
|----------|----------|----------------|
| `DB_HOST` | Yes | Aiven host |
| `DB_PORT` | Yes | Aiven port |
| `DB_USER` | Yes | Aiven user |
| `DB_PASSWORD` | Yes | Aiven password |
| `DB_NAME` | Yes | Aiven database |
| `DB_TLS` | Yes for Aiven | `skip-verify` (or stricter later) |
| `JWT_SECRET` | Yes | Random secret |
| `CORS_ORIGINS` | Yes if browser hits Render directly | `https://your-app.vercel.app` |
| `PORT` | No | Set by Render |

### Vercel (frontend build)

| Variable | Required | Note |
|----------|----------|------|
| `VITE_API_URL` | No if using rewrites | Empty → relative `/api` |
| `VITE_API_URL` | Yes if no rewrites | `https://xxx.onrender.com` |

---

## Final public URLs (fill in after deploy)

| Component | Your URL |
|-----------|----------|
| **Frontend** | `https://<project>.vercel.app` |
| **Backend API** | `https://<service>.onrender.com` |
| **Database** | Not public (Aiven private host only) |

---

## Commands (local checks)

```bash
# Backend health (replace URL)
curl -sS https://YOUR-SERVICE.onrender.com/healthz

# Register (replace URL)
curl -sS -X POST https://YOUR-SERVICE.onrender.com/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"yourlongpass"}'
```

---

## Optional: `vercel.json` in the repo

If you use the rewrite approach, commit `frontend/vercel.json` with your real Render hostname so deploys stay reproducible. The backend change for `DB_TLS` is already supported via the `DB_TLS` environment variable.

---

## If Aiven free MySQL is unavailable in your region

Fallback order (still aiming for free / minimal cost):

1. **MariaDB SkySQL / other MySQL-compatible** free trial—use the same env vars; set `DB_TLS` as their docs require.
2. **Render PostgreSQL** (free) — requires switching the backend to `gorm.io/driver/postgres` and a new DSN (not covered in this doc; say if you want a Postgres variant).

---

## Security reminders (free tier ≠ “public database”)

- Never commit real `JWT_SECRET` or DB passwords; use dashboard env vars only.
- Rotate Aiven and Render secrets if they were ever pasted into chat or logs.
- Replace `DB_TLS=skip-verify` with CA-based verification when you harden the deployment.
