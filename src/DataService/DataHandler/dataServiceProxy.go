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
	"DataService/Common"
	"InventoryService/InvRecord"
	"InventoryService/RefRecord"
	"fmt"
	"net/http"

	"github.com/salesforce/UniTAO/lib/Schema/CmtIndex"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

type DataServiceProxy struct {
	handler *Handler
	Url     string
	DsInfo  map[string]*InvRecord.DataServiceInfo
}

func CreateDsProxy(hdl *Handler) *DataServiceProxy {
	inv := DataServiceProxy{
		handler: hdl,
		Url:     hdl.Config.Inv.Url,
		DsInfo:  map[string]*InvRecord.DataServiceInfo{},
	}
	inv.refresh()
	return &inv
}

func (i *DataServiceProxy) log(message string) {
	i.handler.log.Printf("DsInvProxy: %s", message)
}

func (i *DataServiceProxy) refresh() {
	schemaUrl, ex := Http.URLPathJoin(i.Url, JsonKey.Schema)
	if ex != nil {
		i.log(fmt.Sprintf("failed to build inv schema url, Error:%s", ex))
		return
	}
	data, code, ex := Http.GetRestData(*schemaUrl)
	if ex != nil {
		i.log(fmt.Sprintf("inventory=[%s] does not work, code: %d Error: %s", *schemaUrl, code, ex))
		return
	}
	typeList, ok := data.([]interface{})
	if !ok {
		i.log("bad result from inventory, failed convert to array")
		return
	}
	for _, item := range typeList {
		dataType := item.(string)
		if _, ok := Common.InternalTypes[dataType]; !ok {
			i.DsInfo[dataType] = nil
		}
	}
}

func (i *DataServiceProxy) GetDsInfo(dataType string) (*InvRecord.DataServiceInfo, *Http.HttpError) {
	if dataType == "" {
		return nil, Http.NewHttpError("dataType is empty, failed to get Data Source", http.StatusBadRequest)
	}
	_, ok := Common.InternalTypes[dataType]
	if ok {
		return nil, Http.NewHttpError(fmt.Sprintf("internal data type of [%s], not for inventory", dataType), http.StatusBadRequest)
	}
	dsInfo, ok := i.DsInfo[dataType]
	if !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("unknwon data type of [%s] for inventory", dataType), http.StatusBadRequest)
	}
	if dsInfo == nil {
		refUrl, _ := Http.URLPathJoin(i.Url, RefRecord.Referral, dataType)
		dsReferralInfo, status, err := Http.GetRestData(*refUrl)
		if err != nil {
			return nil, Http.WrapError(err, fmt.Sprintf("failed to get referral data type=[%s] from inventory=[%s]", dataType, i.Url), status)
		}
		dsReferralData, ok := dsReferralInfo.(map[string]interface{})
		if !ok {
			return nil, Http.NewHttpError(fmt.Sprintf("invalid [%s] return, not a map. Url=[%s]", RefRecord.Referral, *refUrl), http.StatusInternalServerError)
		}
		dsRefRecord, err := Record.LoadMap(dsReferralData)
		if err != nil {
			return nil, Http.WrapError(err, fmt.Sprintf("invalid [%s] return, failed to load as record. Url=[%s]", RefRecord.Referral, *refUrl), http.StatusInternalServerError)
		}
		dsRef := RefRecord.ReferralData{}
		err = Json.CopyTo(dsRefRecord.Data, &dsRef)
		if err != nil {
			return nil, Http.WrapError(err, fmt.Sprintf("invalid [%s] return, failed to load as ReferralData. Url=[%s]", RefRecord.Referral, *refUrl), http.StatusInternalServerError)
		}
		i.DsInfo[dataType] = dsRef.DsInfo
		dsInfo = dsRef.DsInfo
	}
	return dsInfo, nil
}

func (i *DataServiceProxy) getIdUrl(dataType string, dataId string) (string, *Http.HttpError) {
	queryType := dataType
	if dataType == CmtIndex.KeyCmtIdx {
		queryType = dataId
	}
	dsInfo, ex := i.GetDsInfo(queryType)
	if ex != nil {
		return "", ex
	}
	dsUrl, err := dsInfo.GetUrl()
	if err != nil {
		return "", Http.WrapError(err, "failed to get Data Service [working] Url", http.StatusInternalServerError)
	}
	idUrl, err := Http.URLPathJoin(dsUrl, dataType, dataId)
	return *idUrl, nil
}

func (i *DataServiceProxy) List(dataType string) ([]interface{}, *Http.HttpError) {
	_, err := i.handler.LocalSchema(dataType, "")
	if err == nil {
		return i.handler.List(dataType)
	}
	if err.Status != http.StatusNotFound {
		return nil, err
	}
	// get list from inventory Url directly
	if _, ok := i.DsInfo[dataType]; !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("unknown type of [%s]", dataType), http.StatusBadRequest)
	}
	typeUrl, ex := Http.URLPathJoin(i.handler.Config.Inv.Url, dataType)
	if ex != nil {
		return nil, Http.WrapError(err, "failed to build data list url", http.StatusInternalServerError)
	}
	data, code, ex := Http.GetRestData(*typeUrl)
	if ex != nil {
		return nil, Http.WrapError(ex, fmt.Sprintf("inventory query=[%s] does not work", *typeUrl), code)
	}
	idList, ok := data.([]interface{})
	if !ok {
		return nil, Http.NewHttpError("bad result from inventory, failed convert to array", http.StatusBadRequest)
	}
	return idList, nil
}

