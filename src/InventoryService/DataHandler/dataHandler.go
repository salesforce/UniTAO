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
	"InventoryService/RefRecord"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/SchemaPath"
	SchemaPathData "github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type Handler struct {
	log *log.Logger
	Db  DbIface.Database
}

var InvTypes = map[string]bool{
	Schema.Inventory:    true,
	RefRecord.Referral:  true,
	SchemaPath.PathName: true,
}

var EditableTypes = map[string]bool{
	SchemaPath.PathName: true,
}

func New(config DbConfig.DatabaseConfig, logger *log.Logger) (*Handler, error) {
	if logger == nil {
		logger = log.Default()
	}
	db, err := Data.ConnectDb(config, logger)
	if err != nil {
		return nil, err
	}
	handler := Handler{
		log: logger,
		Db:  db,
	}
	handler.Db = db
	err = handler.init()
	if err != nil {
		return nil, err
	}
	return &handler, nil
}

func (h *Handler) Log(msg string) {
	h.log.Printf(fmt.Sprintf("InvSrvHandler: %s", msg))
}

func (h *Handler) init() error {
	tbList, err := h.Db.ListTable()
	if err != nil {
		return err
	}
	for _, name := range []string{Schema.Inventory, RefRecord.Referral} {
		tblExists := false
		for _, tbl := range tbList {
			if tbl == name {
				tblExists = true
			}
		}
		if !tblExists {
			h.Log(fmt.Sprintf("missing table=[%s], create one", name))
			err := h.Db.CreateTable(name, nil)
			if err != nil {
				err = fmt.Errorf("failed to creat table=[%s], Err:%s", name, err)
				return err
			}
		}
	}
	return nil
}

func (h *Handler) List(dataType string) ([]interface{}, *Http.HttpError) {
	if _, ok := InvTypes[dataType]; ok {
		result, err := h.ListData(dataType)
		if err != nil {
			return nil, err
		}
		dsList := make([]interface{}, 0, len(result))
		dataKey := Record.DataId
		for _, data := range result {
			dsList = append(dsList, data[dataKey].(string))
		}
		return dsList, nil
	}
	if dataType == JsonKey.Schema {
		// we don't want to store Schema data locally in Inventory.
		// but list of dataType supported is the list of referral
		refList, err := h.List(RefRecord.Referral)
		if err != nil {
			return nil, err
		}
		// Add internal Inventory type to data type list
		for key := range InvTypes {
			refList = append(refList, key)
		}
		return refList, nil
	}
	referral, err := h.GetReferral(dataType)
	if err != nil {
		return nil, err
	}
	dsUrl, e := referral.DsInfo.GetUrl()
	if e != nil {
		return nil, Http.NewHttpError(e.Error(), http.StatusInternalServerError)
	}
	urlPath, e := Http.URLPathJoin(dsUrl, dataType)
	if e != nil {
		return nil, Http.WrapError(e, fmt.Sprintf("failed to parse url from data service [%s]=[%s], url=[%s]", Record.DataId, referral.DsInfo.Id, dsUrl), http.StatusInternalServerError)
	}
	data, code, e := Http.GetRestData(*urlPath)
	if e != nil {
		return nil, Http.WrapError(e, fmt.Sprintf("failed to get data from REST URL=[%s]", *urlPath), code)

	}
	return data.([]interface{}), nil
}

func (h *Handler) Get(dataType string, dataPath string) (interface{}, *Http.HttpError) {
	// if we get to this function, it means dataPath is not empty string already
	dataId, nextPath := Util.ParsePath(dataPath)
	if _, ok := InvTypes[dataType]; ok || dataType == JsonKey.Schema {
		if nextPath != "" {
			return nil, Http.NewHttpError(fmt.Sprintf("path=[%s] not supported on type=[%s]", dataPath, dataType), http.StatusBadRequest)
		}
		return h.GetRecord(dataType, dataId)
	}
	return h.GetDataByPath(dataType, dataId, nextPath)
}

func (h *Handler) GetSchemaRecord(dataType string) (*Record.Record, *Http.HttpError) {
	if _, ok := InvTypes[dataType]; ok {
		var record *Record.Record
		switch dataType {
		case RefRecord.Referral:
			record, _ = Record.LoadStr(RefRecord.SchemaRecord)
		case Schema.Inventory:
			record, _ = Record.LoadStr(InvRecord.SchemaRecord)
		case SchemaPath.PathName:
			record, _ = Record.LoadStr(SchemaPath.PathDataSchema)
		default:
			return nil, Http.NewHttpError(fmt.Sprintf("don't know how to get schema for Inventory Type[%s]", dataType), http.StatusInternalServerError)
		}
		return record, nil
	}
	schemaName, _, ex := SchemaDoc.ParseDataType(dataType)
	if ex != nil {
		return nil, Http.WrapError(ex, fmt.Sprintf("failed to parse schema type[%s]", dataType), http.StatusBadRequest)
	}
	referral, err := h.GetReferral(schemaName)
	if err != nil {
		return nil, err
	}
	return referral.GetSchema(dataType, h.log)
}

