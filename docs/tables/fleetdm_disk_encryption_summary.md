---
title: "Steampipe Table: fleetdm_disk_encryption_summary - Query FleetDM Disk Encryption Summary using SQL"
description: "Allows users to query FleetDM's disk encryption status counts."
---

# Table: fleetdm_disk_encryption_summary - Query FleetDM Disk Encryption Summary using SQL

Fleet enforces disk encryption (FileVault, BitLocker, LUKS) through MDM. Uses the `/disk_encryption` API endpoint (Fleet Premium).

## Table Usage Guide

The `fleetdm_disk_encryption_summary` table returns one row per status with host counts per platform. Filter with `team_id` (0 for hosts with no team).

## Examples

### Show hosts whose disk encryption needs attention

```sql+postgres
select
  status,
  macos,
  windows,
  linux
from
  fleetdm_disk_encryption_summary
where
  status in ('action_required', 'failed');
```

```sql+sqlite
select
  status,
  macos,
  windows,
  linux
from
  fleetdm_disk_encryption_summary
where
  status in ('action_required', 'failed');
```
