---
title: "Steampipe Table: fleetdm_certificate_template - Query FleetDM Certificate Templates using SQL"
description: "Allows users to query the certificates FleetDM requests for each team's hosts."
---

# Table: fleetdm_certificate_template - Query FleetDM Certificate Templates using SQL

Certificate templates define the certificates Fleet requests from a certificate authority for a team's hosts. Uses the `/certificates` API endpoint.

## Table Usage Guide

Without a `team_id` filter the `fleetdm_certificate_template` table lists templates for hosts with no team (`team_id = 0`) plus every team, one request per team.

## Examples

### List certificate templates with their certificate authority

```sql+postgres
select
  t.name,
  t.subject_name,
  t.certificate_authority_name,
  t.team_id
from
  fleetdm_certificate_template t
order by
  t.team_id,
  t.name;
```

```sql+sqlite
select
  t.name,
  t.subject_name,
  t.certificate_authority_name,
  t.team_id
from
  fleetdm_certificate_template t
order by
  t.team_id,
  t.name;
```
