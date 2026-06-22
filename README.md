cd ..   # back to openchoreo-practice/ root
cat > README.md << 'EOF'
# OpenChoreo Practice Project — Progressive Delivery Prep

A 4-day hands-on practice project built ahead of a WSO2 internship on
**Progressive Delivery for OpenChoreo** (Project 682). The goal: get fluent
with OpenChoreo's developer/platform abstractions and Cell-based visibility
model, and personally reproduce the exact problem the internship project
exists to solve — that OpenChoreo currently has no way to roll out a new
version gradually, only a standard Kubernetes rolling update.

## What this is

A tiny URL shortener, deliberately split into three components to exercise
different parts of OpenChoreo:

| Service | Visibility | Role |
|---|---|---|
| `gateway` | Public | Exposes `POST /shorten` and `GET /:code`, forwards to `core-api` |
| `core-api` | Project-internal | Owns short-code generation/resolution; **the progressive-delivery rollout target** |
| `worker` | Private (scheduled) | Periodically calls `core-api`'s `/internal/cleanup` to purge expired links |

`core-api` ships a `BUG_MODE` flag used to simulate a bad release (flaky
lookups / added latency), so the rolling-update gap and a manual canary fix
can both be demonstrated with real metrics behind them.

## Status

- [x] Day 1 — environment setup, OpenChoreo architecture deep-read, `core-api` skeleton
- [ ] Day 2 — `core-api` deployed across environments, `gateway` built and deployed
- [ ] Day 3 — `worker` built and deployed, `BUG_MODE` v2 gap captured, observability pass
- [ ] Day 4 — manual Argo Rollouts canary on `core-api`, deliverable packaged

## Running locally

```bash
cd core-api && go run main.go   # :8080
```

Other services follow the same pattern once they exist.

## Repo layout

gateway/      Public entrypoint

core-api/     Rollout target — the interesting one

worker/       Scheduled cleanup

openchoreo/   Component/pipeline manifests

rollout-experiment/   Manual Argo Rollouts + Gateway API setup (Day 4)


## Why this exists

Practice repo for understanding OpenChoreo deeply before contributing
progressive delivery support (canary/blue-green, gateway-agnostic traffic
shifting, automated analysis and rollback) to the platform itself.
EOF