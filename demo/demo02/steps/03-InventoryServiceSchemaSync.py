import requests
import json
import os
import time

print("""
#####################################################################################
# Sync Schema between Data Service and Inventory Service
# so that inventory service can understand where to find data accordingly
#####################################################################################
""")

print("get [DataService01] http://localhost:8001/schema")

resp = requests.get("http://localhost:8001/schema")

print(json.dumps(resp.json(), indent=4))

print("get [DataService02] http://localhost:8002/schema")

resp = requests.get("http://localhost:8002/schema")

print(json.dumps(resp.json(), indent=4))

print("get [InventoryService] http://localhost:8004/schema")

resp = requests.get("http://localhost:8004/schema")

print(json.dumps(resp.json(), indent=4))

os.system("docker start unitao-inv-service-admin")

print("sleep for 5 seconds")
time.sleep(5)

print("get [InventoryService] http://localhost:8004/schema")

resp = requests.get("http://localhost:8004/schema")

print(json.dumps(resp.json(), indent=4))
