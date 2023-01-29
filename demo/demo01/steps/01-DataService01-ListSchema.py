print("""
####################################################################
# list all schema in dataservice 01 [http://localhost:8001/schema]
####################################################################
""")

import requests
import json

resp = requests.get("http://localhost:8001")

respObj = resp.json()

print(json.dumps(respObj, indent=4))
