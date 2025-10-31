#!/bin/sh
# Create a backup of the server config file(s).
set -eux

SOURCE_DIRECTORY=${SOURCE_DIRECTORY:-server}
OUTPUT_DIRECTORY=${OUTPUT_DIRECTORY:-backups}

python3 scripts/archive_config.py $SOURCE_DIRECTORY $OUTPUT_DIRECTORY
exit $?
