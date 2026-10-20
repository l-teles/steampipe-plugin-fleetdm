---
title: "Steampipe Table: fleetdm_ab_token - Query FleetDM Apple Business Tokens using SQL"
description: "Allows users to query Apple Business (Manager) tokens in FleetDM."
---

# Table: fleetdm_ab_token - Query FleetDM Apple Business Tokens using SQL

Apple Business tokens connect Fleet to Apple Business (formerly Apple Business Manager) for automatic device enrollment. Uses the `/ab_tokens` API endpoint (`/abm_tokens` on older Fleet versions; Fleet Premium).

## Table Usage Guide

The `fleetdm_ab_token` table helps you track token expiry and the default team per platform. Token values are never returned.

## Examples

### Find tokens expiring in the next 30 days

```sql+postgres
select
  org_name,
  renew_date,
  terms_expired,
  token_invalid
from
  fleetdm_ab_token
where
  renew_date < now() + interval '30 days';
```

```sql+sqlite
select
  org_name,
  renew_date,
  terms_expired,
  token_invalid
from
  fleetdm_ab_token
where
  renew_date < datetime('now', '+30 days');
```
