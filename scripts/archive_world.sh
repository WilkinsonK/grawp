#!/bin/sh
# Create a backup of game world.
set -eux

SOURCE_DIRECTORY=${SOURCE_DIRECTORY:-server}
OUTPUT_DIRECTORY=${OUTPUT_DIRECTORY:-backups}

python3 scripts/archive_world.py $SOURCE_DIRECTORY $OUTPUT_DIRECTORY
exit $?
