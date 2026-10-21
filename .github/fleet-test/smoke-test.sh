#!/usr/bin/env bash
# smoke-test.sh – Build the plugin, load it into Steampipe and query every
# table it exposes against a live Fleet. Catches runtime breakage (SDK/cache
# regressions, API envelope changes) that unit tests cannot see.
#
# Requires steampipe on PATH and FLEETDM_URL / FLEETDM_API_TOKEN in the
# environment (from setup-fleet.sh). The connection config holds no
# credentials: the plugin reads them from the environment. It refuses a
# non-local Fleet unless SMOKE_ALLOW_REMOTE=1.
#
# It runs its own Steampipe (install dir STEAMPIPE_INSTALL_DIR, by default a
# cache dir, service on SMOKE_DB_PORT), so it never touches ~/.steampipe or a
# service you already run, and every run loads the plugin it just built.

set -euo pipefail

default_install_dir="${XDG_CACHE_HOME:-$HOME/.cache}/fleetdm-smoke-steampipe"
INSTALL_DIR="${STEAMPIPE_INSTALL_DIR:-$default_install_dir}"
export STEAMPIPE_INSTALL_DIR="$INSTALL_DIR"
SMOKE_DB_PORT="${SMOKE_DB_PORT:-9194}"
CONN="fleetdm_smoke"
: "${FLEETDM_URL:?FLEETDM_URL must be set}"
: "${FLEETDM_API_TOKEN:?FLEETDM_API_TOKEN must be set}"

case "$FLEETDM_URL" in
  http://localhost:* | http://127.0.0.1:* | http://localhost | http://127.0.0.1) ;;
  *)
    if [ "${SMOKE_ALLOW_REMOTE:-}" != "1" ]; then
      echo "ERROR: FLEETDM_URL=${FLEETDM_URL} is not a local test Fleet; set SMOKE_ALLOW_REMOTE=1 to query it anyway." >&2
      exit 1
    fi
    ;;
esac

# Tables with required key columns get a qual; a missing row is not an error.
extra_where() {
  case "$1" in
    fleetdm_host_detail) echo "where id = 0" ;;
  esac
}

plugin_dir="${INSTALL_DIR}/plugins/local/fleetdmsmoke"   # never a developer's own local/fleetdm
spc="${INSTALL_DIR}/config/${CONN}.spc"
if [ -e "$plugin_dir" ]; then
  if [ "$INSTALL_DIR" != "$default_install_dir" ]; then
    echo "ERROR: ${plugin_dir} already exists; this script would replace and then delete it." >&2
    exit 1
  fi
  rm -rf "$plugin_dir" # left by an interrupted run in the script's own install dir
fi
stop_service=0
# shellcheck disable=SC2329 # invoked by the EXIT trap
cleanup() {
  rm -f "$spc"
  rm -rf "$plugin_dir"
  if [ "$stop_service" = 1 ]; then
    steampipe service stop >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

mkdir -p "$plugin_dir" "${INSTALL_DIR}/config"
go build -o "${plugin_dir}/fleetdmsmoke.plugin" .

cat > "$spc" <<EOF
connection "${CONN}" {
  plugin = "local/fleetdmsmoke"
}
EOF

# Each standalone `steampipe query` shuts its database down when it ends,
# which can abort the background load of a new connection (seen on Linux:
# "RefreshConnections failed ... closed pool"). Keep a service up instead.
# A service that is already running keeps old plugin processes alive, so the
# new build might not load: start a fresh one (loopback only), stop it on exit.
# "Not running" text differs on a fresh install, so test for running.
if steampipe service status 2>&1 | grep -q 'Steampipe service is running'; then
  if [ "$INSTALL_DIR" != "$default_install_dir" ]; then
    echo "ERROR: a Steampipe service is running in ${INSTALL_DIR}; stop it or use another STEAMPIPE_INSTALL_DIR." >&2
    exit 1
  fi
  steampipe service stop --force >/dev/null 2>&1 || true # left by an interrupted run
fi
if ! out=$(steampipe service start --database-listen local --database-port "$SMOKE_DB_PORT" 2>&1); then
  echo "ERROR: could not start the Steampipe service: $out" >&2
  exit 1
fi
stop_service=1

# A new connection loads asynchronously; wait until it is ready (a schema
# left over from an earlier run can exist before then), then list its tables.
state=""
for _ in $(seq 1 20); do
  state=$(steampipe query --output csv --header=false \
    "select state from steampipe_connection where name = '${CONN}'" 2>/dev/null) || state=""
  case "$state" in
    ready) break ;;
    error)
      echo "ERROR: connection ${CONN} failed to load:" >&2
      steampipe query --output csv --header=false "select error from steampipe_connection where name = '${CONN}'" >&2 || true
      exit 1
      ;;
  esac
  sleep 3
done
if [ "$state" != "ready" ]; then
  echo "ERROR: connection ${CONN} did not become ready (state: ${state:-unknown})." >&2
  exit 1
fi
tables=$(steampipe query --output csv --header=false \
  "select table_name from information_schema.tables where table_schema = '${CONN}' order by table_name")
if [ -z "$tables" ]; then
  echo "ERROR: plugin exposed no tables; it failed to load." >&2
  exit 1
fi

failed=0
for t in $tables; do
  sql="select count(*) from ${CONN}.${t} $(extra_where "$t")"
  # steampipe can exit 0 on a failed query, so also inspect the output.
  if out=$(steampipe query --output csv --header=false "$sql" 2>&1) && ! grep -qE '^(Error|Warning)' <<<"$out"; then
    printf '%-32s ok (%s rows)\n' "$t" "$out"
  else
    printf '%-32s FAILED: %s\n' "$t" "$out"
    failed=1
  fi
done

exit "$failed"
