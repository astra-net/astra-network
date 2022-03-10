#!/bin/bash

ME=$(basename "$0")

function usage() {
   local MSG=$1

   cat<<-EOT
$MSG
This script will rclone the astra db to datadir/archive directory.

Usage: $ME [options] datadir shard

datadir:    the root directory of the astra db (default: /home/astra)
shard:      the shard number to sync (valid value: 0,1,2,3)

Options:
   -h       print this help message
   -c       clean up backup db after rclone
   -a       sync archival db, instead of regular db

EOT
   exit 1
}

CLEAN=false
FOLDER=mainnet.min
CONFIG=/etc/astra/rclone.conf

while getopts ":hca" opt; do
   case $opt in
      c) CLEAN=true ;;
      a) FOLDER=mainnet.archival ;;
      *) usage ;;
   esac
done

shift $((OPTIND - 1))

if [ $# != 2 ]; then
   usage
fi

DATADIR="$1"
SHARD="$2"

if [ ! -d "$DATADIR" ]; then
   usage "ERROR: no datadir directory: $DATADIR"
fi

case "$SHARD" in
   0|1|2|3) ;;
   *) usage "ERROR: invalid shard number: $SHARD" ;;
esac

mkdir -p "${DATADIR}/archive"

rclone --config "${CONFIG}" sync -vvv "astra:pub.astra.one/${FOLDER}/astra_db_${SHARD}" "${DATADIR}/archive/astra_db_${SHARD}" > "${DATADIR}/archive/archive-${SHARD}.log" 2>&1

[ -d "${DATADIR}/astra_db_${SHARD}" ] && mv -f "${DATADIR}/astra_db_${SHARD}" "${DATADIR}/archive/astra_db_${SHARD}.bak"
mv -f "${DATADIR}/archive/astra_db_${SHARD}" "${DATADIR}/astra_db_${SHARD}"

if $CLEAN; then
   rm -rf "${DATADIR}/archive/astra_db_${SHARD}.bak"
fi
