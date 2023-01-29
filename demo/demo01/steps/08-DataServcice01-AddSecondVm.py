import requests
import json

print("""
###############################################################################################################
# Add Second VM that attach to NetBridge[br-ext] in VmHost[vmhost-test01]
#  1, Create VirtualMachine[vm-srv-wireguard-02]
#  2, Create VirtualHardDisk[vhd-wireguard-02-sys] and attach it to VM
#  3, Create VmLink[vm-srv-wireguard-01-link-ext] and attach it to VM, NetBridge
###############################################################################################################
""")

input("press ANY key to continue...")

vm = {
    "__id": "vm-srv-wireguard-02",
    "__type": "VirtualMachine",
    "__ver": "0.0.1",
    "data": {
        "name": "srv-wireguard-02",
        "storage": [],
        "network":[]
    }
}
print("POST http://localhost:8001 {}".format(json.dumps(vm, indent=4)))
resp = requests.post("http://localhost:8001", json=vm)
print(resp.text)

print("PATCH http://localhost:8002/VmHost/vmhost-test01/virtualMachine 'vm-srv-wireguard-02'")
resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/virtualMachine", "vm-srv-wireguard-02")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

vhd = {
    "__id": "vhd-wireguard-02-sys",
    "__type": "VirtualHardDisk",
    "__ver": "0.0.1",
    "data": {
        "name": "wireguard-02-sys",
        "path": "c:\compute\vm-srv-wireguard-02\sys.vhd"
    }
}

print("POST http://localhost:8001 {}",format(json.dumps(vhd, indent=4)))
resp = requests.post("http://localhost:8001", json=vhd)
print(resp.text)

storage = {
    "name": "sys",
    "vhd": "vhd-wireguard-02-sys"
}

print("PATCH http://localhost:8001/VirtualMachine/vm-srv-wireguard-02/storage {}".format(json.dumps(storage, indent=4)))
resp = requests.patch("http://localhost:8001/VirtualMachine/vm-srv-wireguard-02/storage", json=storage)
print(json.dumps(resp.json(), indent=4))

print("PATCH http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk 'vhd-wireguard-02-sys'")
resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk", "vhd-wireguard-02-sys")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

vmLink = {
    "__id": "vm-srv-wireguard-02-link-ext",
    "__type": "VmLink",
    "__ver": "0.0.2",
    "data": {
        "vm": "vm-srv-wireguard-02",
        "name": "ext",
        "host": "vmhost-test01",
        "bridge": "br-ext"
    }
}
print("POST http://localhost:8001 {}".format(json.dumps(vmLink, indent=4)))
resp = requests.post("http://localhost:8001", json=vmLink)
print(resp.text)

network = {
    "name": "eth0",
    "link": "vm-srv-wireguard-02-link-ext"
}

print("PATCH http://localhost:8001/VirtualMachine/vm-srv-wireguard-02/network {}".format(json.dumps(network, indent=4)))
resp = requests.patch("http://localhost:8001/VirtualMachine/vm-srv-wireguard-02/network", json=network)
print(json.dumps(resp.json(), indent=4))

print("PATCH http://localhost:8001/NetBridge/br-ext/links 'vm-srv-wireguard-02-link-ext'")
resp = requests.patch("http://localhost:8001/NetBridge/br-ext/links", "vm-srv-wireguard-02-link-ext")
print(json.dumps(resp.json(), indent=4))

print("PATCH http://localhost:8002/VmHost/vmhost-test01/vmLink 'vm-srv-wireguard-02-link-ext'")
resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/vmLink", "vm-srv-wireguard-02-link-ext")
print(json.dumps(resp.json(), indent=4))