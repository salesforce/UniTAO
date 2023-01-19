import requests
import json

print("""
#################################################################################3
#   The process of Add a NetBridge and link it to VMHost record
#   1, add NetBridge record
#   2, patch VMHost Record to point to the new NetBridge
#################################################################################3
""")

input("press Enter to continue")

bridgeStr = """
{
    "__id": "br-ext",
    "__type": "NetBridge",
    "__ver": "0.0.1",
    "data": {
        "name": "ext",
        "links": []
    }
}
"""
print("1, post to [http://localhost:8001]\n{}".format(bridgeStr))

bridgeRecord = json.loads(bridgeStr)

resp = requests.post("http://localhost:8001", json=bridgeRecord)

print(resp.text)

linkStr = """
{
    "__id": "br-ext-link-vm-srv-wireguard-01",
    "__type": "NetLink",
    "__ver": "0.0.1",
    "data": {
        "bridge": "br-ext",
        "vm": "vm-srv-wireguard-01"
    }
}
"""