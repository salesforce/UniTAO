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
    print("Submit data loaded")
    for ds_name in ds_map:
        ds_url = ds_map[ds_name]["url"]
        print("Data Service[{}][{}]".format(ds_name, ds_url))
        if ds_name in ds_map:
            print("submit {} data to {}[{}]".format(len(data[ds_name]), ds_name, ds_url))
            submit_data(ds_url, data[ds_name])
        else:
            print("No more data to submit for {}, {}".format(ds_name, json.dumps(data, indent=4)))


def submit_data(ds_url, data_list):
    data_to_add = data_list
    while True:
        failed_list = []
        for data in data_to_add:
            if not create_record(ds_url, data):
                failed_list.append(data)
        if len(failed_list) == 0:
            print("no more data to submit for [{}]".format(ds_url))
            break
        print("submit {} more records, sleep 1 second".format(len(failed_list)))
        time.sleep(1)
        data_to_add = failed_list


def create_record(ds_url, record):
    record_url = "{}/{}/{}".format(ds_url, record["__type"], record["__id"])
    res = requests.get(record_url)
    if requests.codes.ok <= res.status_code < 300:
        print("data [{}/{}] already exists.".format(record["__type"], record["__id"]))
        return True
    if res.status_code != requests.codes.not_found:
        print("failed to get data [{}/{}] from [{}]".format(record["__type"], record["__id"], ds_url))
        return False
    res = requests.post(ds_url, json=record)
    if requests.codes.ok <= res.status_code < 300:
        print("data [{}/{}] created.".format(record["__type"], record["__id"]))
        return True


if __name__ == "__main__":
    print("submit data")
    main()

