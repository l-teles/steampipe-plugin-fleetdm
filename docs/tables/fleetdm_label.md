---
title: "Steampipe Table: fleetdm_label - Query FleetDM Labels using SQL"
description: "Allows users to query FleetDM labels, providing insights into host groupings, dynamic labeling rules, and label distributions within your FleetDM instance."
---

# Table: fleetdm_label - Query FleetDM Labels using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. Labels in FleetDM are used to group hosts either manually or dynamically based on SQL queries, enabling flexible host categorization and management.

## Table Usage Guide

The `fleetdm_label` table provides comprehensive insights into label configurations within your FleetDM instance. As a system administrator, you can use this table to manage host groupings, monitor label distributions, and maintain dynamic labeling rules. The table helps you understand how hosts are categorized and how labels are being used across your fleet.

**Note:** This table supports the following optional key column for server-side API filtering:

- `team_id` — Filter labels by team (Fleet Premium). Use `'global'` to return only global labels.

Using this key column in your `WHERE` clause pushes the filtering to the FleetDM API.

## Examples

### List all labels and their host counts

Get an overview of all labels in your FleetDM instance, including their type and associated host counts.

```sql+postgres
select
  id,
  name,
  label_type,
  label_membership_type,
  host_count,
  platform
from
  fleetdm_label
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  label_type,
  label_membership_type,
  host_count,
  platform
from
  fleetdm_label
order by
  name;
```

### Find all system labels

Identify built-in system labels that are automatically managed by FleetDM.

```sql+postgres
select
  id,
  name,
  display_text,
  host_count
from
  fleetdm_label
where
  label_type = 'builtin';
```

```sql+sqlite
select
  id,
  name,
  display_text,
  host_count
from
  fleetdm_label
where
  label_type = 'builtin';
```

### List dynamic labels and their queries

Examine dynamic labels and their associated SQL queries to understand automated host categorization rules.

```sql+postgres
select
  name,
  query_sql,
  host_count
from
  fleetdm_label
where
  label_membership_type = 'dynamic'
order by
  name;
```

```sql+sqlite
select
  name,
  query_sql,
  host_count
from
  fleetdm_label
where
  label_membership_type = 'dynamic'
order by
  name;
```

### List labels for a specific team (Fleet Premium)

Filter labels by team to see only those relevant to a particular team.

```sql+postgres
select
  id,
  name,
  host_count
from
  fleetdm_label
where
  team_id = '5'
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  host_count
from
  fleetdm_label
where
  team_id = '5'
order by
  name;
```

### Find labels specifically for macOS hosts

Identify labels that are specifically configured for macOS devices.

```sql+postgres
select
  id,
  name,
  description
from
  fleetdm_label
where
  platform = 'darwin';
```

```sql+sqlite
select
  id,
  name,
  description
from
  fleetdm_label
where
  platform = 'darwin';
```
