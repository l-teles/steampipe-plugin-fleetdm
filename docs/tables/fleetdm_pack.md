# Table: fleetdm_pack

This table lists the query packs in your FleetDM instance. Packs are collections of queries that can be scheduled to run against targeted hosts or labels.

## Columns

| Name                            | Type        | Description                                                                                                |
| ------------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------- |
| id                              | `INT`       | Unique ID of the pack.                                                                                     |
| name                            | `TEXT`      | Name of the pack.                                                                                          |
| description                     | `TEXT`      | Description of the pack.                                                                                   |
| platform                        | `TEXT`      | Target platform(s) for the pack (comma-separated, or empty for all).                                       |
| disabled                        | `BOOLEAN`   | Indicates if the pack is disabled.                                                                         |
| type                            | `TEXT`      | Type of the pack (e.g., 'global', 'team').                                                                 |
| team_id                         | `INT`       | ID of the team the pack belongs to. Null if it's a global pack.                                            |
| target_count                    | `INT`       | Number of targets (hosts/labels/teams) for this pack.                                                      |
| total_scheduled_queries_count   | `INT`       | Total number of scheduled queries in this pack.                                                            |
| created_at                      | `TIMESTAMP` | Timestamp when the pack was created.                                                                       |
| updated_at                      | `TIMESTAMP` | Timestamp when the pack was last updated.                                                                  |
| targets                         | `JSONB`     | Target hosts, labels, and teams for this pack. (Details available on GET /api/v1/fleet/packs/{id})         |
| scheduled_queries               | `JSONB`     | Scheduled queries within this pack. (Details available on GET /api/v1/fleet/packs/{id})                    |
| agent_options                   | `JSONB`     | Agent options associated with the pack (if it's a team pack). (Details available on GET /api/v1/fleet/packs/{id}) |
| host_ids                        | `JSONB`     | List of host IDs targeted by this pack. (Details available on GET /api/v1/fleet/packs/{id})                |
| label_ids                       | `JSONB`     | List of label IDs targeted by this pack. (Details available on GET /api/v1/fleet/packs/{id})               |
| team_ids_targeted               | `JSONB`     | List of team IDs targeted by this pack, typically for global packs. (Details available on GET /api/v1/fleet/packs/{id}) |
| server_url                      | `TEXT`      | FleetDM server URL from connection config.                                                                 |

## Example Queries

**List all query packs:**
```sql
SELECT
  id,
  name,
  type,
  platform,
  disabled,
  target_count,
  total_scheduled_queries_count
FROM
  fleetdm_pack
ORDER BY
  name;
```

**Find disabled packs:**
```sql
SELECT
  id,
  name,
  type,
  platform
FROM
  fleetdm_pack
WHERE
  disabled = TRUE;
```

**List packs targeted at a specific platform, e.g., 'darwin':**
```sql
SELECT
  id,
  name,
  description
FROM
  fleetdm_pack
WHERE
  platform = 'darwin' OR platform = '' OR platform IS NULL -- '' or NULL often means all platforms
ORDER BY
  name;
```

**Get details for a specific pack, including its scheduled queries (requires `getPack` hydration):**
```sql
SELECT
  name,
  scheduled_queries
FROM
  fleetdm_pack
WHERE
  id = 1; -- Replace with an actual pack ID
