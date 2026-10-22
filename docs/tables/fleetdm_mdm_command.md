---
title: "Steampipe Table: fleetdm_mdm_command - Query FleetDM MDM Commands using SQL"
description: "Allows users to query the MDM commands sent to a FleetDM host."
---

# Table: fleetdm_mdm_command - Query FleetDM MDM Commands using SQL

Fleet sends MDM commands to Apple, Windows and Android hosts, for example to install profiles or lock a device. Uses the `/commands` API endpoint.

## Table Usage Guide

The `fleetdm_mdm_command` table requires `host_identifier` (hostname, UUID or hardware serial), as Fleet deprecated listing commands for every host. `request_type` and `command_status` filters are sent to the API. Fleet serves at most 100 pages of 1,000 commands per host this way.

## Examples

### List failed commands for a host

```sql+postgres
select
  command_uuid,
  request_type,
  status,
  updated_at
from
  fleetdm_mdm_command
where
  host_identifier = 'host-a'
  and command_status = 'failed';
```

```sql+sqlite
select
  command_uuid,
  request_type,
  status,
  updated_at
from
  fleetdm_mdm_command
where
  host_identifier = 'host-a'
  and command_status = 'failed';
```
