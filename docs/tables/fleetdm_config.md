---
title: "Steampipe Table: fleetdm_config - Query FleetDM Configuration using SQL"
description: "Allows users to query FleetDM's global configuration, one JSON column per settings section."
---

# Table: fleetdm_config - Query FleetDM Configuration using SQL

FleetDM's global configuration covers organization info, server settings, MDM, features, vulnerability processing, integrations and more. Uses the `/config` API endpoint and returns a single row.

## Table Usage Guide

The `fleetdm_config` table lets you audit global settings with SQL. Fleet masks the SMTP password itself. `integrations`, `webhook_settings` and `logging` are NULL unless `expose_secrets = true` is set in the connection config, because their URLs can embed credentials.

## Examples

### Check the license tier and expiry

```sql+postgres
select
  license ->> 'tier' as tier,
  license ->> 'expiration' as expires,
  license ->> 'device_count' as device_count
from
  fleetdm_config;
```

```sql+sqlite
select
  json_extract(license, '$.tier') as tier,
  json_extract(license, '$.expiration') as expires,
  json_extract(license, '$.device_count') as device_count
from
  fleetdm_config;
```

### Review global macOS update enforcement

```sql+postgres
select
  mdm -> 'macos_updates' ->> 'minimum_version' as minimum_version,
  mdm -> 'macos_updates' ->> 'deadline' as deadline
from
  fleetdm_config;
```

```sql+sqlite
select
  json_extract(mdm, '$.macos_updates.minimum_version') as minimum_version,
  json_extract(mdm, '$.macos_updates.deadline') as deadline
from
  fleetdm_config;
```