func (h *Handler) GetRecord(dataType string, dataId string) (*Record.Record, *Http.HttpError) {
	var data interface{}
	var err *Http.HttpError
	switch dataType {
	case JsonKey.Schema:
		return h.GetSchemaRecord(dataId)
	case Schema.Inventory:
		data, err = h.GetData(Schema.Inventory, dataId)
	case RefRecord.Referral:
		referral, err := h.GetReferral(dataId)
		if err != nil {
			return nil, err
		}
		// when query referral, we also want to display schema of the dataType
		record, err := referral.GetSchema(dataId, h.log)
		if err != nil {
			return nil, err
		}
		referral.Schema = record.Data
		return referral.GetRecord(), nil
	case SchemaPath.PathName:
		data, err = h.GetData(SchemaPath.PathName, dataId)
	default:
		data, err = h.GetDataServiceData(dataType, dataId)
	}
	if err != nil {
		return nil, err
	}
	record, e := Record.LoadMap(data.(map[string]interface{}))
	if e != nil {
		return nil, Http.WrapError(e, "failed to load result data as Record", http.StatusInternalServerError)
	}
	return record, nil
}

func (h *Handler) ListData(dataType string) ([]map[string]interface{}, *Http.HttpError) {
	if _, ok := InvTypes[dataType]; !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("[type]=[%s] is not supported", dataType), http.StatusBadRequest)
	}
	err := h.Db.CreateTable(dataType, nil)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	args := make(map[string]interface{})
	args[DbIface.Table] = dataType
	result, err := h.Db.Get(args)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	return result, nil
}

func (h *Handler) GetDataByPath(dataType string, idPath string, nextPath string) (interface{}, *Http.HttpError) {
	conn := SchemaPathData.Connection{
		FuncRecord: h.GetRecord,
	}
	dataPath := idPath
	if nextPath != "" {
		dataPath = fmt.Sprintf("%s/%s", idPath, nextPath)
	}
	query, err := SchemaPath.CreateQuery(&conn, dataType, dataPath)
	if err != nil {
		return nil, err
	}
	result, err := query.WalkValue()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, Http.NewHttpError(fmt.Sprintf("walk SchemaPath with no value.from [path]=[%s]", dataPath), http.StatusNotFound)
	}
	return result, nil
}

func (h *Handler) GetDataServiceData(dataType string, dataId string) (interface{}, *Http.HttpError) {
	referral, err := h.GetReferral(dataType)
	if err != nil {
		return nil, err
	}
	dsUrl, e := referral.DsInfo.GetUrl()
	if e != nil {
		return nil, Http.WrapError(e, fmt.Sprintf("no good url for DataService=[%s]", referral.DsId), http.StatusInternalServerError)
	}
	idPath, e := Http.URLPathJoin(dsUrl, dataType, dataId)
	if e != nil {
		return nil, Http.WrapError(e, fmt.Sprintf("failed to parse url from data service [%s]=[%s], url=[%s]", Record.DataId, referral.DsId, dsUrl), http.StatusInternalServerError)
	}
	data, code, e := Http.GetRestData(*idPath)
	if e != nil {
		if code == http.StatusNotFound {
			return nil, Http.NewHttpError(e.Error(), code)
		}
		return nil, Http.NewHttpError(fmt.Sprintf("failed to get data from REST URL=[%s]", *idPath), code)

	}
	result, ok := data.(map[string]interface{})
	if !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("data from [%s] is not a validate record data map[string]interface{}", *idPath), http.StatusInternalServerError)
	}
	return result, nil
}

func (h *Handler) GetData(dataType string, dataId string) (interface{}, *Http.HttpError) {
	err := h.Db.CreateTable(dataType, nil)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	args := make(map[string]interface{})
	args[DbIface.Table] = dataType
	args[Record.DataId] = dataId
	recordList, err := h.Db.Get(args)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	if len(recordList) == 0 {
		return nil, Http.NewHttpError(fmt.Sprintf("object of type “%s” with id <%s> not found", dataType, dataId), http.StatusNotFound)

	}
	return recordList[0], nil
}

