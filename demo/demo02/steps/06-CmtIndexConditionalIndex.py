import requests
import json
import time

print("""
##############################################################################################################################
# the indexTemplate format: [{registryType}/{registryId}/{path to registry attribute}]
# the CmtIndex process will use the target record attribute value to build the registry path.
# if path does not exists, the registry op will not happen
# when new path created, the CmtIndex process will scan all contentMediaType records to fill the registry attribute
# this way, we are using registry record data to filter what to fill in the registry attribute
# Steps:
#  1, upgrade schema of VirtualMachine to subscribe virtualHardDrive with indexTemplate
#  2, add VirtualMachine record as newer version 0.0.2 with 1 storage device[sys]
#  3, add new VirtualHardDrive record and point it to Vm, so only vhd for sys will show up
#  4, add entry route [data] for storage, so vhd for data will show up
#  5, remove vhd for sys, and see the entry being removed from VirtualMachine
#  6, add back vhd record for sys, and see the entry being put back into VirtualMachine
##############################################################################################################################
""")
input("press ANY KEY to continue...")

print("""
Current VirtualMachine Schema:
GET http://localhost:8001/schema/VirtualMachine
""")

resp = requests.get("http://localhost:8001/schema/VirtualMachine")
print(json.dumps(resp.json(), indent=4))

print("""
##############################################################################################################################
#  1, upgrade schema of VirtualMachine to subscribe virtualHardDrive with indexTemplate
##############################################################################################################################
""")

input("press ANY KEY to continue...")

vmSchemaV2 = {
    "__id": "VirtualMachine",
    "__type": "schema",
    "__ver": "0.0.1",
    "data": {
        "name": "VirtualMachine",
        "description": "virtual machine",
        "version": "0.0.2",
        "key": "vm-{name}",
        "properties": {
            "name": {
                "type": "string"
            },
            "storage": {
                "type": "array",
                "items": {
                    "type": "object",
                    "$ref": "#/definitions/storage"
                }
            },
            "network": {
                "type": "array",
                "items": {
                    "type": "object",
                    "$ref": "#/definitions/nic"
                }
            },
            "host": {
                "type": "string",
                "contentMediaType": "inventory/VmHost"
            }
        },
        "definitions": {
            "nic": {
                "key": "{name}",
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "ip": {
                        "type": "string",
                        "required": False
                    },
                    "link": {
                        "type": "array",
                        "items": {
                            "type": "string",
                            "contentMediaType": "inventory/VmLink",
                            "indexTemplate": "VirtualMachine/{owner}/network[{alias}]/link"
                        }
                        
                    }
                }
            },
            "storage": {
                "key": "{name}",
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "virtualDrive": {
                        "type": "array",
                        "items": {
                            "type": "string",
                            "contentMediaType": "inventory/VirtualHardDisk",
                            "indexTemplate": "VirtualMachine/{owner}/storage[{alias}]/virtualDrive"
                        }
                    },
                    "device": {
                        "type": "array",
                        "items": {
                            "type": "string",
                            "contentMediaType": "inventory/StorageDevice",
                            "indexTemplate": "VirtualMachine/{owner}/storage[{alias}]/device"
                        }
                    }
                    
                }
            }
        }
    }
}
print("POST http://localhost:8001 {}".format(json.dumps(vmSchemaV2, indent=4)))
resp = requests.post("http://localhost:8001", json=vmSchemaV2)
print(resp.text)

input("press ANY KEY to continue...")

print("""
##############################################################################################################################
#  2, add VirtualMachine record as newer version 0.0.2 with 1 storage device[sys]
##############################################################################################################################
""")

input("press ANY KEY to continue...")

vm = {
    "__id": "vm-test01",
    "__type": "VirtualMachine",
    "__ver": "0.0.2",
    "data": {
        "name": "test01",
        "storage": [
            {
                "name": "sys",
                "virtualDrive": [],
                "device": []
            }
        ],
        "network": [],
        "host": "vmhost-test01"
    }
}

print("POST http://localhost:8001 {}".format(json.dumps(vm, indent=4)))
resp = requests.post("http://localhost:8001", json=vm)
print(resp.text)

print("get http://localhost:8002/VmHost/vmhost-test01")
resp = requests.get("http://localhost:8002/VmHost/vmhost-test01")
print(json.dumps(resp.json(), indent=4))

print("sleep 1 second")
time.sleep(1)

print("get http://localhost:8002/VmHost/vmhost-test01")
resp = requests.get("http://localhost:8002/VmHost/vmhost-test01")
print(json.dumps(resp.json(), indent=4))