func (i *DataServiceProxy) IsLocal(dataType string, dataId string) (bool, *Http.HttpError) {
	var err *Http.HttpError
	if dataType == JsonKey.Schema {
		if dataId == "" {
			return false, nil
		}
		_, ex := i.handler.LocalSchema(dataId, "")
		err = ex
	} else {
		_, ex := i.handler.LocalSchema(dataType, "")
		err = ex
	}
	if err == nil {
		return true, nil
	}
	if err.Status == http.StatusNotFound {
		return false, nil
	}
	return false, err
}

func (i *DataServiceProxy) Get(dataType string, dataId string) (*Record.Record, *Http.HttpError) {
	isLocal, err := i.IsLocal(dataType, dataId)
	if err != nil {
		return nil, err
	}
	if isLocal {
		data, err := i.handler.Get(dataType, dataId)
		if err != nil {
			return nil, err
		}
		record, ex := Record.LoadMap(data)
		if ex != nil {
			return nil, Http.WrapError(ex, "failed to load data as Record", http.StatusInternalServerError)
		}
		return record, nil
	}
	queryUrl, err := i.getIdUrl(dataType, dataId)
	if err != nil {
		return nil, err
	}
	data, code, ex := Http.GetRestData(queryUrl)
	if ex == nil {
		mapData, ok := data.(map[string]interface{})
		if !ok {
			return nil, Http.NewHttpError(fmt.Sprintf("return data is not an object. [url]=[%s]", queryUrl), http.StatusBadRequest)
		}
		record, err := Record.LoadMap(mapData)
		if err != nil {
			return nil, Http.WrapError(err, fmt.Sprintf("failed to load data as record. url=[%s]", queryUrl), http.StatusInternalServerError)
		}
		return record, nil
	}
	return nil, Http.NewHttpError(ex.Error(), code)
}

func (i *DataServiceProxy) Post(record *Record.Record) *Http.HttpError {
	isLocal, err := i.IsLocal(record.Type, record.Id)
	if err != nil {
		return err
	}
	if isLocal {
		return i.handler.Add(record)
	}
	queryUrl, err := i.getIdUrl(record.Type, record.Id)
	if err != nil {
		return err
	}
	_, status, ex := Http.SubmitPayload(queryUrl, http.MethodPost, nil, record)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to post [%s]", queryUrl), http.StatusInternalServerError)
	}
	if status != http.StatusAccepted && status != http.StatusOK {
		return Http.NewHttpError(fmt.Sprintf("failed to post [%s]", queryUrl), status)
	}
	return nil
}

func (i *DataServiceProxy) Put(record *Record.Record) *Http.HttpError {
	isLocal, err := i.IsLocal(record.Type, record.Id)
	if err != nil {
		return err
	}
	if isLocal {
		return i.handler.Set(record)
	}
	queryUrl, err := i.getIdUrl(record.Type, record.Id)
	if err != nil {
		return err
	}
	_, status, ex := Http.SubmitPayload(queryUrl, http.MethodPut, nil, record)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to put [%s]", queryUrl), http.StatusInternalServerError)
	}
	if status != http.StatusAccepted && status != http.StatusOK {
		return Http.NewHttpError(fmt.Sprintf("failed to put [%s]", queryUrl), status)
	}
	return nil
}

func (i *DataServiceProxy) Patch(dataType string, dataId string, dataPath string, headers map[string]string, data interface{}) *Http.HttpError {
	isLocal, err := i.IsLocal(dataType, dataId)
	if err != nil {
		return err
	}
	if isLocal {
		idPath := fmt.Sprintf("%s/%s", dataId, dataPath)
		_, err := i.handler.Patch(dataType, idPath, data)
		if err != nil {
			return err
		}
		return nil
	}
	idUrl, err := i.getIdUrl(dataType, dataId)
	if err != nil {
		return err
	}
	pUrl := fmt.Sprintf("%s/%s", idUrl, dataPath)
	_, status, ex := Http.SubmitPayload(pUrl, http.MethodPatch, headers, data)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to patch [%s]", pUrl), http.StatusInternalServerError)
	}
	if status != http.StatusAccepted && status != http.StatusOK {
		return Http.NewHttpError(fmt.Sprintf("failed to create [%s]", pUrl), status)
	}
	return nil
}

func (i *DataServiceProxy) Delete(dataType string, dataId string) *Http.HttpError {
	isLocal, err := i.IsLocal(dataType, dataId)
	if err != nil {
		return err
	}
	if isLocal {
		return i.handler.Inventory.handler.Delete(dataType, dataId)
	}
	idUrl, err := i.getIdUrl(dataType, dataId)
	if err != nil {
		return err
	}
	_, status, ex := Http.SubmitPayload(idUrl, http.MethodDelete, nil, nil)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to delete [%s]", idUrl), http.StatusInternalServerError)
	}
	if status != http.StatusAccepted && status != http.StatusOK {
		return Http.NewHttpError(fmt.Sprintf("failed to delete [%s]", idUrl), status)
	}
	return nil
}
