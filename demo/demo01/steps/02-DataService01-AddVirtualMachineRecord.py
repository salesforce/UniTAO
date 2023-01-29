import requests
import json

print("""
#=====================================================================
# TRY to create a new record (w/o VirtualMachine Schema)
#=====================================================================
""")

payloadStr = """
{
    "__id": "vm-srv-wireguard-01",
    "__type": "VirtualMachine",
    "__ver": "0.0.1",
    "data": {
        "name": "srv-wireguard-01",
        "storage": [],
        "network": []
    }
}
"""
print("post(http://localhost:8001)\n{}".format(payloadStr))

payload = json.loads(payloadStr)

resp = requests.post("http://localhost:8001", json=payload)

print(resp.text)

if 200 <= resp.status_code < 300:
    print("record created. Get [http://localhost:8001/VirtualMachine/vm-srv-wireguard-01]")
    resp = requests.get("http://localhost:8001/VirtualMachine/vm-srv-wireguard-01")
    respObj = resp.json()
    print(json.dumps(respObj, indent=4))