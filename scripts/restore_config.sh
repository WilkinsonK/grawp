#!/bin/sh
# Restore a backup of the server config file(s).
set -eux

SOURCE_DIRECTORY=${SOURCE_DIRECTORY:-server}
BACKUPS_DIRECTORY=${BACKUPS_DIRECTORY:-backups}
BACKUP_WORLD_DATE=${BACKUP_WORLD_DATE:-`date '+%m%d%Y'`}

unzip ${BACKUPS_DIRECTORY}/${SOURCE_DIRECTORY}.config.${BACKUP_WORLD_DATE}
exit $?
