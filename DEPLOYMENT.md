# Perjadin Kemnaker RI - Production Deployment Workflow

This document outlines the complete, automated deployment workflow for the Perjadin backend and frontend to the production environment. The architecture strictly adheres to a Zero-Trust, Rootless Systemd Quadlet, and Doppler-driven infrastructure.

---

## 1. Triggering a Deployment
Deployments to production are **fully automated**. 
To trigger a deployment, simply push or merge your code into the `main` branch.

```bash
git add .
git commit -m "feat: added new dashboard metrics"
git push origin main
```

---

## 2. Continuous Integration (CI) - GitHub Actions
Once code hits the `main` branch, the `.github/workflows/deploy.yml` pipeline starts automatically:

1. **Authentication**: Actions logs into the GitHub Container Registry (`ghcr.io`).
2. **Build Backend**: Builds the Go backend using multi-stage builds (`backend/Dockerfile`) and pushes it as `ghcr.io/raxbyte-org/perjadin-kemnaker-ri-backend:latest`.
3. **Build Frontend**: Builds the SvelteKit frontend natively (`frontend/Dockerfile`) and pushes it as `ghcr.io/raxbyte-org/perjadin-kemnaker-ri-frontend:latest`.
4. **Caching**: Uses GitHub Actions cache (`type=gha`) to ensure blazingly fast subsequent builds.

---

## 3. Continuous Deployment (CD) - Production Server Sync
After the images are pushed to GHCR, the pipeline connects to the Production Server via SSH (`appleboy/ssh-action`) and executes the following sequence:

1. **Pull Infrastructure Changes**: Runs `git pull origin main` in the `~/perjadin-app` directory to pull any changes made to the Systemd Quadlet files (`deploy/quadlets/`).
2. **Sync Quadlets**: Copies the Quadlet configuration files (`.container`, `.network`, `.volume`, `.service`) directly into the Systemd directories (`~/.config/containers/systemd/`).
3. **Daemon Reload**: Executes `systemctl --user daemon-reload` so Systemd recognizes any updated infrastructure definitions.

---

## 4. Zero-Downtime Updates (Podman Auto-Update)
The final command executed by the CI/CD pipeline is:
```bash
podman auto-update
```
1. Podman scans all running containers for the label `io.containers.autoupdate=registry`.
2. It reaches out to GHCR and compares the cryptographic hash of the `latest` image against the one currently running.
3. If the image has changed (e.g., your new Go code or Svelte code), Podman pulls the new image.
4. Systemd performs a **graceful restart**: It shuts down the old container cleanly and starts the new one, minimizing downtime.

*Note: If you only changed Quadlet configs but did not change application code, `podman auto-update` will skip restarting. You must run `systemctl --user restart <service_name>` manually for infrastructure-only changes.*

---

## 5. Zero-Trust Secrets Management (Doppler)
Secrets (Database Passwords, JWT tokens, etc.) are **NEVER** stored on the server's hard drive.

Every time the server boots or a container is started, the `perjadin-secrets.service` kicks in automatically:
1. It connects to Doppler API using a scoped Service Token.
2. It fetches all production secrets and injects them directly into a **Volatile RAM Disk** at `/dev/shm/perjadin.env`.
3. The Quadlet containers read their environment variables directly from this RAM disk (`EnvironmentFile=/dev/shm/perjadin.env`).
4. If the server loses power, the RAM disk vanishes instantly, leaving zero traces of credentials.

---

## 6. Manual Operations Cheat Sheet

Although the pipeline is automated, here are some useful commands for inspecting the production environment via SSH:

*   **View Real-time Logs (Backend)**: `journalctl --user -u perjadin-backend.service -f`
*   **View Real-time Logs (Frontend)**: `journalctl --user -u perjadin-frontend.service -f`
*   **Check Container Status**: `podman ps`
*   **Force Pull & Restart**: `systemctl --user restart podman-auto-update.service`
*   **View RAM Disk Secrets (Requires Rootless User)**: `cat /dev/shm/perjadin.env`
