---
title: "Steampipe Table: fleetdm_android_enterprise - Query FleetDM Android Enterprise using SQL"
description: "Allows users to query the Android Enterprise FleetDM uses for Android MDM."
---

# Table: fleetdm_android_enterprise - Query FleetDM Android Enterprise using SQL

Fleet manages Android devices through an Android Enterprise bound to Google. Uses the `/android_enterprise` API endpoint.

## Table Usage Guide

The `fleetdm_android_enterprise` table returns one row with the enterprise ID, or no rows when Android MDM is not set up.

## Examples

### Check whether Android MDM is set up

```sql+postgres
select
  android_enterprise_id
from
  fleetdm_android_enterprise;
```

```sql+sqlite
select
  android_enterprise_id
from
  fleetdm_android_enterprise;
```
