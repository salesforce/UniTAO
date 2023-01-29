import json
import requests

storage = {
    "name": "sys",
    "vhd": "vhd-wireguard-01-sys"
}

network = {
    "name": "eth0",
    "link": "vm-srv-wireguard-01-link-ext"
}

print("""
#######################################################################################
# Add Storage and Nic to VM for it to work
# 1, sys drive link to a VHD record
{}
# 2, create net adapter eth0 that bind to a VmLink entity
{}
#######################################################################################
""".format(json.dumps(storage, indent=4), json.dumps(network, indent=4)))

input("press ENTER to continue...")

print("""
#######################################################################################
# per schema of VirtualMachine, 
#  - storage link to VirtualHardDisk/StorageDevice
#  - network link to VmLink
# Get VirtualMachine Schema from http://localhost:8001/schema/VirtualMachine
#######################################################################################
""")

resp = requests.get("http://localhost:8001/schema/VirtualMachine")

respObj = resp.json()

print(json.dumps(respObj, indent=4))

print("""
#######################################################################################
# if we add reference item first, the patch will be failed
#######################################################################################
""")

input("press Enter to continue...")

print("patch http://localhost:8001/VirtualMachine/vm-srv-wireguard-01/storage {}".format(json.dumps(storage, indent=4)))

resp = requests.patch("http://localhost:8001/VirtualMachine/vm-srv-wireguard-01/storage", json=storage)

print(json.dumps(resp.json(), indent=4))

input("press ANY Key to continue...")

print("patch http://localhost:8001/VirtualMachine/vm-srv-wireguard-01/network {}".format(json.dumps(network, indent=4)))

resp = requests.patch("http://localhost:8001/VirtualMachine/vm-srv-wireguard-01/network", json=network)

print(json.dumps(resp.json(), indent=4))

input("press ANY Key to continue...")

vhd = {
    "__id": "vhd-wireguard-01-sys",
    "__type": "VirtualHardDisk",
    "__ver": "0.0.1",
    "data": {
        "name": "wireguard-01-sys",
        "path": "c:\compute\vm-srv-wireguard-01\sys.vhd"
    }
}

vmLink = {
    "__id": "vm-srv-wireguard-01-link-ext",
    "__type": "VmLink",
    "__ver": "0.0.1",
    "data": {
        "vm": "vm-srv-wireguard-01",
        "name": "ext"
    }
}

print("""
#######################################################################################
# Add referece records and try again
# VirtualHardDisk, VmLink
#######################################################################################
""")

input("press ANY key to continue...")

print("POST http://localhost:8001 {}",format(json.dumps(vhd, indent=4)))

resp = requests.post("http://localhost:8001", json=vhd)

print(resp.text)

input("press ANY key to continue...")

print("POST http://localhost:8001 {}",format(json.dumps(vmLink, indent=4)))

resp = requests.post("http://localhost:8001", json=vmLink)

print(resp.text)

input("press ANY key to continue...")

print("patch http://localhost:8001/VirtualMachine/vm-srv-wireguard-01/storage {}".format(json.dumps(storage, indent=4)))

respStorage = requests.patch("http://localhost:8001/VirtualMachine/vm-srv-wireguard-01/storage", json=storage)

print(json.dumps(respStorage.json(), indent=4))

input("press ANY Key to continue...")

print("patch http://localhost:8001/VirtualMachine/vm-srv-wireguard-01/network {}".format(json.dumps(network, indent=4)))

respNetwork = requests.patch("http://localhost:8001/VirtualMachine/vm-srv-wireguard-01/network", json=network)

print(json.dumps(respNetwork.json(), indent=4))
