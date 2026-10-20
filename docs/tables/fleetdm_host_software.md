---
title: "Steampipe Table: fleetdm_host_software - Query FleetDM Host Software using SQL"
description: "Allows users to query the software on each FleetDM host."
---

# Table: fleetdm_host_software - Query FleetDM Host Software using SQL

Fleet inventories the software on each host, with installed versions, paths, code signatures and vulnerabilities. Uses the `/hosts/{id}/software` API endpoint.

## Table Usage Guide

The `fleetdm_host_software` table lists software per host. Without a `host_id` filter the table queries every host (up to 10 requests at a time), which can take a while on a large fleet; filter by `host_id` for a single host. For fleet-wide counts, `fleetdm_software_title` is much cheaper. `software_package.package_url` is null unless `expose_secrets = true`, as installer URLs can embed credentials.

## Examples

### List the software on one host

```sql+postgres
select
  name,
  source,
  jsonb_array_length(installed_versions) as versions
from
  fleetdm_host_software
where
  host_id = 1
order by
  name;
```

```sql+sqlite
select
  name,
  source,
  json_array_length(installed_versions) as versions
from
  fleetdm_host_software
where
  host_id = 1
order by
  name;
```

### Find software with known vulnerabilities on one host

```sql+postgres
select
  s.name,
  v ->> 'version' as version,
  v -> 'vulnerabilities' as vulnerabilities
from
  fleetdm_host_software s,
  jsonb_array_elements(s.installed_versions) v
where
  s.host_id = 1
  and jsonb_typeof(v -> 'vulnerabilities') = 'array';
```

```sql+sqlite
select
  s.name,
  json_extract(v.value, '$.version') as version,
  json_extract(v.value, '$.vulnerabilities') as vulnerabilities
from
  fleetdm_host_software s,
  json_each(s.installed_versions) v
where
  s.host_id = 1
  and json_type(v.value, '$.vulnerabilities') = 'array';
```
