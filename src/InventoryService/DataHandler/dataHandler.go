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
	"InventoryService/ReferalRecord"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/SchemaPath"
	"github.com/salesforce/UniTAO/lib/Util"
)

const (
	Referal = "referal"
)

type Handler struct {
	Db DbIface.Database
}

func New(config DbConfig.DatabaseConfig) (*Handler, error) {
	db, err := Data.ConnectDb(config)
	if err != nil {
		return nil, err
	}
	handler := Handler{
		Db: db,
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
	for _, name := range []string{JsonKey.Schema, Schema.Inventory, Referal} {
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
	if Util.SearchStrList([]string{JsonKey.Schema, Schema.Inventory, Referal}, dataType) {
		result, code, err := h.ListData(dataType)
		if err != nil {
			return nil, code, err
		}
		dsList := make([]string, 0, len(result))
		dataKey := Record.DataId
		if dataType == Referal {
			dataKey = Record.DataType
		}
		for _, data := range result {
			dsList = append(dsList, data[dataKey].(string))
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
	dsUrl, err := dsInfo.GetUrl()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	urlPath, err := Util.URLPathJoin(dsUrl, dataType)
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

func (h *Handler) Get(dataType string, dataPath string) (interface{}, int, error) {
	_, nextPath := Util.ParsePath(dataPath)
	if dataType == JsonKey.Schema {
		if nextPath != "" {
			return nil, http.StatusBadRequest, fmt.Errorf("path=[%s] not supported on type=[%s]", dataPath, dataType)
		}
		return h.GetData(JsonKey.Schema, dataPath)
	}
	if dataType == Schema.Inventory {
		if nextPath != "" {
			return nil, http.StatusBadRequest, fmt.Errorf("path=[%s] not supported on type=[%s]", dataPath, dataType)
		}
		// retrieve data service record from Inventory
		dsInfo, code, err := h.GetDsInfo(dataPath)
		if err != nil {
			return nil, code, err
		}
		return dsInfo, http.StatusOK, nil
	}
	if dataType == Referal {
		if nextPath != "" {
			return nil, http.StatusBadRequest, fmt.Errorf("path=[%s] not supported on type=[%s]", dataPath, dataType)
		}
		referal, code, err := h.GetReferal(dataPath)
		if err != nil {
			return nil, code, err
		}
		dsInfo, code, err := h.GetDsInfo(referal.DsId)
		if err != nil {
			return nil, code, err
		}
		err = referal.SetDsInfo(dsInfo)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		code, err = referal.GetSchema()
		if err != nil {
			return nil, code, err
		}
		return referal, http.StatusOK, nil
	}
	return h.GetDataByPath(fmt.Sprintf("%s/%s", dataType, dataPath))
}

func (h *Handler) ListData(dataType string) ([]map[string]interface{}, int, error) {
	if !Util.SearchStrList([]string{JsonKey.Schema, Schema.Inventory, Referal}, dataType) {
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

func (h *Handler) GetSchema(dataType string) (*SchemaDoc.SchemaDoc, error) {
	data, _, err := h.GetData(JsonKey.Schema, dataType)
	if err != nil {
		return nil, err
	}
	record, err := Record.LoadMap(data.(map[string]interface{}))
	if err != nil {
		return nil, fmt.Errorf("failed to load schema record data. [type]=[%s], Error: %s", dataType, err)
	}
	return SchemaDoc.New(record.Data, dataType, nil)
}

func (h *Handler) GetDataServiceRecord(dataType string, dataId string) (*Record.Record, error) {
	data, code, err := h.GetDataServiceData(dataType, dataId)
	if err != nil {
		return nil, &SchemaPath.SchemaPathErr{
			Code:    code,
			PathErr: err,
		}
	}
	record, err := Record.LoadMap(data.(map[string]interface{}))
	if err != nil {
		return nil, fmt.Errorf("failed to load data as Record. Error:%s", err)
	}
	return record, nil
}

func (h *Handler) GetDataByPath(dataPath string) (interface{}, int, error) {
	conn := SchemaPath.Connection{
		FuncSchema: h.GetSchema,
		FuncRecord: h.GetDataServiceRecord,
	}
	schemaPath, err := SchemaPath.NewFromPath(&conn, dataPath, nil)
	if err != nil {
		pathErr, ok := err.(*SchemaPath.SchemaPathErr)
		if ok && pathErr.Code == http.StatusNotFound {
			return nil, http.StatusNotFound, pathErr
		}
		return nil, http.StatusBadRequest, fmt.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", dataPath, err)
	}
	result, err := schemaPath.WalkValue()
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("failed to walk SchemaPath. from [path]=[%s], Error: %s", dataPath, err)
	}
	if result == nil {
		return nil, http.StatusNotFound, fmt.Errorf("walk SchemaPath with no value.from [path]=[%s]", dataPath)
	}
	return result, http.StatusOK, nil
}

func (h *Handler) GetDataServiceData(dataType string, dataId string) (interface{}, int, error) {
	dsInfo, code, err := h.GetDataServiceInfo(dataType)
	if err != nil {
		return nil, code, err
	}
	dsUrl, err := dsInfo.GetUrl()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no good url for DataService=[%s], Error: %s", dsInfo.Id, err)
	}
	idPath, err := Util.URLPathJoin(dsUrl, dataType, dataId)
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
	referal, code, err := h.GetReferal(dataType)
	if err != nil {
		return nil, code, err
	}
	dsInfo, code, err := h.GetDsInfo(referal.DsId)
	if err != nil {
		return nil, code, err
	}
	err = referal.SetDsInfo(dsInfo)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return referal.DsInfo, http.StatusOK, nil
}

func (h *Handler) GetReferal(dataType string) (*ReferalRecord.ReferalRecord, int, error) {
	referalData, code, err := h.GetData(Referal, dataType)
	if err != nil {
		return nil, code, fmt.Errorf("failed to get referal record for [type]=[%s]", dataType)
	}
	referal, err := ReferalRecord.LoadMap(referalData.(map[string]interface{}))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return referal, http.StatusOK, nil
}

func (h *Handler) GetDsInfo(dsId string) (*InvRecord.DataServiceInfo, int, error) {
	dsInfoData, code, err := h.GetData(Schema.Inventory, dsId)
	if err != nil {
		return nil, code, err
	}
	dsInfo, err := InvRecord.CreateDsInfo(dsInfoData)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return dsInfo, http.StatusOK, nil
}
