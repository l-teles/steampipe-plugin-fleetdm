---
title: "Steampipe Table: fleetdm_vpp_token - Query FleetDM VPP Tokens using SQL"
description: "Allows users to query Apple Volume Purchasing Program (VPP) tokens in FleetDM."
---

# Table: fleetdm_vpp_token - Query FleetDM VPP Tokens using SQL

VPP tokens let Fleet install App Store apps on Apple devices. Uses the `/vpp_tokens` API endpoint (Fleet Premium).

## Table Usage Guide

The `fleetdm_vpp_token` table helps you track token expiry and which teams use each token. Token values are never returned.

## Examples

### Find VPP tokens expiring in the next 30 days

```sql+postgres
select
  org_name,
  location,
  renew_date
from
  fleetdm_vpp_token
where
  renew_date < now() + interval '30 days';
```

```sql+sqlite
select
  org_name,
  location,
  renew_date
from
  fleetdm_vpp_token
where
  renew_date < datetime('now', '+30 days');
```
