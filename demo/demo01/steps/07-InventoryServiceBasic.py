import requests
import json

print("""
############################################################################################
# After sync up schema with all Data Service.
# Inventory Service can help explorer all data accross Data Services
# by following the ContentMediaType link defined within Schema
############################################################################################
""")

input("press ANY key to continue...")

print("""
###########################################################################################
# Local Data Types from Inventory Services [http://localhost:8004]:
# 1, inventory: store information about Data Services. how access them
# 2, schema: latest schema sync-ed from all data service
# 3, referral: store the relationship between dataType and DataService.
###########################################################################################
""")

input("press ANY key to continue...")

print("get http://localhost:8004/inventory")
resp = requests.get("http://localhost:8004/inventory")
print(json.dumps(resp.json(), indent=4))

print("get http://localhost:8004/inventory/DataService_01")
resp = requests.get("http://localhost:8004/inventory/DataService_01")
print(json.dumps(resp.json(), indent=4))

print("get http://localhost:8004/inventory/DataService_02")
resp = requests.get("http://localhost:8004/inventory/DataService_02")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("get http://localhost:8004/schema")
resp = requests.get("http://localhost:8004/schema")
print(json.dumps(resp.json(), indent=4))

print("get http://localhost:8004/schema/VirtualMachine")
resp = requests.get("http://localhost:8004/schema/VirtualMachine")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("get http://localhost:8004/referral")
resp = requests.get("http://localhost:8004/referral")
print(json.dumps(resp.json(), indent=4))

print("get http://localhost:8004/referral/VirtualMachine")
resp = requests.get("http://localhost:8004/referral/VirtualMachine")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("""
####################################################################################
# To explore data along network, we need to prepare several more information change
#  1, add link to NetBridge and VmHost from VmLink (fields[bridge, host])
#     make sure the new fields are optional so it's compatible with old schema version
#  2, add NetBridge[br-ext] Record that reflect NetBride entity
#  3, add 1 more Vm with VmLink and attach it to br-ext
#  4, attach both VmLink from both Vm to the NetBridge record to simulate connected machines
####################################################################################
""")

VmLinkSchemaV2 = {
    "__id": "VmLink",
    "__type": "schema",
    "__ver": "0.0.1",
    "data": {
        "name": "VmLink",
        "description": "the link entity that attached to VM nic card during creation",
        "version":"0.0.2",
        "key": "{vm}-link-{name}",
        "properties": {
            "vm": {
                "type": "string",
                "contentMediaType":"inventory/VirtualMachine"
            },
            "name": {
                "type":"string"
            },
            "bridge": {
                "type": "string",
                "contentMediaType": "inventory/NetBridge",
                "required": False
            },
            "host": {
                "type": "string",
                "contentMediaType": "inventory/VmHost",
                "required": False
            }
        }
    }
}

print("POST http://localhost:8001 {}".format(json.dumps(VmLinkSchemaV2, indent=4)))
resp = requests.post("http://localhost:8001", json=VmLinkSchemaV2)
print(resp.text)

print("GET http://localhost:8001/schema/VmLink")

resp = requests.get("http://localhost:8001/schema/VmLink")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("""
Upgrade VmLink to newer version [0.0.2] directly since 0.0.2 is compatible with 0.0.1
PATCH http://localhost:8001/VmLink/vm-srv-wireguard-01-link-ext/__ver '0.0.2'
""")
resp = requests.patch("http://localhost:8001/VmLink/vm-srv-wireguard-01-link-ext/__ver", "0.0.2")
print(json.dumps(resp.json(), indent=4))

print("""
Add NetBridge[br-ext] and attach VmLink to this bridge add VmHost name      
""")
input("press ANY key to continue...")

bridge = {
    "__id": "br-ext",
    "__type": "NetBridge",
    "__ver": "0.0.1",
    "data": {
        "name": "ext",
        "links": []
    }
}
print("POST http://localhost:8001 {}".format(json.dumps(bridge, indent=4)))
resp = requests.post("http://localhost:8001", json=bridge)
print(resp.text)

input("press ANY key to continue...")

print("PATCH http://localhost:8001/VmLink/vm-srv-wireguard-01-link-ext/bridge 'br-ext'")
resp = requests.patch("http://localhost:8001/VmLink/vm-srv-wireguard-01-link-ext/bridge", "br-ext")
print(json.dumps(resp.json(), indent=4))

print("PATCH http://localhost:8001/VmLink/vm-srv-wireguard-01-link-ext/host 'vmhost-test01'")
resp = requests.patch("http://localhost:8001/VmLink/vm-srv-wireguard-01-link-ext/host", "vmhost-test01")
print(json.dumps(resp.json(), indent=4))

print("PATCH http://localhost:8001/NetBridge/br-ext/links 'vm-srv-wireguard-01-link-ext'")
resp = requests.patch("http://localhost:8001/NetBridge/br-ext/links", "vm-srv-wireguard-01-link-ext")
print(json.dumps(resp.json(), indent=4))

