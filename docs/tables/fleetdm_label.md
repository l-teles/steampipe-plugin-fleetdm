# Table: fleetdm_label

This table lists the labels defined in your FleetDM instance. Labels are used for grouping hosts, either manually or dynamically based on a SQL query.

## Columns

| Name                    | Type        | Description                                                                                         |
| ----------------------- | ----------- | --------------------------------------------------------------------------------------------------- |
| id                      | `INT`       | Unique ID of the label.                                                                             |
| name                    | `TEXT`      | Name of the label.                                                                                  |
| display_text            | `TEXT`      | Display text for the label, usually the same as the name.                                           |
| description             | `TEXT`      | Description of the label.                                                                           |
| query_sql               | `TEXT`      | The SQL query used for dynamic labeling. (Mapped from API field `query`)                            |
| platform                | `TEXT`      | Target platform(s) for the label (e.g., 'darwin', 'windows', 'linux', or empty for all).            |
| label_type              | `TEXT`      | Type of the label, e.g., 'regular' or 'builtin'.                                                    |
| label_membership_type   | `TEXT`      | Membership type, e.g., 'dynamic' or 'manual'.                                                       |
| host_count              | `INT`       | Number of hosts associated with this label.                                                         |
| built_in                | `BOOLEAN`   | Indicates if the label is a built-in label (derived from `label_type`).                             |
| created_at              | `TIMESTAMP` | Timestamp when the label was created.                                                               |
| updated_at              | `TIMESTAMP` | Timestamp when the label was last updated.                                                          |
| server_url              | `TEXT`      | FleetDM server URL from connection config.                                                          |

## Example Queries

**List all labels and their host counts:**
```sql
SELECT
  id,
  name,
  label_type,
  label_membership_type,
  host_count,
  platform
FROM
  fleetdm_label
ORDER BY
  name;
```

**Find all built-in labels:**
```sql
SELECT
  id,
  name,
  display_text,
  host_count
FROM
  fleetdm_label
WHERE
  built_in = TRUE;
```

**List dynamic labels and their queries:**
```sql
SELECT
  name,
  query_sql,
  host_count
FROM
  fleetdm_label
WHERE
  label_membership_type = 'dynamic'
ORDER BY
  name;
```

**Find labels specifically for macOS hosts:**
```sql
SELECT
  id,
  name,
  description
FROM
  fleetdm_label
WHERE
  platform = 'darwin';
