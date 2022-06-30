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

package DataHandler

import (
	"fmt"
	"log"
	"net/http"

	"Data"
	"Data/DbConfig"
	"Data/DbIface"
	"InventoryService/InvRecord"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
)

type Handler struct {
	Db DbIface.Database
}

func New(config DbConfig.DatabaseConfig) (*Handler, error) {
	handler := Handler{}
	db, err := Data.ConnectDb(config)
	if err != nil {
		return nil, err
	}
	handler.Db = db
	err = handler.init()
	if err != nil {
		return nil, err
	}
	return &handler, nil
}

func (h *Handler) init() error {
	tbList, err := h.Db.ListTable()
	if err != nil {
		return err
	}
	for _, name := range []string{JsonKey.Schema, Schema.Inventory} {
		tblExists := false
		for _, tbl := range tbList {
			if *tbl == name {
				tblExists = true
			}
		}
		if !tblExists {
			log.Printf("missing table=[%s], create one", name)
			err := h.Db.CreateTable(name, nil)
			if err != nil {
				err = fmt.Errorf("failed to creat table=[%s], Err:%s", name, err)
				return err
			}
		}
	}
	return nil
}

func (h *Handler) List(dataType string) ([]string, int, error) {
	if dataType == JsonKey.Schema || dataType == Schema.Inventory {
		result, code, err := h.ListData(dataType)
		if err != nil {
			return nil, code, err
		}
		dsList := make([]string, 0, len(result))
		for _, data := range result {
			dsList = append(dsList, data[Record.DataId].(string))
		}
		return dsList, http.StatusOK, nil
	}
	_, code, err := h.GetData(JsonKey.Schema, dataType)
	if err != nil {
		return nil, code, err
	}
	dsInfo, code, err := h.GetDataServiceInfo(dataType)
	if err != nil {
		return nil, code, err
	}
	urlPath, err := Util.URLPathJoin(dsInfo.URL, dataType)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to parse url from data service [%s]=[%s], url=[%s], Error:%s", Record.DataId, dsInfo.Id, dsInfo.URL, err)
	}
	dataList, code, err := Util.GetRestData(*urlPath)
	if err != nil {
		return nil, code, fmt.Errorf("failed to get data from REST URL=[%s], Code=[%d], Err:%s", *urlPath, code, err)
	}
	idList := []string{}
	for _, id := range dataList.([]interface{}) {
		idList = append(idList, id.(string))
	}
	return idList, http.StatusOK, nil
}

func (h *Handler) Get(dataType string, dataId string) (interface{}, int, error) {
	if dataType == JsonKey.Schema {
		return h.GetData(JsonKey.Schema, dataId)
	}
	if dataType == Schema.Inventory {
		// retrieve data service record from Inventory
		inv, code, err := h.GetData(Schema.Inventory, dataId)
		if err != nil {
			return nil, code, err
		}
		return inv, code, err
	}
	_, code, err := h.GetData(JsonKey.Schema, dataType)
	if err != nil {
		err = fmt.Errorf("failed to get schema for type [%s], Err:%s", dataType, err)
		return nil, code, err
	}
	dsInfo, code, err := h.GetDataServiceInfo(dataType)
	if err != nil {
		return nil, code, err
	}
	idPath, err := Util.URLPathJoin(dsInfo.URL, dataType, dataId)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to parse url from data service [%s]=[%s], url=[%s], Error:%s", Record.DataId, dsInfo.Id, dsInfo.URL, err)
	}
	data, code, err := Util.GetRestData(*idPath)
	if err != nil {
		if code == http.StatusNotFound {
			return data, code, err
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to get data from REST URL=[%s], Code=[%d], Err:%s", *idPath, code, err)
	}
	result, ok := data.(map[string]interface{})
	if !ok {
		err = fmt.Errorf("data from [%s] is not a validate record data map[string]interface{}", *idPath)
		return nil, http.StatusInternalServerError, err
	}
	return result, http.StatusOK, nil
}

func (h *Handler) ListData(dataType string) ([]map[string]interface{}, int, error) {
	if dataType != JsonKey.Schema && dataType != Schema.Inventory {
		return nil, http.StatusBadRequest, fmt.Errorf("[type]=[%s] is not supported", dataType)
	}
	args := make(map[string]interface{})
	args[DbIface.Table] = dataType
	result, err := h.Db.Get(args)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return result, http.StatusOK, nil
}

func (h *Handler) GetData(dataType string, dataId string) (interface{}, int, error) {
	args := make(map[string]interface{})
	args[DbIface.Table] = dataType
	args[Record.DataId] = dataId
	recordList, err := h.Db.Get(args)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if len(recordList) == 0 {
		return nil, http.StatusNotFound, fmt.Errorf("failed to find [{type}/{id}]=[%s/%s]", dataType, dataId)
	}
	return recordList[0], http.StatusOK, nil
}

func (h *Handler) GetDataServiceInfo(dataType string) (*InvRecord.DataServiceInfo, int, error) {
	invList, code, err := h.ListData(Schema.Inventory)
	if err != nil {
		return nil, code, err
	}
	for _, inv := range invList {
		dsInfo, err := InvRecord.CreateDsInfo(inv)
		if err != nil {
			log.Printf("Failed to load record [%s]=[%s]", Record.DataId, inv[Record.DataId])
			continue
		}
		for _, dsDataType := range dsInfo.TypeList {
			if dsDataType == dataType {
				return dsInfo, http.StatusOK, nil
			}
		}
	}
	return nil, http.StatusNotFound, fmt.Errorf("failed to find Data Service for [type]=[%s]", dataType)
}
