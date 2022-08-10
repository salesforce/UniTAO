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
	"net/http"
	"os"

	"InventoryService/Config"
	"InventoryService/DataHandler"
	"InventoryService/InvRecord"
	"InventoryService/ReferalRecord"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
)

type AdminArgs struct {
	cmd    string
	config string
	ops    OpsCmd
}

type OpsCmd struct {
	url string
	id  string
}

const (
	CMD_ADD   = "add"
	CMD_DEL   = "delete"
	CMD_DS    = "ds"
	CMD_DS_ID = "id"
	CMD_SYNC  = "sync"
)

type Admin struct {
	args    *AdminArgs
	config  Config.ServerConfig
	handler *DataHandler.Handler
}

func (a *Admin) argHandler() (*AdminArgs, error) {
	addCmd := flag.NewFlagSet(CMD_ADD, flag.ExitOnError)
	addDbConfig := addCmd.String("config", "", "database connection config")
	addDs := addCmd.String(CMD_DS, "", "data service url to be registered with inventory service")
	addDsId := addCmd.String(CMD_DS_ID, "", "data service unique id within Inventory Service")

	syncCmd := flag.NewFlagSet(CMD_SYNC, flag.ExitOnError)
	syncDbConfig := syncCmd.String("config", "", "database connection config")
	syncDsId := syncCmd.String(CMD_DS_ID, "", "data service unique id to sync data with. if empty, then all ds will be sync-ed")

	delCmd := flag.NewFlagSet(CMD_DEL, flag.ExitOnError)
	delDbConfig := delCmd.String("config", "", "database connection config")
	delDsId := delCmd.String(CMD_DS_ID, "", "data service unique id to be deleted")

	if len(os.Args) < 2 {
		for _, cmd := range []flag.FlagSet{*addCmd, *syncCmd, *delCmd} {
			cmd.Usage()
		}
		return nil, fmt.Errorf("expected [%s, %s, %s]] subcommands", CMD_ADD, CMD_SYNC, CMD_DEL)
	}
	args := AdminArgs{
		cmd: os.Args[1],
	}
	switch args.cmd {
	case CMD_ADD:
		addCmd.Parse(os.Args[2:])
		args.config = *addDbConfig
		args.ops = OpsCmd{
			url: *addDs,
			id:  *addDsId,
		}
		if args.config == "" || args.ops.id == "" || args.ops.url == "" {
			addCmd.Usage()
		}
	case CMD_SYNC:
		syncCmd.Parse(os.Args[2:])
		args.config = *syncDbConfig
		args.ops = OpsCmd{
			id: *syncDsId,
		}
		if args.config == "" {
			syncCmd.Usage()
		}
	case CMD_DEL:
		delCmd.Parse(os.Args[2:])
		args.config = *delDbConfig
		args.ops = OpsCmd{
			id: *delDsId,
		}
		if args.config == "" || args.ops.id == "" {
			delCmd.Usage()
		}
	default:
		for _, cmd := range []flag.FlagSet{*addCmd, *syncCmd, *delCmd} {
			cmd.Usage()
		}
	}
	return &args, nil
}

func (a *Admin) Init() error {
	args, err := a.argHandler()
	if err != nil {
		return err
	}
	a.args = args
	err = Config.Read(a.args.config, &a.config)
	if err != nil {
		return fmt.Errorf("failed to load Inventory Service Configuration,[%s], Error:%s", a.args.config, err)

	}
	handler, err := DataHandler.New(a.config.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize data layer, Err:%s", err)
	}
	a.handler = handler
	return nil
}

func (a *Admin) Run() error {
	switch a.args.cmd {
	case CMD_ADD:
		return a.addDsRecord()
	case CMD_SYNC:
		return a.syncDsSchema()
	case CMD_DEL:
		return a.removeDsRecord()
	}
	return nil
}

func (a *Admin) addDsRecord() error {
	_, code, err := a.handler.GetData(Schema.Inventory, a.args.ops.id)
	if err == nil {
		err = fmt.Errorf("Data Server Record already exists, [%s]=[%s]", Record.DataId, a.args.ops.id)
		return err
	}
	if code != http.StatusNotFound {
		err = fmt.Errorf("failed to query Data Service record, [%s]=[%s], CODE:%d, Error:%s", Record.DataId, a.args.ops.id, code, err)
		return err
	}
	dsInfo := InvRecord.NewDsInfo(a.args.ops.id, a.args.ops.url)
	payload, err := dsInfo.ToIface()
	if err != nil {
		return err
	}
	a.handler.Db.Create(Schema.Inventory, payload)
	return nil
}

func (a *Admin) syncDsSchema() error {
	if a.args.ops.id != "" {
		return a.syncSchemaWithId(a.args.ops.id)
	}
	idList, _, err := a.handler.List(Schema.Inventory)
	if err != nil {
		return fmt.Errorf("failed to list all inventorys. Error: %s", err)
	}
	for _, dsId := range idList {
		err = a.syncSchemaWithId(dsId)
		if err != nil {
			return fmt.Errorf("failed to get schema from DS_Id=[%s], Error:%s", dsId, err)
		}
	}
	return nil
}

func (a *Admin) getCurrentTypeList(dsId string) ([]string, error) {
	typeList, _, err := a.handler.List(DataHandler.Referal)
	if err != nil {
		return nil, err
	}
	currentTypes := []string{}
	for _, dataType := range typeList {
		referal, _, err := a.handler.GetReferal(dataType)
		if err != nil {
			a.removeType(dsId, dataType)
			continue
		}
		if referal.DsId == dsId {
			currentTypes = append(currentTypes, dataType)
		}
	}
	return currentTypes, nil
}

