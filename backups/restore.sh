#!/usr/bin/env bash
#
# Restores a Heroku PostgreSQL backup (pg_dump custom format) into the local
# geekswimmers database.
#
# Usage:
#   backups/restore.sh [path/to/backup.backup]
#
# If no path is given, the most recently modified *.backup file in
# backups/ is used. The database connection URL is read from config.toml
# unless DATABASE_URL is set in the environment.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKUPS_DIR="$ROOT_DIR/backups"
CONFIG_FILE="$ROOT_DIR/config.toml"

BACKUP_FILE="${1:-}"
if [ -z "$BACKUP_FILE" ]; then
    BACKUP_FILE="$(ls -t "$BACKUPS_DIR"/*.backup 2>/dev/null | head -n1 || true)"
    if [ -z "$BACKUP_FILE" ]; then
        echo "No .backup file found in $BACKUPS_DIR and none given as argument." >&2
        exit 1
    fi
fi

if [ ! -f "$BACKUP_FILE" ]; then
    echo "Backup file not found: $BACKUP_FILE" >&2
    exit 1
fi

if [ -z "${DATABASE_URL:-}" ]; then
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "config.toml not found and DATABASE_URL not set." >&2
        exit 1
    fi
    DATABASE_URL="$(awk -F'"' '/^[[:space:]]*url[[:space:]]*=/{print $2; exit}' "$CONFIG_FILE")"
    if [ -z "$DATABASE_URL" ]; then
        echo "Could not read database.url from $CONFIG_FILE." >&2
        exit 1
    fi
fi

DB_NAME="$(basename "$DATABASE_URL")"
DB_NAME="${DB_NAME%%\?*}"

echo "Backup file : $BACKUP_FILE"
echo "Database    : $DATABASE_URL"
echo
echo "This will DROP and REPLACE existing objects in the target database."
read -r -p "Continue? [y/N] " CONFIRM
case "$CONFIRM" in
    [yY]|[yY][eE][sS]) ;;
    *) echo "Aborted."; exit 1 ;;
esac

pg_restore \
    --no-owner \
    --no-privileges \
    --clean \
    --if-exists \
    --dbname="$DATABASE_URL" \
    "$BACKUP_FILE"

echo
echo "Restore finished."
