#!/usr/bin/env python3
import argparse
import json
import os
import time

import requests


def parse_arguments():
    parser = argparse.ArgumentParser()
    parser.add_argument("-m", "--map", type=str, help="data service map file path")
    parser.add_argument("-d", "--data", type=str, help="data file path")
    args = parser.parse_args()
    return args


def main():
    args = parse_arguments()
    print("load Data Service Map file")
    if args.map is None or not os.path.exists(args.map):
        print("Data Service Map file does not exists, [{}]".format(args.map))
        exit(-1)
    try:
        with open(args.map) as fp:
            ds_map = json.load(fp)
    except Exception as e:
        print("failed loading DataService Map file as json. {} ".format(e))
        exit(-2)
    print("DS Map file loaded")
    print("Load Data to submit")
    if args.data is None or not os.path.exists(args.data):
        print("Data File does not exists. [{}]".format(args.data))
    try:
        with open(args.data) as fp:
            data = json.load(fp)
    except Exception as e:
        print("failed loading Submit Data file as json. {} ".format(e))
        exit(-2)

    data_list = []
    for ds_name in data:
        print("check url for data service[{}]".format(ds_name))
        if ds_name not in ds_map:
            print("DataService[{}] does not exists,skip".format(ds_name))
            continue
        ds_url = ds_map[ds_name]["url"]
        print("found url[{}]".format(ds_url))
        for item in data[ds_name]:
            dataItem = {
                "dsName": ds_name,
                "url": ds_url,
                "data": item
            }
            data_list.append(dataItem)
    if len(data_list) == 0:
        print("no record to submit")
        return
    while True:
        failed_list = submit_data(data_list)
        if len(failed_list) ==0:
            print("all record submitted")
            return
        print("[{}] record failed to create, sleep 1 second, try again".format(len(failed_list)))
        data_list = failed_list
        time.sleep(1)

def submit_data(data_list):
    failed_list = []    
    for data in data_list:
        if not create_record(data["url"], data["data"]):
            failed_list.append(data)
    return failed_list


def create_record(ds_url, record):
    print("post record to [{}]".format(ds_url))
    record_url = "{}/{}/{}".format(ds_url, record["__type"], record["__id"])
    res = requests.get(record_url)
    if requests.codes.ok <= res.status_code < 300:
        print("data [{}/{}] already exists.".format(record["__type"], record["__id"]))
        return True
    if res.status_code != requests.codes.not_found:
        print("failed to get data [{}/{}] from [{}]".format(record["__type"], record["__id"], ds_url))
        return False
    print("post {} {}".format(ds_url, json.dumps(record, indent=4)))
    res = requests.post(ds_url, json=record)
    if requests.codes.ok <= res.status_code < 300:
        print("data [{}/{}] created.".format(record["__type"], record["__id"]))
        return True


if __name__ == "__main__":
    print("submit data")
    main()

