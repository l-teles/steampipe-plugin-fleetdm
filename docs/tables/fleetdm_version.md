---
title: "Steampipe Table: fleetdm_version - Query FleetDM Server Version using SQL"
description: "Allows users to query the FleetDM server version and build information, useful for inventory, upgrade planning, and compatibility checks."
---

# Table: fleetdm_version - Query FleetDM Server Version using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The version endpoint reports the Fleet server's version and build details.

## Table Usage Guide

The `fleetdm_version` table returns a single row with the Fleet server's version and build information. As a system administrator, you can use this table to record which server version each connection talks to, plan upgrades, and confirm compatibility before relying on newer API features.

## Examples

### Get the server version

Check which Fleet version the connection is talking to.

```sql+postgres
select
  version,
  branch,
  revision
from
  fleetdm_version;
```

```sql+sqlite
select
  version,
  branch,
  revision
from
  fleetdm_version;
```

### Get full build information

Inspect how and when the server binary was built.

```sql+postgres
select
  version,
  go_version,
  build_date,
  build_user
from
  fleetdm_version;
```

```sql+sqlite
select
  version,
  go_version,
  build_date,
  build_user
from
  fleetdm_version;
```

### Flag servers older than a minimum version

Compare the reported version against a required minimum (simple string comparison; adjust for your versioning scheme).

```sql+postgres
select
  version,
  version < '4.50.0' as needs_upgrade
from
  fleetdm_version;
```

```sql+sqlite
select
  version,
  version < '4.50.0' as needs_upgrade
from
  fleetdm_version;
```
