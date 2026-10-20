---
title: "Steampipe Table: fleetdm_apple_mdm - Query FleetDM Apple MDM using SQL"
description: "Allows users to query the APNs certificate FleetDM uses for Apple MDM."
---

# Table: fleetdm_apple_mdm - Query FleetDM Apple MDM using SQL

Fleet uses an Apple Push Notification service (APNs) certificate to manage Apple devices. When it expires, devices stop receiving MDM commands. Uses the `/mdm/apple` API endpoint.

## Table Usage Guide

The `fleetdm_apple_mdm` table returns one row with the certificate details, or no rows when Apple MDM is off.

## Examples

### Check how long until the APNs certificate expires

```sql+postgres
select
  common_name,
  renew_date,
  renew_date - now() as time_left
from
  fleetdm_apple_mdm;
```

```sql+sqlite
select
  common_name,
  renew_date,
  julianday(renew_date) - julianday('now') as days_left
from
  fleetdm_apple_mdm;
```
