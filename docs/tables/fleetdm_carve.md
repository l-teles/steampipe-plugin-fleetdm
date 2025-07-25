---
title: "Steampipe Table: fleetdm_carve - Query FleetDM File Carves using SQL"
description: "Allows users to query FleetDM file carving sessions, providing details on carved files, host origins, session status, and any associated errors."
---

# Table: fleetdm_carve - Query FleetDM File Carves using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. File carving in FleetDM is the process of extracting a file from a host, often for forensic analysis, initiated by an osquery query.

## Table Usage Guide

The `fleetdm_carve` table provides insights into file carving sessions within your FleetDM instance. As a security analyst or incident responder, you can use this table to track forensic activities, review the status of file extractions, and identify any errors that occurred during the process. This table is essential for monitoring and auditing forensic data collection across your fleet.

## Examples

### List the 50 most recent file carving sessions
Get an overview of the latest file carving activities across all hosts.

```sql+postgres
select
  id,
  name,
  host_id,
  carve_size,
  created_at
from
  fleetdm_carve
order by
  id desc
limit 50;
```sql+sqlite
select
  id,
  name,
  host_id,
  carve_size,
  created_at
from
  fleetdm_carve
order by
  id desc
limit 50;
```

### Find all carving sessions that resulted in an error
Identify carving sessions that failed, allowing you to investigate potential issues with osquery, storage configuration, or host connectivity.

```sql+postgres
select
  id,
  name,
  host_id,
  error,
  created_at
from
  fleetdm_carve
where
  error is not null
order by
  created_at desc;
```sql+sqlite
select
  id,
  name,
  host_id,
  error,
  created_at
from
  fleetdm_carve
where
  error is not null
order by
  created_at desc;
```

### List all carves for a specific host
Review all historical file carving activities for a particular host to aid in an investigation.

```sql+postgres
select
  id,
  name,
  carve_size,
  expired,
  created_at
from
  fleetdm_carve
where
  host_id = 7 -- Replace with an actual host ID
order by
  id desc;
```sql+sqlite
select
  id,
  name,
  carve_size,
  expired,
  created_at
from
  fleetdm_carve
where
  host_id = 7 -- Replace with an actual host ID
order by
  id desc;
```
