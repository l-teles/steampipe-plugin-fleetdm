---
title: "Steampipe Table: fleetdm_host_macadmins - Query FleetDM Host Macadmins using SQL"
description: "Allows users to query Munki and self-reported MDM enrollment for FleetDM hosts."
---

# Table: fleetdm_host_macadmins - Query FleetDM Host Macadmins using SQL

Fleet collects Munki details (macOS) and the MDM enrollment the host itself reports (macOS, Windows). Uses the `/hosts/{id}/macadmins` API endpoint.

## Table Usage Guide

The `fleetdm_host_macadmins` table has one row per host with such data; hosts without it (for example Linux) have no row. Without a `host_id` filter the table queries every host (up to 10 requests at a time), which can take a while on a large fleet; filter by `host_id` for a single host.

## Examples

### Find hosts that report no MDM enrollment

```sql+postgres
select
  m.host_id,
  h.hostname,
  m.mdm_enrollment_status
from
  fleetdm_host_macadmins m
  join fleetdm_host h on h.id = m.host_id
where
  m.mdm_enrollment_status = 'Off';
```

```sql+sqlite
select
  m.host_id,
  h.hostname,
  m.mdm_enrollment_status
from
  fleetdm_host_macadmins m
  join fleetdm_host h on h.id = m.host_id
where
  m.mdm_enrollment_status = 'Off';
```
