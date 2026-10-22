---
title: "Steampipe Table: fleetdm_host_configuration_profile - Query FleetDM Host Configuration Profiles using SQL"
description: "Allows users to query the delivery status of configuration profiles on each FleetDM host."
---

# Table: fleetdm_host_configuration_profile - Query FleetDM Host Configuration Profiles using SQL

Fleet delivers configuration profiles (OS settings) to hosts through MDM. This table lists, for each MDM-enrolled host, the profiles Fleet is installing or removing and their delivery status. Uses the `/hosts/{id}` API endpoint (without software); hosts not enrolled in MDM have no rows.

## Table Usage Guide

The `fleetdm_host_configuration_profile` table helps you find hosts where a profile failed or is still pending. Without a `host_id` filter the table queries every host (up to 10 requests at a time), which can take a while on a large fleet; filter by `host_id` for a single host.

## Examples

### List the profiles on one host

```sql+postgres
select
  name,
  platform,
  status,
  operation_type
from
  fleetdm_host_configuration_profile
where
  host_id = 1;
```

```sql+sqlite
select
  name,
  platform,
  status,
  operation_type
from
  fleetdm_host_configuration_profile
where
  host_id = 1;
```

### Find failed profiles across all hosts

```sql+postgres
select
  hostname,
  name,
  detail
from
  fleetdm_host_configuration_profile
where
  status = 'failed';
```

```sql+sqlite
select
  hostname,
  name,
  detail
from
  fleetdm_host_configuration_profile
where
  status = 'failed';
```
