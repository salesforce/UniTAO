#!/bin/bash -v
function pause(){
   read -p "$*"
}
clear

#==================================
# list all schema in dataservice 01
#==================================

curl http://localhost:8001/schema 

pause