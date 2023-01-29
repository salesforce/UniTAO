import requests
import json

print("""
#################################################################################
#   Up till now, all operation are done in the same DataService.
#   As this step, we are going to demo the same reference validation capability
#   Across DataServices
#   1, DataType (VmHost) in DataService02 with value CMT reference to DataService01 entities
#   2, Create VmHost Record in DataService02
#   3, add value in VmHost that reference to DataService01 entities
#################################################################################
""")

input("press ANY key to continue...")

print("get http://localhost:8002/schema")

resp = requests.get("http://localhost:8002/schema")

print(json.dumps(resp.json(), indent=4))

print("get http://localhost:8002/schema/VmHost")

resp = requests.get("http://localhost:8002/schema/VmHost")

print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

vmHost = {
    "__id": "vmhost-test01",
    "__type": "VmHost",
    "__ver": "0.0.1",
    "data": {
        "name": "test01",
        "netBridge": [],
        "vmLink": [],
        "virtualHardDisk": [],
        "storageDevice": [],
        "virtualMachine": [],
        "nic": []
    }
}
print("post http://localhost:8002 {}".format(json.dumps(vmHost, indent=4)))

resp = requests.post("http://localhost:8002", json=vmHost)

print(resp.text)

input("press ANY key to continue...")

print("""
#################################################################################
# Now try add reference from DataService02 VmHost to DataService01 Entities
# netBridge -> NetBridge
# vmLink -> VmLink
# virtualHardDisk -> VirtualHardDisk
# storageDevice -> StorageDevice
# virtualMachine -> VirtualMachine
# nic -> NetworkCard
#################################################################################
""")

input("press ANY key to continue...")

nicEth001 = {
            "__id": "nic-eth1000-aef",
            "__type": "NetworkCard",
            "__ver": "0.0.1",
            "data": {
                "name": "eth1000-aef"
            }
        }

print("""
try to add not exists network card

PATCH http://localhost:8002/VmHost/vmhost-test01/nic nic-eth1000-aef
""")

resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/nic", "nic-eth1000-aef")

print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("""
Add Network Card entity [nic-eth1000-aef]

POST http://localhost:8001 {}
""".format(json.dumps(nicEth001, indent=4)))

resp = requests.post("http://localhost:8001", json=nicEth001)

print(resp.text)

input("press ANY key to continue...")

print("""
Add reference to NetworkCard [nic-eth1000-aef]
PATCH http://localhost:8002/VmHost/vmhost-test01/nic nic-eth1000-aef
""")

resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/nic", "nic-eth1000-aef")

print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("""
#################################################################################
# Add rest of the references
# vmLink -> VmLink
# virtualHardDisk -> VirtualHardDisk
# storageDevice -> StorageDevice
# virtualMachine -> VirtualMachine
# nic -> NetworkCard
#################################################################################
""")

input("press ANY key to continue...")

print("""
Add reference to Virtual Hard Disk
    PATCH http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk vhd-wireguard-01-sys
""")
resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk", "vhd-wireguard-01-sys")
print(json.dumps(resp.json(), indent=4))

print("""
Add reference to VmLink
    PATCH http://localhost:8002/VmHost/vmhost-test01/vmLink vm-srv-wireguard-01-link-ext
""")

resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/vmLink", "vm-srv-wireguard-01-link-ext")
print(json.dumps(resp.json(), indent=4))

print("""
Add referenceto Virtual Machine
    PATCH http://localhost:8002/VmHost/vmhost-test01/virtualMachine vm-srv-wireguard-01
""")

resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/virtualMachine", "vm-srv-wireguard-01")
print(json.dumps(resp.json(), indent=4))