func (h *Handler) GetReferralRecord(dataType string) (*Record.Record, *Http.HttpError) {
	referralData, err := h.GetData(RefRecord.Referral, dataType)
	if err != nil {
		return nil, err
	}
	referralMap, ok := referralData.(map[string]interface{})
	if !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("failed to convert data of [%s] to map[string]interface{}", dataType), http.StatusInternalServerError)

	}
	record, e := Record.LoadMap(referralMap)
	if e != nil {
		return nil, Http.WrapError(e, fmt.Sprintf("failed to load referral data of [%s] as Record", dataType), http.StatusInternalServerError)
	}
	return record, nil
}

func (h *Handler) GetReferral(dataType string) (*RefRecord.ReferralData, *Http.HttpError) {
	record, err := h.GetReferralRecord(dataType)
	if err != nil {
		return nil, err
	}
	referral, e := RefRecord.LoadMap(record.Data)
	if e != nil {
		return nil, Http.NewHttpError(e.Error(), http.StatusBadRequest)
	}
	dsRecord, err := h.GetDsRecord(referral.DsId)
	if err != nil {
		return nil, err
	}
	dsInfo, e := InvRecord.CreateDsInfo(dsRecord.Data)
	if e != nil {
		return nil, Http.NewHttpError(e.Error(), http.StatusBadRequest)
	}
	referral.DsInfo = dsInfo
	return referral, nil
}

func (h *Handler) GetDsRecord(dsId string) (*Record.Record, *Http.HttpError) {
	dsInfoData, err := h.GetData(Schema.Inventory, dsId)
	if err != nil {
		return nil, err
	}
	recordMap, ok := dsInfoData.(map[string]interface{})
	if !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("%s:%s invalid data. convert to map[string]interface{} failed", Schema.Inventory, dsId), http.StatusInternalServerError)

	}
	record, e := Record.LoadMap(recordMap)
	if e != nil {
		return nil, Http.NewHttpError(fmt.Sprintf("%s:%s invalid data. failed to load as Record", Schema.Inventory, dsId), http.StatusInternalServerError)
	}
	return record, nil
}

func (h *Handler) GetDsInfo(dsId string) (*InvRecord.DataServiceInfo, *Http.HttpError) {
	record, err := h.GetDsRecord(dsId)
	if err != nil {
		return nil, err
	}
	dsInfo, e := InvRecord.CreateDsInfo(record.Data)
	if e != nil {
		return nil, Http.NewHttpError(e.Error(), http.StatusInternalServerError)
	}
	return dsInfo, nil
}

func (h *Handler) PutData(data map[string]interface{}) (string, *Http.HttpError) {
	record, err := Record.LoadMap(data)
	if err != nil {
		return "", Http.WrapError(err, "payload failed to be load as Record", http.StatusBadRequest)
	}
	if _, ok := EditableTypes[record.Type]; !ok {
		return "", Http.NewHttpError(fmt.Sprintf("update on type=[%s] not editable", record.Type), http.StatusBadRequest)
	}
	err = h.Db.CreateTable(record.Type, nil)
	if err != nil {
		return "", Http.WrapError(err, fmt.Sprintf("failed to init table=[%s]", record.Type), http.StatusInternalServerError)
	}
	if record.Type == SchemaPath.PathName {
		_, e := SchemaPath.LoadPathDataMap(record.Data)
		if e != nil {
			return "", e
		}
	}
	args := make(map[string]interface{})
	args[DbIface.Table] = record.Type
	args[Record.DataId] = record.Id
	err = h.Db.Replace(record.Type, args, record.Map())
	if err != nil {
		return "", Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	return record.Id, nil
}

func (h *Handler) DeleteData(dataType string, dataId string) *Http.HttpError {
	if dataType == "" || dataId == "" {
		return Http.NewHttpError("invalid url for delete, expected=[{dataType}/{dataId}]", http.StatusBadRequest)
	}
	if _, ok := EditableTypes[dataType]; !ok {
		return Http.NewHttpError(fmt.Sprintf("delete on type=[%s] not supported", dataType), http.StatusBadRequest)
	}
	err := h.Db.CreateTable(dataType, nil)
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to init table=[%s]", dataType), http.StatusInternalServerError)
	}
	_, e := h.GetData(dataType, dataId)
	if e != nil {
		if e.Status == http.StatusNotFound {
			return nil
		}
		return e
	}
	err = h.Db.Delete(dataType, map[string]interface{}{Record.DataId: dataId})
	if err != nil {
		return Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	return nil
}
