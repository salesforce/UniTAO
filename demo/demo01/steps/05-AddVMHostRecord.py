import requests
import json

print("""
########################################################################################
# Add new record of VMHost: vmhost-kvm001
# in the schema, we use conentMediaType for relationship
# between string value and referred data id
# contentMediaType format "inventory/{link data type}"
########################################################################################
""")

input("press Enter to continue")

print("Get schema for VMHost[http://localhost:8001/schema/VMHost]")
resp = requests.get("http://localhost:8001/schema/VMHost")
print(json.dumps(resp.json(), indent=4))

input("press Enter to continue")

VMHostStrBad = """
{
    "__id": "vmhost-kvm001",
    "__type": "VMHost",
    "__ver": "0.0.1",
    "data": {
        "name": "kvm001",
        "bridges": [],
        "links": [],
        "vms": [
            "vm-srv-wireguard-02"
        ]
    }
}
"""
print("post to [http://localhost:8001] a bad reference record [vm-srv-wireguard-02] should be [vm-srv-wireguard-01]:\n{}".format(VMHostStrBad))

payload = json.loads(VMHostStrBad)

resp = requests.post("http://localhost:8001", json=payload)

print(resp.text)

input("press Enter to continue")

VMHostStr = """
{
    "__id": "vmhost-kvm001",
    "__type": "VMHost",
    "__ver": "0.0.1",
    "data": {
        "name": "kvm001",
        "bridges": [],
        "links": [],
        "vms": [
            "vm-srv-wireguard-01"
        ]
    }
}
"""

print("post to [http://localhost:8001] the good ref record = [vm-srv-wireguard-01]:\n{}".format(VMHostStr))

payload = json.loads(VMHostStr)

resp = requests.post("http://localhost:8001", json=payload)

print(resp.text)

