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
	"InventoryService/RefRecord"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
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
	LatestVer = "0.0.1"
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
	_, err := a.handler.GetData(Schema.Inventory, a.args.ops.id)
	if err == nil {
		return fmt.Errorf("data server record already exists, [%s]=[%s]", Record.DataId, a.args.ops.id)
	}
	if err.Status != http.StatusNotFound {
		return fmt.Errorf("failed to query Data Service record, [%s]=[%s], Status:%d, Error:%s", Record.DataId, a.args.ops.id, err.Status, err)
	}
	dsRecord := InvRecord.NewDsInfo(a.args.ops.id, a.args.ops.url)
	payload, e := Util.StructToMap(dsRecord)
	if e != nil {
		return e
	}
	a.handler.Db.Create(Schema.Inventory, payload)
	return nil
}

func (a *Admin) syncDsSchema() error {
	if a.args.ops.id != "" {
		return a.syncSchemaWithId(a.args.ops.id)
	}
	idList, err := a.handler.List(Schema.Inventory)
	if err != nil {
		return fmt.Errorf("failed to list all inventorys. Error: %s", err)
	}
	for _, dsId := range idList {
		e := a.syncSchemaWithId(dsId.(string))
		if e != nil {
			return fmt.Errorf("failed to get schema from DS_Id=[%s], Error:%s", dsId, e)
		}
	}
	return nil
}

func (a *Admin) getReferralTypes(dsId string) (map[string]bool, error) {
	typeList, err := a.handler.List(RefRecord.Referral)
	if err != nil {
		return nil, err
	}
	refTypes := map[string]bool{}
	for _, dataType := range typeList {
		referral, err := a.handler.GetReferral(dataType.(string))
		if err != nil {
			a.removeType(dsId, dataType.(string))
			continue
		}
		if referral.DsId == dsId {
			refTypes[dataType.(string)] = true
		}
	}
	return refTypes, nil
}

func (a *Admin) getDsTypeHash(dsId string) (map[string]bool, error) {
	ds, err := a.handler.GetDsInfo(dsId)
	if err != nil {
		return nil, err
	}
	dsUrl, e := ds.GetUrl()
	if e != nil {
		return nil, e
	}
	schemaUrl, e := Http.URLPathJoin(dsUrl, JsonKey.Schema)
	if e != nil {
		return nil, fmt.Errorf("failed to parse url from DS record [%s]=[%s], Err:%s", Record.DataId, a.args.ops.id, err)
	}
	result, code, e := Http.GetRestData(*schemaUrl)
	if e != nil {
		return nil, fmt.Errorf("failed to Rest Data from [path]=[%s], Code:%d", *schemaUrl, code)
	}
	typeHash := map[string]bool{}
	for _, dataType := range result.([]interface{}) {
		if dataType != JsonKey.Schema && dataType != Record.KeyRecord {
			typeHash[dataType.(string)] = true
		}
	}
	return typeHash, nil
}

func (a *Admin) syncSchemaWithId(dsId string) error {
	refTypes, err := a.getReferralTypes(dsId)
	if err != nil {
		return err
	}
	dsTypes, err := a.getDsTypeHash(dsId)
	if err != nil {
		return err
	}
	for dataType := range refTypes {
		if _, ok := dsTypes[dataType]; !ok {
			a.removeType(dsId, dataType)
		}
	}
	for dataType := range dsTypes {
		if _, ok := refTypes[dataType]; !ok {
			a.addType(dsId, dataType)
		}
	}
	return nil
}

func (a *Admin) addType(dsId string, dataType string) error {
	dsInfo, err := a.handler.GetDsInfo(dsId)
	if err != nil {
		return err
	}
	dsUrl, e := dsInfo.GetUrl()
	if e != nil {
		return e
	}
	schemaUrl := fmt.Sprintf("%s/%s/%s", dsUrl, JsonKey.Schema, dataType)
	schemaRecord, _, e := Http.GetRestData(schemaUrl)
	if e != nil {
		return e
	}
	a.removeData(JsonKey.Schema, dataType)
	a.handler.Db.Create(JsonKey.Schema, schemaRecord)
	referral := RefRecord.ReferralData{
		DataType:  dataType,
		SchemaVer: schemaRecord.(map[string]interface{})[Record.Version].(string),
		DsId:      dsId,
	}
	referralData, _ := Util.StructToMap(referral.GetRecord())
	a.handler.Db.Create(RefRecord.Referral, referralData)
	return nil
}

func (a *Admin) removeType(dsId string, dataType string) error {
	a.removeData(RefRecord.Referral, dataType)
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
