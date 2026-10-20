---
title: "Steampipe Table: fleetdm_host_certificate - Query FleetDM Host Certificates using SQL"
description: "Allows users to query the certificates installed on each FleetDM host."
---

# Table: fleetdm_host_certificate - Query FleetDM Host Certificates using SQL

Fleet collects the certificates in each host's keychains and certificate stores. Uses the `/hosts/{id}/certificates` API endpoint.

## Table Usage Guide

The `fleetdm_host_certificate` table lists certificates per host. Without a `host_id` filter the table queries every host (up to 10 requests at a time), which can take a while on a large fleet; filter by `host_id` for a single host.

## Examples

### Find certificates expiring in the next 30 days

```sql+postgres
select
  host_id,
  common_name,
  issuer_common_name,
  not_valid_after
from
  fleetdm_host_certificate
where
  not_valid_after < now() + interval '30 days'
order by
  not_valid_after;
```

```sql+sqlite
select
  host_id,
  common_name,
  issuer_common_name,
  not_valid_after
from
  fleetdm_host_certificate
where
  not_valid_after < datetime('now', '+30 days')
order by
  not_valid_after;
```
