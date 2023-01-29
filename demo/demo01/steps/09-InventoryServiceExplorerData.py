import requests
import json

print("""
########################################################################################################
# Inventory Service can 
#  - explore all data that it understand by proxy your request to DataService
#  - follow data schema and contentMediaType link to walk through data
#  - using * and iterator to list all possible paths to help filter data
########################################################################################################
""")

input("press ANY key to continue...")

print("Get Data virtual machine from Data Service 01 and Inventory Service")

print("get http://localhost:8001/VirtualMachine")

resp = requests.get("http://localhost:8001/VirtualMachine")
print(json.dumps(resp.json(), indent=4))

print("get http://localhost:8001/VirtualMachine/vm-srv-wireguard-01")

resp = requests.get("http://localhost:8001/VirtualMachine/vm-srv-wireguard-01")
print(json.dumps(resp.json(), indent=4))

print("get http://localhost:8004/VirtualMachine/vm-srv-wireguard-01")

resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("""
####################################################################################################
# from VirtualMachine[vm-srv-wireguard-01] going through network
# and visit the second VirturlMachine on the same network [br-ext]
# for every step of the explore, we can use the path command ?schema to display schema data
####################################################################################################
""")

input("press ANY key to continue...")

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01")

resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01")
print(json.dumps(resp.json(), indent=4))

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network")
print(json.dumps(resp.json(), indent=4))

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]")
print(json.dumps(resp.json(), indent=4))

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]/vm")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]/vm")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]/vm?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]/vm?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("""
###########################################################################################################################
# the data display is very messy when trieve the whole complex of data every step
# so we create a new flag command ?flat, so we only display a flat data on the layer.
# this command make it easier to see the option of each step
# so we can try the above steps again.
###########################################################################################################################
""")

input("press ANY key to continue...")

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01?flat")

resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01?flat")
print(json.dumps(resp.json(), indent=4))

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network?flat")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network?flat")
print(json.dumps(resp.json(), indent=4))

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]?flat")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]?flat")
print(json.dumps(resp.json(), indent=4))

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link?flat")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link?flat")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge?flat")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge?flat")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links?flat")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links?flat")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]?flat")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]?flat")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]?schema")
print(json.dumps(resp.json(), indent=4))
input("press ANY key to continue...")

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]/vm?flat")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]/vm?flat")
print(json.dumps(resp.json(), indent=4))

print("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]/vm?schema")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[eth0]/link/bridge/links[vm-srv-wireguard-02-link-ext]/vm?schema")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("""
#########################################################################################################################################
# for the array/map options during the data explore
# we can use * to replace the index string so we can walk deeper with all options
# so we perform the filter later with leaf data collection
# with ?iterator command, we can also list all the option on each leaf for easier filter
#########################################################################################################################################
""")

input("press ANY key to continue...")

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]")
print(json.dumps(resp.json(), indent=4))

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]?iterator")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]?iterator")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]/link")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]/link")
print(json.dumps(resp.json(), indent=4))

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]/link?iterator")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]/link?iterator")
print(json.dumps(resp.json(), indent=4))

input("press ANY key to continue...")

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]/link/bridge/links[*]/vm")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]/link/bridge/links[*]/vm")
print(json.dumps(resp.json(), indent=4))

print("GET http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]/link/bridge/links[*]/vm?iterator")
resp = requests.get("http://localhost:8004/VirtualMachine/vm-srv-wireguard-01/network[*]/link/bridge/links[*]/vm?iterator")
print(json.dumps(resp.json(), indent=4))