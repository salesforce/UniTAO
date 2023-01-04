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
	"io/ioutil"
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

func (i *DataServiceProxy) Log(message string) {
	i.handler.log.Printf("DsInvProxy: %s", message)
}

func (i *DataServiceProxy) refresh() {
	schemaUrl, ex := Http.URLPathJoin(i.Url, JsonKey.Schema)
	if ex != nil {
		i.Log(fmt.Sprintf("failed to build inv schema url, Error:%s", ex))
		return
	}
	data, code, ex := Http.GetRestData(*schemaUrl)
	if ex != nil {
		i.Log(fmt.Sprintf("inventory=[%s] does not work, code: %d Error: %s", *schemaUrl, code, ex))
		return
	}
	typeList, ok := data.([]interface{})
	if !ok {
		i.Log("bad result from inventory, failed convert to array")
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
		errMsg := "dataType is empty, failed to get Data Source"
		i.Log(errMsg)
		return nil, Http.NewHttpError(errMsg, http.StatusBadRequest)
	}
	_, ok := Common.InternalTypes[dataType]
	if ok {
		errMsg := fmt.Sprintf("internal data type of [%s], not for inventory", dataType)
		i.Log(errMsg)
		return nil, Http.NewHttpError(errMsg, http.StatusBadRequest)
	}
	if _, ok := i.DsInfo[dataType]; !ok {
		i.refresh()
	}
	dsInfo, ok := i.DsInfo[dataType]
	if !ok {
		errMsg := fmt.Sprintf("unknwon data type of [%s] for inventory", dataType)
		i.Log(errMsg)
		return nil, Http.NewHttpError(errMsg, http.StatusBadRequest)
	}
	if dsInfo == nil {
		refUrl, _ := Http.URLPathJoin(i.Url, RefRecord.Referral, dataType)
		dsReferralInfo, status, err := Http.GetRestData(*refUrl)
		if err != nil {
			errMsg := fmt.Sprintf("failed to get referral data type=[%s] from inventory=[%s]", dataType, i.Url)
			i.Log(errMsg)
			i.Log(err.Error())
			return nil, Http.WrapError(err, errMsg, status)
		}
		dsReferralData, ok := dsReferralInfo.(map[string]interface{})
		if !ok {
			errMsg := fmt.Sprintf("invalid [%s] return, not a map. Url=[%s]", RefRecord.Referral, *refUrl)
			i.Log(errMsg)
			return nil, Http.NewHttpError(errMsg, http.StatusInternalServerError)
		}
		dsRefRecord, err := Record.LoadMap(dsReferralData)
		if err != nil {
			errMsg := fmt.Sprintf("invalid [%s] return, failed to load as record. Url=[%s]", RefRecord.Referral, *refUrl)
			i.Log(errMsg)
			i.Log(err.Error())
			return nil, Http.WrapError(err, errMsg, http.StatusInternalServerError)
		}
		dsRef := RefRecord.ReferralData{}
		err = Json.CopyTo(dsRefRecord.Data, &dsRef)
		if err != nil {
			errMsg := fmt.Sprintf("invalid [%s] return, failed to load as ReferralData. Url=[%s]", RefRecord.Referral, *refUrl)
			i.Log(errMsg)
			i.Log(err.Error())
			return nil, Http.WrapError(err, errMsg, http.StatusInternalServerError)
		}
		i.DsInfo[dataType] = dsRef.DsInfo
		dsInfo = dsRef.DsInfo
	}
	i.Log(fmt.Sprintf("DataService[%s] for dataType[%s]", dsInfo.Id, dataType))
	return dsInfo, nil
}

func (i *DataServiceProxy) getDsUrl(dataType string, dataId string) (string, *Http.HttpError) {
	queryType := dataType
	if dataType == CmtIndex.KeyCmtIdx || dataType == JsonKey.Schema {
		queryType = dataId
	}
	dsInfo, ex := i.GetDsInfo(queryType)
	if ex != nil {
		return "", ex
	}
	dsUrl, err := dsInfo.GetUrl()
	if err != nil {
		errMsg := "failed to get Data Service [working] Url"
		i.Log(errMsg)
		i.Log(err.Error())
		return "", Http.WrapError(err, errMsg, http.StatusInternalServerError)
	}
	i.Log(fmt.Sprintf("Data Service: %s[%s] for [%s/%s]", dsInfo.Id, dsUrl, dataType, dataId))
	return dsUrl, nil
}

