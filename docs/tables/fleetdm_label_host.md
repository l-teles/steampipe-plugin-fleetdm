---
title: "Steampipe Table: fleetdm_label_host - Query FleetDM Label Hosts using SQL"
description: "Allows users to query which hosts belong to each FleetDM label."
---

# Table: fleetdm_label_host - Query FleetDM Label Hosts using SQL

Labels group hosts by a query or a manual list. Uses the `/labels/{id}/hosts` API endpoint.

## Table Usage Guide

The `fleetdm_label_host` table has one row per host in each label. Without a `label_id` filter every label is read, which includes built-in labels such as All Hosts.

## Examples

### List the hosts in a label

```sql+postgres
select
  hostname,
  platform,
  status
from
  fleetdm_label_host
where
  label_id = 1;
```

```sql+sqlite
select
  hostname,
  platform,
  status
from
  fleetdm_label_host
where
  label_id = 1;
```

### Count hosts per label

```sql+postgres
select
  l.name,
  count(lh.host_id) as hosts
from
  fleetdm_label l
  left join fleetdm_label_host lh on lh.label_id = l.id
group by
  l.name
order by
  hosts desc;
```

```sql+sqlite
select
  l.name,
  count(lh.host_id) as hosts
from
  fleetdm_label l
  left join fleetdm_label_host lh on lh.label_id = l.id
group by
  l.name
order by
  hosts desc;
```
