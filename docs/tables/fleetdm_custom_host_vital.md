---
title: "Steampipe Table: fleetdm_custom_host_vital - Query FleetDM Custom Host Vitals using SQL"
description: "Allows users to query FleetDM custom host vitals."
---

# Table: fleetdm_custom_host_vital - Query FleetDM Custom Host Vitals using SQL

Custom host vitals (Fleet 4.90+) are extra host attributes, filled in by a query, that labels can target. Uses the `/custom_host_vitals` API endpoint.

## Table Usage Guide

The `fleetdm_custom_host_vital` table lists the vitals defined in Fleet. On older Fleet versions it returns no rows.

## Examples

### List custom host vitals

```sql+postgres
select
  id,
  name,
  created_at
from
  fleetdm_custom_host_vital;
```

```sql+sqlite
select
  id,
  name,
  created_at
from
  fleetdm_custom_host_vital;
```
