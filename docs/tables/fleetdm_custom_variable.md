---
title: "Steampipe Table: fleetdm_custom_variable - Query FleetDM Custom Variables using SQL"
description: "Allows users to query FleetDM custom (secret) variable names."
---

# Table: fleetdm_custom_variable - Query FleetDM Custom Variables using SQL

Custom variables are secrets that scripts and configuration profiles reference as `$FLEET_SECRET_<NAME>`. Uses the `/custom_variables` API endpoint; values are never returned.

## Table Usage Guide

The `fleetdm_custom_variable` table lists which variables exist and when they changed.

## Examples

### List custom variables by last update

```sql+postgres
select
  name,
  updated_at
from
  fleetdm_custom_variable
order by
  updated_at desc;
```

```sql+sqlite
select
  name,
  updated_at
from
  fleetdm_custom_variable
order by
  updated_at desc;
```