func (a *Admin) getDsTypeList(dsId string) ([]string, error) {
	ds, _, err := a.handler.GetDsInfo(dsId)
	if err != nil {
		return nil, err
	}
	dsUrl, err := ds.GetUrl()
	if err != nil {
		return nil, err
	}
	schemaUrl, err := Util.URLPathJoin(dsUrl, JsonKey.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url from DS record [%s]=[%s], Err:%s", Record.DataId, a.args.ops.id, err)
	}
	result, code, err := Util.GetRestData(*schemaUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to Rest Data from [path]=[%s], Code:%d", *schemaUrl, code)
	}
	newTypeList := []string{}
	for _, dataType := range result.([]interface{}) {
		if dataType != JsonKey.Schema && dataType != Record.KeyRecord {
			newTypeList = append(newTypeList, dataType.(string))
		}
	}
	return newTypeList, nil
}

func (a *Admin) syncSchemaWithId(dsId string) error {
	currentTypes, err := a.getCurrentTypeList(dsId)
	if err != nil {
		return err
	}
	dsTypes, err := a.getDsTypeList(dsId)
	if err != nil {
		return err
	}
	for _, dataType := range currentTypes {
		if !Util.SearchStrList(dsTypes, dataType) {
			a.removeType(dsId, dataType)
		}
	}
	for _, dataType := range dsTypes {
		if !Util.SearchStrList(currentTypes, dataType) {
			a.addType(dsId, dataType)
		}
	}

	// for _, dataType := range ds.TypeList {
	// 	if Util.SearchStrList(newTypeList, dataType) {
	// 		continue
	// 	}
	// 	log.Printf("data type [%s] removed from ds [%s], remove schema", dataType, ds.Id)
	// 	err = a.removeData(JsonKey.Schema, dataType)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to remove schema [%s], Err:%s", dataType, err)
	// 	}
	// }
	// ds.TypeList = newTypeList
	// payload, err := ds.ToIface()
	// if err != nil {
	// 	return fmt.Errorf("failed to convert Data Service info to record, Err:%s", err)
	// }
	// keys := make(map[string]interface{})
	// ds.LastSyncTime = time.Now().Format(time.RFC850)
	// keys[Record.DataId] = ds.Id
	// log.Printf("refresh DataService record for data type mapping")
	// err = a.handler.Db.Replace(Schema.Inventory, keys, payload)
	// if err != nil {
	// 	return fmt.Errorf("failed to replace [%s]=[%s], Err:%s", Record.DataId, ds.Id, err)
	// }
	// for _, dataType := range ds.TypeList {
	// 	_, code, err := a.handler.Get(JsonKey.Schema, dataType)
	// 	if err == nil {
	// 		log.Printf("data type [%s] schema exists, next", dataType)
	// 		continue
	// 	}
	// 	if code != http.StatusNotFound {
	// 		return fmt.Errorf("failed to get schema for [type]=[%s], Err:%s", dataType, err)
	// 	}
	// 	typeSchemaUrl, err := Util.URLPathJoin(*schemaUrl, dataType)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to build schema url. for type=[%s], Error:%s", dataType, err)
	// 	}
	// 	log.Printf("download schema for type=[%s], from url=[%s]", dataType, *typeSchemaUrl)
	// 	typeSchema, code, err := Util.GetRestData(*typeSchemaUrl)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to retrieve data from [url]=[%s], Error:%s", *typeSchemaUrl, err)
	// 	}
	// 	log.Printf("save schema for type=[%s], code=[%d]", dataType, code)
	// 	err = a.handler.Db.Create(JsonKey.Schema, typeSchema)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to create schema [type]=[%s], Err:%s", dataType, err)
	// 	}
	// }
	return nil
}

func (a *Admin) addType(dsId string, dataType string) error {
	dsInfo, _, err := a.handler.GetDsInfo(dsId)
	if err != nil {
		return err
	}
	dsUrl, err := dsInfo.GetUrl()
	if err != nil {
		return err
	}
	schemaUrl := fmt.Sprintf("%s/%s/%s", dsUrl, JsonKey.Schema, dataType)
	schemaRecord, _, err := Util.GetRestData(schemaUrl)
	if err != nil {
		return err
	}
	a.removeData(JsonKey.Schema, dataType)
	a.handler.Db.Create(JsonKey.Schema, schemaRecord)
	referal := ReferalRecord.ReferalRecord{
		Id:        dataType,
		DataType:  dataType,
		SchemaVer: schemaRecord.(map[string]interface{})[Record.Version].(string),
		DsId:      dsId,
	}
	referakData, err := referal.ToMap()
	if err != nil {
		return nil
	}
	a.handler.Db.Create(DataHandler.Referal, referakData)
	return nil
}

func (a *Admin) removeType(dsId string, dataType string) error {
	a.removeData(DataHandler.Referal, dataType)
	a.removeData(JsonKey.Schema, dataType)
	return nil
}

func (a *Admin) removeDsRecord() error {
	err := a.removeData(Schema.Inventory, a.args.ops.id)
	if err != nil {
		return fmt.Errorf("failed to delete Data Service, Err:%s", err)
	}
	return nil
}

func (a *Admin) removeData(dataType string, id string) error {
	keys := make(map[string]interface{})
	keys[Record.DataId] = id
	err := a.handler.Db.Delete(dataType, keys)
	if err != nil {
		return fmt.Errorf("failed to delete Data  [type/%s]=[%s/%s], Err:%s", Record.DataId, dataType, id, err)
	}
	return nil
}

func main() {
	admin := Admin{}
	err := admin.Init()
	if err != nil {
		log.Fatalf("failed to init Inventory Admin, Err:%s", err)
	}
	err = admin.Run()
	if err != nil {
		log.Fatalf("Run Inventory Admin failed.\n Error:%s\n", err)
	}
}
