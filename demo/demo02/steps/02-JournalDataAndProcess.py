import requests
import json

print("""
####################################################################################################################################
# Journal Process:
# In order to react to data changes, we are creating Journal DataType to record all data changes.
# Journal log is created whenever there is a data changes. (POST, PUT, PATCH, DELETE)
# JournalPrcess will handle each journal log entry
####################################################################################################################################
""")
input("press ANY key to continue...")

print("""
Journal Record Schema:

Get http://localhost:8001/schema/journal
""")

input("press ANY key to continue...")

resp = requests.get("http://localhost:8001/schema/journal")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("""
for each data record, we keep a series of journal log records.
the id of the record show in the schema template:
dataType:{dataType}_dataId:{dataId}_page:{idx}

Get list of journal of previous schema upload

Get http://localhost:8001/journal
""")
input("press ANY key to continue...")

resp = requests.get("http://localhost:8001/journal")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("""
For each record, a batch of journal log stored within.
the log entry which is not finished, stay in the active list
the log entry that all journal process finished, will be put under archived list

GET http://localhost:8001/journal/dataType:schema_dataId:VirtualHardDisk_page:1
""")

input("press ANY key to continue...")

resp = requests.get("http://localhost:8001/journal/dataType:schema_dataId:VirtualHardDisk_page:1")
print(json.dumps(resp.json(), indent=4))