func (i *DataServiceProxy) getIdUrl(dataType string, dataId string) (string, *Http.HttpError) {
	dsUrl, ex := i.getDsUrl(dataType, dataId)
	if ex != nil {
		return "", ex
	}
	idUrl, err := Http.URLPathJoin(dsUrl, dataType, dataId)
	if err != nil {
		return "", Http.WrapError(err, fmt.Sprintf("failed to build path with [%s, %s, %s]", dsUrl, dataType, dataId), http.StatusInternalServerError)
	}
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
	if dataType == JsonKey.Schema || dataType == CmtIndex.KeyCmtIdx {
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
		i.Log(fmt.Sprintf("%s/%s is local data", dataType, dataId))
		data, err := i.handler.Get(dataType, dataId)
		if err != nil {
			i.Log(fmt.Sprintf("local GET failed. [%s/%s]", dataType, dataId))
			i.Log(err.Error())
			return nil, err
		}
		record, ex := Record.LoadMap(data)
		if ex != nil {
			i.Log(fmt.Sprintf("local data load as record failed. [%s/%s]", dataType, dataId))
			i.Log(ex.Error())
			return nil, Http.WrapError(ex, "failed to load data as Record", http.StatusInternalServerError)
		}
		return record, nil
	}
	i.Log(fmt.Sprintf("%s/%s is not local data", dataType, dataId))
	queryUrl, err := i.getIdUrl(dataType, dataId)
	if err != nil {
		i.Log(fmt.Sprintf("failed get url for [%s/%s]", dataType, dataId))
		i.Log(err.Error())
		return nil, err
	}
	i.Log(fmt.Sprintf("Request GET from [%s]", queryUrl))
	data, code, ex := Http.GetRestData(queryUrl)
	if ex == nil {
		mapData, ok := data.(map[string]interface{})
		if !ok {
			errMsg := fmt.Sprintf("return data is not an object. [url]=[%s]", queryUrl)
			i.Log(errMsg)
			return nil, Http.NewHttpError(errMsg, http.StatusBadRequest)
		}
		record, ex := Record.LoadMap(mapData)
		if ex != nil {
			errMsg := fmt.Sprintf("failed to load data as record. url=[%s]", queryUrl)
			i.Log(errMsg)
			i.Log(ex.Error())
			return nil, Http.WrapError(ex, errMsg, http.StatusInternalServerError)
		}
		return record, nil
	}
	i.Log(fmt.Sprintf("Get failed. code[%d], Error: %s", code, ex.Error()))
	i.Log(ex.Error())
	return nil, Http.NewHttpError(ex.Error(), code)
}

func (i *DataServiceProxy) Post(record *Record.Record) *Http.HttpError {
	isLocal, err := i.IsLocal(record.Type, record.Id)
	if err != nil {
		i.Log(fmt.Sprintf("failed to query local [%s/%s]", record.Type, record.Id))
		i.Log(err.Error())
		return err
	}
	if isLocal {
		i.Log(fmt.Sprintf("[%s] is to be added locally", record.Type))
		return i.handler.Add(record)
	}
	i.Log(fmt.Sprintf("[%s] is to be add on other Data Service", record.Type))
	queryUrl, err := i.getDsUrl(record.Type, record.Id)
	if err != nil {
		i.Log(fmt.Sprintf("failed get url for [%s/%s]", record.Type, record.Id))
		i.Log(err.Error())
		return err
	}
	_, status, ex := Http.SubmitPayload(queryUrl, http.MethodPost, nil, record.Map())
	if ex != nil {
		errMsg := fmt.Sprintf("failed to post [%s]", queryUrl)
		i.Log(errMsg)
		i.Log(ex.Error())
		return Http.WrapError(ex, errMsg, status)
	}
	if status != http.StatusAccepted && status != http.StatusOK {
		errMsg := fmt.Sprintf("failed to post [%s], [%d] is not Accepted[%d]/Ok[%d]", queryUrl, status, http.StatusAccepted, http.StatusOK)
		i.Log(errMsg)
		return Http.NewHttpError(errMsg, status)
	}
	return nil
}

func (i *DataServiceProxy) Put(record *Record.Record) *Http.HttpError {
	isLocal, err := i.IsLocal(record.Type, record.Id)
	if err != nil {
		return err
	}
	if isLocal {
		return i.handler.Set("", "", record)
	}
	queryUrl, err := i.getDsUrl(record.Type, record.Id)
	if err != nil {
		return err
	}
	_, status, ex := Http.SubmitPayload(queryUrl, http.MethodPut, nil, record.Map())
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to put [%s]", queryUrl), http.StatusInternalServerError)
	}
	if status != http.StatusAccepted && status != http.StatusOK {
		return Http.NewHttpError(fmt.Sprintf("failed to put [%s]", queryUrl), status)
	}
	return nil
}

func (i *DataServiceProxy) Patch(dataType string, dataId string, dataPath string, headers map[string]interface{}, data interface{}) *Http.HttpError {
	isLocal, err := i.IsLocal(dataType, dataId)
	if err != nil {
		return err
	}
	if isLocal {
		idPath := fmt.Sprintf("%s/%s", dataId, dataPath)
		_, err := i.handler.Patch(dataType, idPath, headers, data)
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
	resp, status, ex := Http.SubmitPayload(pUrl, http.MethodPatch, headers, data)
	respTxt := ""
	if resp != nil {
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			respTxt = fmt.Sprintf("failed to read response body. Error:%s", err)
		} else {
			respTxt = string(respData)
		}
	}
	if ex != nil {
		err = Http.WrapError(ex, fmt.Sprintf("failed to patch [%s]", pUrl), http.StatusInternalServerError)
		if respTxt != "" {
			err.Context = append(err.Context, respTxt)
		}
		return err
	}
	if status != http.StatusAccepted && status != http.StatusOK {
		err = Http.NewHttpError(fmt.Sprintf("failed to create [%s]", pUrl), status)
		if respTxt != "" {
			err.Context = append(err.Context, respTxt)
		}
		return err
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
