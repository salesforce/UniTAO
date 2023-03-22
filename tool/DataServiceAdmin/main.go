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
	"flag"
	"fmt"
	"log"
	"os"

	"Data"
	"Data/DbIface"
	"DataService/Config"

	"github.com/salesforce/UniTAO/lib/Util/CustomLogger"
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

type AdminArgs struct {
	cmd       string
	config    string
	srvConfig Config.Confuguration
	table     TableArgs
	data      DataArgs
	logPath   string
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
	tableLogPath := tableCmd.String("log", "", "path that hold log")

	dataCmd := flag.NewFlagSet(DATA, flag.ExitOnError)
	dataDbConfig := dataCmd.String("config", "", "database connection config")
	dataTable := dataCmd.String(TABLE, "", "data table to import")
	dataFile := dataCmd.String(DATA, "", "data file to be import into database")
	dataLogPath := dataCmd.String("log", "", "path that hold log")

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
		args.logPath = *tableLogPath
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
		args.logPath = *dataLogPath
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

func CreateTables(db DbIface.Database, args AdminArgs, logger *log.Logger) {
	logger.Printf("create table from %s", args.table.meta)
	tableMeta, err := Json.LoadJSONMap(args.table.meta)
	if err != nil {
		logger.Fatalf("failed to load database metadata file [%s]", args.table.meta)
	}
	tableList, err := db.ListTable()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	logger.Printf("current table length, %d", len(tableList))
	configMeta := make(map[string]interface{})
	configTables := args.srvConfig.DataTable.Map()
	logger.Printf("create translated table structure")

	for key, meta := range tableMeta {
		logger.Printf("check for table %s", key)
		if tableName, ok := configTables[key].(string); ok {
			log.Printf("custom table name [%s]=>[%s]", key, tableName)
			configMeta[tableName] = meta
			continue
		}
		logger.Printf("keep table as the same name: [%s]", key)
		configMeta[key] = meta
	}
	log.Print("determine if we should remove existing table")
	for _, table := range tableList {
		logger.Printf("Match table [%s] with expected meta", table)
		tableName := table.(string)
		_, exists := configMeta[tableName]
		if exists {
			logger.Printf("table [%s] exists", table)
			if args.table.reset {
				logger.Printf("remove table [%s]", table)
				db.DeleteTable(tableName)
			} else {
				logger.Printf("remove talbe [%s] from create list", table)
				delete(tableMeta, tableName)
			}
			continue
		}
		logger.Printf("ignore unknown table [%s]", table)
	}
	if len(configMeta) == 0 {
		logger.Print("there is no table to create")
		return
	}
	logger.Printf("create %d tables", len(configMeta))
	for table, meta := range configMeta {
		logger.Printf("create table [%s]", table)
		err := db.CreateTable(table, meta.(map[string]interface{}))
		if err != nil {
			logger.Fatalf("failed to create table %s, Err: %s", table, err)
		}
	}
}

func ImportData(db DbIface.Database, args AdminArgs, logger *log.Logger) {
	tableList, err := db.ListTable()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	if args.data.table == "" {
		ImportTables(db, tableList, args, logger)
		return
	}
	logger.Printf("load record list for table [%s]", args.data.table)
	data, err := Json.LoadJSONList(args.data.file)
	if err != nil {
		logger.Fatalf("failed to load database metadata file [%s]", args.data.file)
	}
	for _, table := range tableList {
		tableName := table.(string)
		if tableName == args.data.table {
			logger.Printf("table [%s] exists, import %d records from file %s", table, len(data), args.data.file)
			for idx, record := range data {
				err := db.Create(tableName, record)
				if err != nil {
					logger.Fatalf("falied to create data @%d for table %s, Err: %s", idx, table, err)
				}
			}
			logger.Print("data loaded")
			return
		}
	}
	logger.Fatalf("table [%s] does not exists in database", args.data.table)
}

func ImportTables(db DbIface.Database, tableList []interface{}, args AdminArgs, logger *log.Logger) {
	tableMap := args.srvConfig.DataTable.Map()
	logger.Printf("load table meta from %s", args.data.file)
	data, err := Json.LoadJSONMap(args.data.file)
	if err != nil {
		logger.Fatalf("failed to load database metadata file [%s]", args.data.file)
	}
	impData := make(map[string]interface{})
	for key, tData := range data {
		if _, ok := tableMap[key]; ok {
			newKey := tableMap[key].(string)
			logger.Printf("create table[%s] as [%s]", key, newKey)
			impData[newKey] = tData
		} else {
			logger.Printf("create table[%s]", key)
			impData[key] = tData
		}
	}

	for _, table := range tableList {
		logger.Printf("Match table [%s] with table data", table)
		tableName := table.(string)
		tableData, exists := impData[tableName].([]interface{})
		if !exists {
			logger.Printf("table [%s], no data import", table)
			continue
		}
		for idx, record := range tableData {
			err := db.Create(tableName, record)
			if err != nil {
				logger.Fatalf("falied to create data @%d for table %s, Err: %s", idx, table, err)
			}
		}
	}
}

func main() {
	args := ArgHandler()
	log.Print("Admin tool for Data Service")
	config := Config.Confuguration{}
	err := Config.Read(args.config, &config)
	if err != nil {
		log.Fatalf("failed to read configuration file=[%s], Err:%s", args.config, err)
	}
	args.srvConfig = config
	logFile, logger, ex := CustomLogger.FileLoger(args.logPath, fmt.Sprintf("%s_admin", config.Http.Id))
	if ex != nil {
		log.Fatalf("failed to create log file @[%s]", args.logPath)
	}
	defer logFile.Close()
	database, err := Data.ConnectDb(config.Database, logger)
	if err != nil {
		logger.Fatalf("failed to connect to database, err:%s", err)
	}
	logger.Println("database connected")
	switch args.cmd {
	case "table":
		CreateTables(database, args, logger)
	case "data":
		ImportData(database, args, logger)
	}
	logger.Printf("Admin Operation %s completed", args.cmd)
}
