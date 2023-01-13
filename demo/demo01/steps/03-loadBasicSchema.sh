#!/bin/bash -v
function pause(){
   read -p "$*"
}

SCRIPT=$(realpath "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
pushd $SCRIPTPATH

clear

#=====================================================================
# Import basic schemas to DataService01
#=====================================================================
../../lib/python/submitDsData.py -m ../data/dsMap.json -d ../data/basicSchema.json

pause
popd
