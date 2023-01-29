import requests
import json

print("""
######################################################################################################################
# cmtIdx: contentMediaType Auto Index
# This feature is trying to automate the effort of maintain registry array or map to contentMediaType reference
######################################################################################################################
""")
input("press ANY key to continue...")

print("""
######################################################################################################################
# contentMediaType index manual maintenance process
# 1, Create VirtualHost Record
# 2, Create a VirtualHardDrive Record
# 3, Add reference in VirtualHost/virtualHardDisk
# 4, delete VirtualHardDrive Record
# 5, remove reference in VirtualHost/virtualHardDisk
######################################################################################################################
""")
input("press ANY key to continue...")

print("""
######################################################################################################################
# 1, create VirtualHost Record
######################################################################################################################
""")

vmHost = {
    "__id": "vmhost-test01",
    "__type": "VmHost",
    "__ver": "0.0.1",
    "data": {
        "name": "test01",
        "netBridge": [],
        "vmLink": [],
        "virtualMachine":[],
        "nic": [],
        "virtualHardDisk": [],
        "storageDevice": []
    }
}
print("POST http://localhost:8002 {}".format(json.dumps(vmHost, indent=4)))
resp = requests.post("http://localhost:8002", json=vmHost)
print(resp.text)
input("press ANY key to continue...")


print("""
######################################################################################################################
# 2, Create a VirtualHardDrive Record
######################################################################################################################
""")

vhd = {
    "__id": "vhd-vm-test01-sys",
    "__type": "VirtualHardDisk",
    "__ver": "0.0.1",
    "data": {
        "name": "vm-test01-sys",
        "path": "/opt/vhd/vm-test01-sys.vhd",
        "host": "vmhost-test01"
    }
}
print("POST http://localhost:8001 {}".format(json.dumps(vhd, indent=4)))

resp = requests.post("http://localhost:8001", json=vhd)
print(resp.text)
input("press ANY key to continue...")

print("""
######################################################################################################################
# 3, Add reference in VirtualHost/virtualHardDisk
######################################################################################################################
""")

print("PATCH http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk 'vhd-vm-test01-sys'")
resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk", "vhd-vm-test01-sys")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("""
######################################################################################################################
# 4, delete VirtualHardDrive Record
######################################################################################################################
""")

print("DELETE http://localhost:8001/VirtualHardDisk/vhd-vm-test01-sys")
resp = requests.delete("http://localhost:8001/VirtualHardDisk/vhd-vm-test01-sys")
print("delete status: [{}]".format(resp.status_code))

input("press ANY key to continue...")

print("""
######################################################################################################################
# 5, remove reference in VirtualHost/virtualHardDisk
######################################################################################################################
""")

print("PATCH http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk[vhd-vm-test01-sys] ''")

resp = requests.patch("http://localhost:8002/VmHost/vmhost-test01/virtualHardDisk[vhd-vm-test01-sys]", "")
print(json.dumps(resp.json(), indent=4))