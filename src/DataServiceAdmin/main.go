/*
************************************************************************************************************
Copyright (c) 2022 Salesforce, Inc.
All rights reserved.

UniTAO was originally created in 2022 by Shai Herzog & Yi Huo as an
Universal No-Coding Heterogeneous Infrastructure Maintenance & Inventory system that is holistically driven by open/community-developed semantic models/schemas.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>

This copyright notice and license applies to all files in this directory or sub-directories, except when stated otherwise explicitly.
************************************************************************************************************
*/

package main

import (
	"UniTao/Data"
	"UniTao/Data/DbIface"
	"UniTao/DataService/lib/Config"
	"flag"
	"log"
	"os"

	"github.com/salesforce/UniTAO/lib/Util"
)

type AdminArgs struct {
	cmd       string
	config    string
	srvConfig Config.Confuguration
	table     TableArgs
	data      DataArgs
}

type TableArgs struct {
	meta  string
	reset bool
}

type DataArgs struct {
	table string
	file  string
}

const (
	TABLE = "table"
	DATA  = "data"
)

func ArgHandler() AdminArgs {

	tableCmd := flag.NewFlagSet(TABLE, flag.ExitOnError)
	tableDbConfig := tableCmd.String("config", "", "database connection config")
	tableMeta := tableCmd.String(TABLE, "", "metadata file describe tables to be create")
	tableReset := tableCmd.Bool("reset", false, "whether we want to reset data in table or not")

	dataCmd := flag.NewFlagSet(DATA, flag.ExitOnError)
	dataDbConfig := dataCmd.String("config", "", "database connection config")
	dataTable := dataCmd.String(TABLE, "", "data table to import")
	dataFile := dataCmd.String(DATA, "", "data file to be import into database")

	if len(os.Args) < 2 {
		log.Fatal("expected [table, data]] subcommands")
	}
	args := AdminArgs{
		cmd: os.Args[1],
	}
	switch args.cmd {
	case TABLE:
		tableCmd.Parse(os.Args[2:])
		args.config = *tableDbConfig
		args.table.meta = *tableMeta
		args.table.reset = *tableReset
		if args.config == "" {
			tableCmd.Usage()
			log.Fatalf("missing configuration for %s", TABLE)
		}
		if args.table.meta == "" {
			tableCmd.Usage()
			log.Fatalf("missing meta for %s", TABLE)
		}
	case DATA:
		dataCmd.Parse(os.Args[2:])
		args.config = *dataDbConfig
		args.data.table = *dataTable
		args.data.file = *dataFile
		if args.config == "" {
			dataCmd.Usage()
			log.Fatalf("missing configuration for %s", DATA)
		}
		if args.data.file == "" {
			tableCmd.Usage()
			log.Fatalf("missing data file for %s", DATA)
		}
	default:
		log.Fatalf("Unknown cmd=%s", args.cmd)
	}
	return args
}

func CreateTables(db DbIface.Database, args AdminArgs) {
	tableMeta, err := Util.LoadJSONMap(args.table.meta)
	if err != nil {
		log.Fatalf("failed to load database metadata file [%s]", args.table.meta)
	}
	tableList, err := db.ListTable()
	if err != nil {
		log.Fatalf(err.Error())
	}
	configMeta := make(map[string]interface{})
	configTables := args.srvConfig.DataTable.Map()
	for key, table := range configTables {
		if _, ok := tableMeta[key]; ok {
			configMeta[table.(string)] = tableMeta[key]
		}
	}
	for _, table := range tableList {
		log.Printf("Match table [%s] with expected meta", *table)
		_, exists := configMeta[*table]
		if exists {
			log.Printf("table [%s] exists", *table)
			if args.table.reset {
				log.Printf("remove table [%s]", *table)
				db.DeleteTable(*table)
			} else {
				log.Printf("remove talbe [%s] from create list", *table)
				delete(tableMeta, *table)
			}
			continue
		}
		log.Printf("ignore unknown table [%s]", *table)
	}
	if len(configMeta) == 0 {
		log.Print("there is no table to create")
		return
	}
	for table, meta := range configMeta {
		log.Printf("create table [%s]", table)
		err := db.CreateTable(table, meta.(map[string]interface{}))
		if err != nil {
			log.Fatalf("failed to create table %s, Err: %s", table, err)
		}
	}
}

func ImportData(db DbIface.Database, args AdminArgs) {
	tableList, err := db.ListTable()
	if err != nil {
		log.Fatalf(err.Error())
	}
	if args.data.table == "" {
		ImportTables(db, tableList, args)
		return
	}
	log.Printf("load record list for table [%s]", args.data.table)
	data, err := Util.LoadJSONList(args.data.file)
	if err != nil {
		log.Fatalf("failed to load database metadata file [%s]", args.data.file)
	}
	for _, table := range tableList {
		if *table == args.data.table {
			log.Printf("table [%s] exists, import %d records from file %s", *table, len(data), args.data.file)
			for idx, record := range data {
				err := db.Create(*table, record)
				if err != nil {
					log.Fatalf("falied to create data @%d for table %s, Err: %s", idx, *table, err)
				}
			}
			log.Print("data loaded")
			return
		}
	}
	log.Fatalf("table [%s] does not exists in database", args.data.table)
}

func ImportTables(db DbIface.Database, tableList []*string, args AdminArgs) {
	tableMap := args.srvConfig.DataTable.Map()
	log.Print("no table specified, load multi-table data file")
	data, err := Util.LoadJSONMap(args.data.file)
	if err != nil {
		log.Fatalf("failed to load database metadata file [%s]", args.data.file)
	}
	impData := make(map[string]interface{})
	for key, tData := range data {
		if _, ok := tableMap[key]; ok {
			impData[tableMap[key].(string)] = tData
		} else {
			impData[key] = tData
		}
	}

	for _, table := range tableList {
		log.Printf("Match table [%s] with table data", *table)
		tableData, exists := impData[*table].([]interface{})
		if !exists {
			log.Printf("table [%s], no data import", *table)
			continue
		}
		for idx, record := range tableData {
			err := db.Create(*table, record)
			if err != nil {
				log.Fatalf("falied to create data @%d for table %s, Err: %s", idx, *table, err)
			}
		}
	}
}

func main() {
	log.Print("Admin tool for Data Service")
	args := ArgHandler()
	config := Config.Confuguration{}
	err := Config.Read(args.config, &config)
	if err != nil {
		log.Fatalf("failed to read configuration file=[%s], Err:%s", args.config, err)
	}
	args.srvConfig = config
	database, err := Data.ConnectDb(config.Database)
	if err != nil {
		log.Fatalf("failed to connect to database, err:%s", err)
	}
	switch args.cmd {
	case "table":
		CreateTables(database, args)
	case "data":
		ImportData(database, args)
	}
	log.Printf("Admin Operation %s completed", args.cmd)
}
