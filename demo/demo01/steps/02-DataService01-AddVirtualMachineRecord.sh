#!/bin/bash -v
function pause(){
   read -p "$*"
}
clear

#=====================================================================
# TRY to create a new record (w/o VirtualMachine Schema)
#=====================================================================

curl -i -X POST http://localhost:8001  -d '
{
    "__id": "vm-srv-wireguard-01",
    "__type": "VirtualMachine",
    "__ver": "0.0.1",
    "data": {
        "name": "srv-wireguard-01"
    }
}
'

pause