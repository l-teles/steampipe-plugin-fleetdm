---
title: "Steampipe Table: fleetdm_enroll_secret - Query FleetDM Enroll Secrets using SQL"
description: "Allows users to query FleetDM enroll secrets, global and per team."
---

# Table: fleetdm_enroll_secret - Query FleetDM Enroll Secrets using SQL

Enroll secrets let an agent enroll a host into Fleet, globally or into a team. Uses the `/spec/enroll_secret` and `/fleets/{id}/secrets` API endpoints.

## Table Usage Guide

Enroll secrets are live credentials. The `fleetdm_enroll_secret` table returns an error unless `expose_secrets = true` is set in the connection config, so they never land in query caches or exports by accident. Global secrets have a null `team_id`.

## Examples

### Find old enroll secrets

Secrets older than a year are candidates for rotation.

```sql+postgres
select
  team_id,
  created_at
from
  fleetdm_enroll_secret
where
  created_at < now() - interval '1 year';
```

```sql+sqlite
select
  team_id,
  created_at
from
  fleetdm_enroll_secret
where
  created_at < datetime('now', '-1 year');
```