input("press ANY KEY to continue...")

print("""
##############################################################################################################################
#  3, add new VirtualHardDrive record and point it to Vm
#     with 2 alias: [sys, data]
#     only vhd for sys will be registered, since only entry of sys exists
##############################################################################################################################
""")

input("press ANY KEY to continue...")

vhdList = [
    {
        "__id": "vhd-vm-test01-sys",
        "__type": "VirtualHardDisk",
        "__ver": "0.0.1",
        "data": {
            "name": "vm-test01-sys",
            "path": "/opt/vhd/vm-test01-sys.vhd",
            "owner": "vm-test01",
            "alias": "sys",
            "host": "vmhost-test01"
        }
    },
    {
        "__id": "vhd-vm-test01-data",
        "__type": "VirtualHardDisk",
        "__ver": "0.0.1",
        "data": {
            "name": "vm-test01-data",
            "path": "/opt/vhd/vm-test01-data.vhd",
            "owner": "vm-test01",
            "alias": "data",
            "host": "vmhost-test01"
        }
    }
]

for vhd in vhdList:
    print("POST http://localhost:8001 {}".format(json.dumps(vhd, indent=4)))
    resp = requests.post("http://localhost:8001", json=vhd)
    
print("sleep 1 second")
time.sleep(1)

print("GET http://localhost:8001/VirtualMachine/vm-test01")
resp = requests.get("http://localhost:8001/VirtualMachine/vm-test01")
print(json.dumps(resp.json(), indent=4))

print("""
##############################################################################################################################
#  4, add entry route [data] for storage, so vhd for data will show up
##############################################################################################################################
""")

input("press ANY KEY to continue...")

dataStorage = {
    "name": "data",
    "virtualDrive": [],
    "device": []
}
print("PATCH http://localhost:8001/VirtualMachine/vm-test01/storage {}".format(json.dumps(dataStorage, indent=4)))
resp = requests.patch("http://localhost:8001/VirtualMachine/vm-test01/storage", json=dataStorage)
print(json.dumps(resp.json(), indent=4))

print("sleep for 1 second")
time.sleep(1)

print("get http://localhost:8001/VirtualMachine/vm-test01")
resp = requests.get("http://localhost:8001/VirtualMachine/vm-test01")
print(json.dumps(resp.json(), indent=4))

input("press ANY KEY to continue...")

print("""
##############################################################################################################################
#  5, remove vhd for sys, and see the entry being removed from VirtualMachine
##############################################################################################################################
""")
input("press ANY KEY to continue...")

print("delete http://localhost:8001/VirtualHardDisk/vhd-vm-test01-sys")
resp = requests.delete("http://localhost:8001/VirtualHardDisk/vhd-vm-test01-sys")
print("status code: [{}]".format(resp.status_code))

print("get http://localhost:8001/VirtualMachine/vm-test01")
resp = requests.get("http://localhost:8001/VirtualMachine/vm-test01")
print(json.dumps(resp.json(), indent=4))

print("sleep for 1 second")
time.sleep(1)

print("get http://localhost:8001/VirtualMachine/vm-test01")
resp = requests.get("http://localhost:8001/VirtualMachine/vm-test01")
print(json.dumps(resp.json(), indent=4))

input("press ANY KEY to continue...")

print("""
##############################################################################################################################
#  6, add back vhd record for sys, and see the entry being put back into VirtualMachine
##############################################################################################################################
""")

vm = {
    "__id": "vhd-vm-test01-sys01",
    "__type": "VirtualHardDisk",
    "__ver": "0.0.1",
    "data": {
        "name": "vm-test01-sys01",
        "path": "/opt/vhd/vm-test01-sys01.vhd",
        "owner": "vm-test01",
        "alias": "sys",
        "host": "vmhost-test01"
    }
}
print("POST http://localhost:8001 {}".format(json.dumps(vm, indent=4)))
resp = requests.post("http://localhost:8001", json=vm)
print(resp.text)

print("get http://localhost:8001/VirtualMachine/vm-test01")
resp = requests.get("http://localhost:8001/VirtualMachine/vm-test01")
print(json.dumps(resp.json(), indent=4))

print("sleep for 1 second")
time.sleep(1)

print("get http://localhost:8001/VirtualMachine/vm-test01")
resp = requests.get("http://localhost:8001/VirtualMachine/vm-test01")
print(json.dumps(resp.json(), indent=4))
