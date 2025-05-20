---
title: "Steampipe Table: fleetdm_team - Query FleetDM Teams using SQL"
description: "Allows users to query FleetDM teams, providing insights into team configurations, user assignments, and host distributions within your FleetDM instance."
---

# Table: fleetdm_team - Query FleetDM Teams using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. Teams in FleetDM are used to group users and can have their own set of hosts, policies, and configurations, enabling granular access control and management.

## Table Usage Guide

The `fleetdm_team` table provides comprehensive insights into team configurations within your FleetDM instance. As a system administrator, you can use this table to manage team structures, monitor team sizes, and track host distributions across different teams. The table helps you understand team organization, resource allocation, and access control patterns.

## Examples

### List all teams and their user/host counts
Get an overview of all teams in your FleetDM instance, including their size and resource distribution.

```sql+postgres
select
  id,
  name,
  description,
  user_count,
  host_count,
  created_at
from
  fleetdm_team
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  description,
  user_count,
  host_count,
  created_at
from
  fleetdm_team
order by
  name;
```

### Find teams with more than 100 hosts
Identify large teams that may need additional management attention or resource allocation.

```sql+postgres
select
  id,
  name,
  host_count
from
  fleetdm_team
where
  host_count > 100
order by
  host_count desc;
```

```sql+sqlite
select
  id,
  name,
  host_count
from
  fleetdm_team
where
  host_count > 100
order by
  host_count desc;
```

### Get agent options for a specific team
Examine the configuration settings for a particular team to ensure proper agent behavior.

```sql+postgres
select
  name,
  agent_options -> 'config' -> 'decorators' ->> 'load' as agent_config_decorators_load
from
  fleetdm_team
where
  name = 'Engineering';
```

```sql+sqlite
select
  name,
  json_extract(json_extract(agent_options, '$.config'), '$.decorators.load') as agent_config_decorators_load
from
  fleetdm_team
where
  name = 'Engineering';
```
