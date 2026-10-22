---
title: "Steampipe Table: fleetdm_scim_status - Query FleetDM SCIM Status using SQL"
description: "Allows users to query the status of FleetDM's SCIM integration."
---

# Table: fleetdm_scim_status - Query FleetDM SCIM Status using SQL

Fleet can receive users and groups from an identity provider over SCIM. Uses the `/scim/details` API endpoint (Fleet Premium).

## Table Usage Guide

The `fleetdm_scim_status` table returns one row with the last SCIM request Fleet received; the columns are null if none was ever received.

## Examples

### Check the last SCIM request

```sql+postgres
select
  last_request_status,
  last_request_details,
  last_request_at
from
  fleetdm_scim_status;
```

```sql+sqlite
select
  last_request_status,
  last_request_details,
  last_request_at
from
  fleetdm_scim_status;
```
