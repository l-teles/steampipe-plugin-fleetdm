---
title: "Steampipe Table: fleetdm_host_upcoming_activity - Query FleetDM Host Upcoming Activities using SQL"
description: "Allows users to query activities queued for each FleetDM host."
---

# Table: fleetdm_host_upcoming_activity - Query FleetDM Host Upcoming Activities using SQL

Scripts, software installs and MDM actions wait in a per-host queue until the host picks them up. Uses the `/hosts/{id}/activities/upcoming` API endpoint.

## Table Usage Guide

The `fleetdm_host_upcoming_activity` table lists queued activities per host. Without a `host_id` filter the table queries every host (up to 10 requests at a time); filter by `host_id` for a single host.

## Examples

### Find hosts with queued activities

```sql+postgres
select
  host_id,
  count(*) as queued
from
  fleetdm_host_upcoming_activity
group by
  host_id
order by
  queued desc;
```

```sql+sqlite
select
  host_id,
  count(*) as queued
from
  fleetdm_host_upcoming_activity
group by
  host_id
order by
  queued desc;
```
