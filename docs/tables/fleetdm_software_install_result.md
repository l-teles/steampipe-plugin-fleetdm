---
title: "Steampipe Table: fleetdm_software_install_result - Query FleetDM Software Install Results using SQL"
description: "Allows users to query the result of a FleetDM software install."
---

# Table: fleetdm_software_install_result - Query FleetDM Software Install Results using SQL

Each software install on a host has an install UUID and a result. Uses the `/software/install/{install_uuid}/results` API endpoint.

## Table Usage Guide

The `fleetdm_software_install_result` table requires `install_uuid` (the `install_uuid` column of `fleetdm_host_activity`). The install, pre-install and post-install outputs are NULL unless `expose_secrets = true` is set in the connection config.

## Examples

### Get a software install result

```sql+postgres
select
  host_id,
  software_title,
  status,
  updated_at
from
  fleetdm_software_install_result
where
  install_uuid = '00000000-0000-0000-0000-000000000000';
```

```sql+sqlite
select
  host_id,
  software_title,
  status,
  updated_at
from
  fleetdm_software_install_result
where
  install_uuid = '00000000-0000-0000-0000-000000000000';
```
