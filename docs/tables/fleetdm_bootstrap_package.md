---
title: "Steampipe Table: fleetdm_bootstrap_package - Query FleetDM Bootstrap Packages using SQL"
description: "Allows users to query the macOS bootstrap package of each FleetDM team."
---

# Table: fleetdm_bootstrap_package - Query FleetDM Bootstrap Packages using SQL

A bootstrap package is installed on macOS hosts during automatic enrollment. Uses the `/bootstrap/{team_id}/metadata` API endpoint (Fleet Premium, Apple MDM).

## Table Usage Guide

The `fleetdm_bootstrap_package` table has one row per team (`team_id = 0` for hosts with no team) that has a bootstrap package. The package's download token is never returned.

## Examples

### List bootstrap packages

```sql+postgres
select
  team_id,
  name,
  sha256,
  created_at
from
  fleetdm_bootstrap_package;
```

```sql+sqlite
select
  team_id,
  name,
  sha256,
  created_at
from
  fleetdm_bootstrap_package;
```
