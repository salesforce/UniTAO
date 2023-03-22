import requests
import json
import time

print("""
######################################################################################################################
# With attribute[indexTemplate] introduced under array/map items definition.
# we can now automatically have VmHost register newly created entities.
# 1, upgrade schema of VmHost that point to all entities from DataService01
# 2, upgrade schema version of vmhost-test01 to new version 0.0.2
# 3, add another record of VirtualHardDrive
# 4, see that new vhd record registered under http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk
# 5, remove the record of VirtualHardDrive
# 6, see that reference being removed from http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk
######################################################################################################################
""")
input("press ANY KEY to continue...")

print("""
######################################################################################################################
# 1, upgrade schema of VmHost that point to all entities from DataService01
######################################################################################################################
""")

vmhostSchemaV2 = {
    "__id": "VmHost",
    "__type": "schema",
    "__ver": "0.0.1",
    "data": {
        "name": "VmHost",
        "description": "Physical Machine that host all virtual machines",
        "version":"0.0.2",
        "key": "vmhost-{name}",
        "properties": {
            "name": {
                "type": "string"
            },
            "netBridge": {
                "type": "array",
                "items": {
                    "type": "string",
                    "contentMediaType": "inventory/NetBridge",
                    "indexTemplate": "VmHost/{host}/netBridge"
                }
            },
            "vmLink": {
                "type": "array",
                "items": {
                    "type": "string",
                    "contentMediaType": "inventory/VmLink",
                    "indexTemplate": "VmHost/{host}/vmLink"
                }
            },
            "virtualMachine": {
                "type": "array",
                "items": {
                    "type": "string",
                    "contentMediaType": "inventory/VirtualMachine",
                    "indexTemplate": "VmHost/{host}/virtualMachine"
                }
            },
            "nic": {
                "type": "array",
                "items": {
                    "type": "string",
                    "contentMediaType": "inventory/NetworkCard",
                    "indexTemplate": "VmHost/{host}/nic"
                }
            },
            "virtualHardDisk": {
                "type": "array",
                "items": {
                    "type": "string",
                    "contentMediaType": "inventory/VirtualHardDisk",
                    "indexTemplate": "VmHost/{host}/virtualHardDisk"
                }
            },
            "storageDevice": {
                "type": "array",
                "items": {
                    "type": "string",
                    "contentMediaType": "inventory/StorageDevice",
                    "indexTemplate": "VmHost/{host}/storageDevice"
                }
            }
        }
    }
}
print("POST http://localhost:8002 {}".format(json.dumps(vmhostSchemaV2, indent=4)))
resp = requests.post("http://localhost:8002", json=vmhostSchemaV2)
print(resp.text)

print("""
######################################################################################################################
# with schema include [indexTemplate], 
# the Journal Process will create corresponding Record of [cmtIdx] to 
# subscribe data change event from CMT reference in this case:
# Data Types from DataService01:
# NetBridge, VmLink, VirtualMachine, NetworkCard, VirtualHardDisk, StorageDevice
######################################################################################################################
""")

print("sleep 10 second for journal process to work on the data subscription of [indexTemplate]")
time.sleep(10)

print("GET http://localhost:8001/cmtIdx")
resp = requests.get("http://localhost:8001/cmtIdx")
print(json.dumps(resp.json(), indent=4))

input("press ANY KEY to continue...")

print("""
######################################################################################################################
# schema of cmtIdx and example cmtIdx subscription on NetBridge
######################################################################################################################
""")

input("press ANY KEY to continue...")

print("GET http://localhost:8001/schema/cmtIdx")
resp = requests.get("http://localhost:8001/schema/cmtIdx")
print(json.dumps(resp.json(), indent=4))

input("press ANY KEY to continue...")

print("GET http://localhost:8001/cmtIdx/NetBridge")
resp = requests.get("http://localhost:8001/cmtIdx/NetBridge")
print(json.dumps(resp.json(), indent=4))

input("press ANY KEY to continue...")

print("""
######################################################################################################################
# 2, upgrade schema version of vmhost-test01 to new version 0.0.2
######################################################################################################################
""")

print("PATCH http://localhost:8002/VmHost/vmhost-test01/__ver '0.0.2'")

resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/__ver", "0.0.2")

print(json.dumps(resp.json(), indent=4))

input("press ANY KEY to continue...")

print("""
######################################################################################################################
# 3, add another record of VirtualHardDisk
######################################################################################################################
""")

vhd = {
    "__id": "vhd-vm-test01-sys",
    "__type": "VirtualHardDisk",
    "__ver": "0.0.1",
    "data": {
        "name": "vm-test01-sys",
        "path": "/opt/vhd/vm-test01-sys.vhd",
        "owner": "vm-test01",
        "alias": "sys",
        "host": "vmhost-test01"
    },
}

print("POST http://localhost:8001 {}".format(json.dumps(vhd, indent=4)))
resp = requests.post("http://localhost:8001", json=vhd)
print(resp.text)

print("GET http://localhost:8002/VmHost/vmhost-test01")
resp = requests.get("http://localhost:8002/VmHost/vmhost-test01")
print(json.dumps(resp.json(), indent=4))

print("sleep 10 second")
time.sleep(10)

print("""
######################################################################################################################
# 4, see that new vhd record registered under http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk
######################################################################################################################
""")

print("GET http://localhost:8002/VmHost/vmhost-test01")
resp = requests.get("http://localhost:8002/VmHost/vmhost-test01")
print(json.dumps(resp.json(), indent=4))

input("press ANY KEY to continue...")

print("""
######################################################################################################################
# 5, remove the record of VirtualHardDrive
######################################################################################################################
""")

print("DELETE http://localhost:8001/VirtualHardDisk/vhd-vm-test01-sys")
resp = requests.delete("http://localhost:8001/VirtualHardDisk/vhd-vm-test01-sys")
print("status {}".format(resp.status_code))

print("""
######################################################################################################################
# 6, see that reference being removed from http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk
######################################################################################################################
""")

print("GET http://localhost:8002/VmHost/vmhost-test01")
resp = requests.get("http://localhost:8002/VmHost/vmhost-test01")
print(json.dumps(resp.json(), indent=4))

print("sleep 5 second")
time.sleep(5)

print("GET http://localhost:8002/VmHost/vmhost-test01")
resp = requests.get("http://localhost:8002/VmHost/vmhost-test01")
print(json.dumps(resp.json(), indent=4))
