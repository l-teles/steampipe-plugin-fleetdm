---
title: "Steampipe Table: fleetdm_policy - Query FleetDM Policies using SQL"
description: "Allows users to query FleetDM policies, providing insights into security configurations, compliance status, and policy distributions within your FleetDM instance."
---

# Table: fleetdm_policy - Query FleetDM Policies using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. Policies in FleetDM are osquery queries that determine if a host is passing or failing specific security or configuration guidelines, enabling automated compliance monitoring and enforcement.

## Table Usage Guide

The `fleetdm_policy` table provides comprehensive insights into policy configurations within your FleetDM instance. As a security administrator or compliance officer, you can use this table to monitor policy compliance, track failing hosts, and maintain security standards. The table helps you understand policy effectiveness and compliance status across your fleet.

**Note:** This table uses different API endpoints depending on filtering:

- **Without `team_id`**: Uses `GET /global/policies` (global policies, no text search support).
- **With `team_id`**: Uses `GET /teams/:id/policies` (team policies with additional filter support).

Supported key columns:

- `team_id` — Filter by team. Switches to the team policies endpoint.
- `filter_search_query` — Search policies by name or query text. **Only works when `team_id` is specified.**
- `merge_inherited` — Include global policies in team policy results (Fleet Premium). Requires `team_id`.

## Examples

### List all global policies and their pass/fail counts

Get an overview of all global policies and their current compliance status across your fleet.

```sql+postgres
select
  id,
  name,
  platform,
  passing_host_count,
  failing_host_count
from
  fleetdm_policy
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  platform,
  passing_host_count,
  failing_host_count
from
  fleetdm_policy
order by
  name;
```

### Find critical policies

Identify critical policies and their current failure rates to prioritize remediation efforts.

```sql+postgres
select
  id,
  name,
  query_text,
  failing_host_count
from
  fleetdm_policy
where
  critical = true
order by
  failing_host_count desc;
```

```sql+sqlite
select
  id,
  name,
  query_text,
  failing_host_count
from
  fleetdm_policy
where
  critical = true
order by
  failing_host_count desc;
```

### List policies for a specific team

View all policies defined for a particular team.

```sql+postgres
select
  id,
  name,
  passing_host_count,
  failing_host_count
from
  fleetdm_policy
where
  team_id = 5
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  passing_host_count,
  failing_host_count
from
  fleetdm_policy
where
  team_id = 5
order by
  name;
```

### List team policies including inherited global policies

Use `merge_inherited` to include global policies alongside team-specific policies.

```sql+postgres
select
  id,
  name,
  team_id,
  critical
from
  fleetdm_policy
where
  team_id = 5
  and merge_inherited = true
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  team_id,
  critical
from
  fleetdm_policy
where
  team_id = 5
  and merge_inherited = true
order by
  name;
```

### Search for policies related to "firewall"

Find policies that are specifically related to firewall configurations or monitoring.

```sql+postgres
select
  id,
  name,
  query_text
from
  fleetdm_policy
where
  description like '%firewall%' or name like '%Firewall%';
```

```sql+sqlite
select
  id,
  name,
  query_text
from
  fleetdm_policy
where
  description like '%firewall%' or name like '%Firewall%';
```

### Count policies by platform

Analyze the distribution of policies across different operating system platforms.

```sql+postgres
select
  platform,
  count(*) as policy_count
from
  fleetdm_policy
group by
  platform
order by
  policy_count desc;
```

```sql+sqlite
select
  platform,
  count(*) as policy_count
from
  fleetdm_policy
group by
  platform
order by
  policy_count desc;
```
