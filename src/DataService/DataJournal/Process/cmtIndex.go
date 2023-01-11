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

// process module for cmtIndex changes
package Process

import (
	"DataService/DataHandler"
	"DataService/DataJournal/ProcessIface"
	"fmt"
	"log"
	"net/http"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/CmtIndex"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Template"
)

type CmtIndexChanges struct {
	Data *DataHandler.Handler
	log  *log.Logger
}

func NewCmtIndexProcess(data *DataHandler.Handler, logger *log.Logger) (ProcessIface.JournalProcess, error) {
	if data == nil {
		return nil, fmt.Errorf("dataHander cannot be nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	process := CmtIndexChanges{
		Data: data,
		log:  logger,
	}
	return &process, nil
}

func (c *CmtIndexChanges) Name() string {
	return "cmtIndex change process"
}

func (s *CmtIndexChanges) HandleType(dataType string, version string) (bool, error) {
	if dataType == CmtIndex.KeyCmtIdx || dataType == JsonKey.Schema {
		s.Log(fmt.Sprintf("[%s] record change, do nothing", dataType))
		return false, nil
	}
	needHandle, err := s.isSubscribedType(dataType)
	if err != nil {
		return false, err
	}
	if needHandle {
		return true, nil
	}
	s.Log(fmt.Sprintf("check if data of [%s/%s] is subscribing to any data need to fill in", dataType, version))
	_, idxList, err := s.getSubscriberSchemaIdx(dataType, version)
	if err != nil {
		return false, err
	}
	if len(idxList) > 0 {
		return true, nil
	}
	return false, nil
}

func (s *CmtIndexChanges) isSubscribedType(dataType string) (bool, *Http.HttpError) {
	s.Log(fmt.Sprintf("check if data of [%s] is subscribed", dataType))
	_, err := s.Data.Get(CmtIndex.KeyCmtIdx, dataType)
	if err != nil {
		if err.Status == http.StatusNotFound {
			return false, nil
		}
		s.Log(err.Error())
		return false, err
	}
	return true, nil
}

func (s *CmtIndexChanges) getSubscriberSchemaIdx(dataType string, version string) (*Schema.SchemaOps, []CmtIndex.AutoIndex, *Http.HttpError) {
	schema, err := s.Data.LocalSchema(dataType, version)
	if err != nil {
		s.Log(fmt.Sprintf("[%s/%s] failed to load schema, Error:%s", dataType, version, err))
		return nil, nil, err
	}
	idxList := CmtIndex.FindAutoIndex(schema.Schema, "")
	return schema, idxList, nil
}

func (s *CmtIndexChanges) Log(message string) {
	s.log.Printf("%s: %s", s.Name(), message)
}

func (s *CmtIndexChanges) ProcessEntry(dataType string, dataId string, entry *ProcessIface.JournalEntry) *Http.HttpError {
	s.Log(fmt.Sprintf("process Entry of %s", dataType))
	err := s.processDataChange(dataType, dataId, entry)
	if err != nil {
		return err
	}
	return s.processSubscriberChange(dataType, dataId, entry)
}

func (s *CmtIndexChanges) processDataChange(dataType string, dataId string, entry *ProcessIface.JournalEntry) *Http.HttpError {
	isSubscribed, err := s.isSubscribedType(dataType)
	if err != nil {
		return err
	}
	if !isSubscribed {
		return nil
	}
	if entry.Before == nil && entry.After == nil {
		return Http.NewHttpError(fmt.Sprintf("[%s/%s] entry page/idx=[%d/%d], have both Before and After empty.", dataType, dataId, entry.Page, entry.Idx), http.StatusNotModified)
	}
	cmtIdxData, err := s.Data.GetData(CmtIndex.KeyCmtIdx, dataType)
	if err != nil {
		return err
	}
	cmtIdx, ex := CmtIndex.LoadMap(cmtIdxData)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to load CmtIdx Data=[%s] as CmtIndex", dataType), http.StatusInternalServerError)
	}
	var beforeRec, afterRec *Record.Record
	if entry.Before != nil {
		rec, ex := Record.LoadMap(entry.Before)
		if ex != nil {
			s.log.Printf("failed to load entry.Before as record. [%s/%s], page/idx=[%d/%d], Error:%s", dataType, dataId, entry.Page, entry.Idx, ex)
		}
		beforeRec = rec
	}
	if entry.After != nil {
		rec, ex := Record.LoadMap(entry.After)
		if ex != nil {
			s.log.Printf("failed to load entry.After as record. [%s/%s], page/idx=[%d/%d], Error:%s", dataType, dataId, entry.Page, entry.Idx, ex)
		}
		_, err = s.Data.GetData(rec.Type, rec.Id)
		if err != nil {
			if err.Status != http.StatusNotFound {
				s.Log(fmt.Sprintf("processDataChange[%s/%s]: failed to get current record, Error:%s", rec.Type, rec.Id, err))
				return err
			}
			s.Log(fmt.Sprintf("processDataChange[%s/%s]: current not exists", rec.Type, rec.Id))
		} else {
			afterRec = rec
		}
	}
	hasChange := false
	for dataType, subscriber := range cmtIdx.Subscriber {
		for version, idxTemp := range subscriber.VersionIndex {
			for _, temp := range idxTemp.IndexTemplate {
				template, ex := Template.ParseStr(temp.(string), "{", "}")
				if ex != nil {
					return Http.WrapError(ex, fmt.Sprintf("failed to parse template for [%s/%s], temp=[%s]", dataType, version, idxTemp), http.StatusInternalServerError)
				}
				beforePath := ""
				if beforeRec != nil {
					dataPath, ex := template.BuildValue(beforeRec.Data)
					if ex != nil {
						s.Log(fmt.Sprintf("not able to build a good idPath. Before. [%s/%s], Error:%s", dataType, dataId, ex))
					} else {
						beforeType, idPath := Util.ParsePath(dataPath)
						if dataType == beforeType {
							beforePath = idPath
						}
					}
				}
				afterPath := ""
				if afterRec != nil {
					dataPath, ex := template.BuildValue(afterRec.Data)
					if ex != nil {
						s.Log(fmt.Sprintf("not able to build a good idPath, After. [%s/%s], Error:%s", dataType, dataId, ex))
					} else {
						afterType, idPath := Util.ParsePath(dataPath)
						if dataType == afterType {
							afterPath = idPath
						}
					}
				}
				if beforePath == afterPath {
					continue
				}
				hasChange = true
				if beforePath != "" {
					removePath := fmt.Sprintf("%s[%s]", beforePath, dataId)
					s.Log(fmt.Sprintf("remove before path [%s/%s/%s]", dataType, version, removePath))
					err = s.setIndex(dataType, version, removePath, "")
					if err != nil && err.Status != http.StatusNotModified && err.Status != http.StatusNotFound {
						return Http.WrapError(err, fmt.Sprintf("failed to delete idx @path=[%s/%s/%s]", dataType, version, beforePath), err.Status)
					}
				}
				if afterPath != "" {
					s.Log(fmt.Sprintf("set after path [%s/%s/%s] with id=[%s]", dataType, version, afterPath, dataId))
					err = s.setIndex(dataType, version, afterPath, dataId)
					if err != nil && err.Status != http.StatusNotModified && err.Status != http.StatusNotFound {
						return Http.WrapError(err, fmt.Sprintf("failed to set idx @path=[%s/%s/%s], with id=[%s]", dataType, version, beforePath, dataId), err.Status)
					}
				}
			}
		}
	}
	if !hasChange {
		return Http.NewHttpError("no change made", http.StatusNotModified)
	}
	return nil
}

func (s *CmtIndexChanges) setIndex(dataType string, version string, dataPath string, idxId string) *Http.HttpError {
	dataId, nextPath := Util.ParsePath(dataPath)
	if dataId == "" {
		s.Log(fmt.Sprintf("empty dataPath, not able to get data Path to write the index into. @path=[%s]", dataPath))
		return nil
	}
	headers := map[string]interface{}{
		JsonKey.Version: version,
	}
	var err *Http.HttpError
	if idxId == "" {
		err = s.Data.Inventory.Patch(dataType, dataId, nextPath, headers, nil)
	} else {
		err = s.Data.Inventory.Patch(dataType, dataId, nextPath, headers, idxId)
	}
	if err != nil && err.Status != http.StatusNotModified {
		s.Log(fmt.Sprintf("SetIdx[%s/%s]: failed to patch [%s] with id=[%s],\nError:%s", dataType, dataId, nextPath, idxId, err))
		return err
	}
	s.Log(fmt.Sprintf("SetIdx[%s/%s]: set [%s] with id[%s]", dataType, dataId, nextPath, idxId))
	return nil
}

func (s *CmtIndexChanges) processSubscriberChange(dataType string, dataId string, entry *ProcessIface.JournalEntry) *Http.HttpError {
	beforeRecord, ex := Record.LoadMap(entry.Before)
	if ex != nil {
		msg := fmt.Sprintf("SubscriberChange[%s/%s]: failed to load before record,", dataType, dataId)
		s.Log(fmt.Sprintf("%s, Error:%s", msg, ex))
		return Http.WrapError(ex, msg, http.StatusInternalServerError)
	}
	afterRecord, ex := Record.LoadMap(entry.After)
	if ex != nil {
		msg := fmt.Sprintf("SubscriberChange[%s/%s]: failed to load after record,", dataType, dataId)
		s.Log(fmt.Sprintf("%s, Error:%s", msg, ex))
		return Http.WrapError(ex, msg, http.StatusInternalServerError)
	}
	if afterRecord == nil {
		s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: deleted, no work to be done", dataType, dataId))
		return nil
	}
	s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: generate Idx compare data", dataType, dataId))
	idxCmp, err := s.getIdxPathChanges(beforeRecord, afterRecord)
	if err != nil {
		return err
	}
	// for added list
	//     scan CmtIdx Type to find match to fill in
	s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: scan all CmtContentType record to fill autoIndexPath", dataType, dataId))
	err = s.fillForNewIdxPath(afterRecord, idxCmp)
	if err != nil {
		return err
	}
	// for sameList, compare value of each list
	//     only check if removed value is a good idx
	//     if it is a good idx, fill it back in
	s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: double check removed idx, and fill back the good ones", dataType, dataId))
	return s.fillBackGoodIdx(afterRecord, idxCmp)
}

type AutoIdxCompare struct {
	Key     string
	Idx     *CmtIndex.AutoIndex
	Before  map[string]map[string]interface{}
	New     map[string]map[string]interface{}
	Removed map[string]map[string]interface{}
}

func newAutoIdxCompare(idx *CmtIndex.AutoIndex) *AutoIdxCompare {
	return &AutoIdxCompare{
		Key:     idx.AttrPath,
		Idx:     idx,
		Before:  map[string]map[string]interface{}{},
		New:     map[string]map[string]interface{}{},
		Removed: map[string]map[string]interface{}{},
	}
}

func (s *CmtIndexChanges) getIdxPathChanges(before *Record.Record, after *Record.Record) (map[string]*AutoIdxCompare, *Http.HttpError) {
	idxDiffMap := map[string]*AutoIdxCompare{}
	schema, idxList, err := s.getSubscriberSchemaIdx(after.Type, after.Version)
	if err != nil {
		return nil, err
	}
	s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: prepare idxDiffMap", after.Type, after.Id))
	prepareBefore := false
	if before != nil && before.Version == after.Version {
		s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: same version[%s] data change, prepare Before", after.Type, after.Id, before.Version))
		prepareBefore = true
	}
	for _, idx := range idxList {
		autoIdxDiff := newAutoIdxCompare(&idx)
		if prepareBefore {
			autoIdxDiff.Before = idx.ExplorerIdxPath(schema.Schema, before)
		}
		idxDiffMap[autoIdxDiff.Key] = autoIdxDiff
	}
	s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: prepare new path list and removed path values", after.Type, after.Id))
	for _, autoIdxDiff := range idxDiffMap {
		pathData := autoIdxDiff.Idx.ExplorerIdxPath(schema.Schema, after)
		for idxPath, idxData := range pathData {
			if _, ok := autoIdxDiff.Before[idxPath]; !ok {
				s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: new idx path added [%s]", after.Type, after.Id, idxPath))
				autoIdxDiff.New[idxPath] = idxData
				continue
			}
			autoIdxDiff.Removed[idxPath] = map[string]interface{}{}
			for idx := range autoIdxDiff.Before[idxPath] {
				if _, ok := idxData[idx]; !ok {
					s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: path [%s] removed idx[%s]", after.Type, after.Id, idxPath, idx))
					autoIdxDiff.Removed[idxPath][idx] = autoIdxDiff.Before[idxPath][idx]
				}
			}
			if len(autoIdxDiff.Removed[idxPath]) == 0 {
				delete(autoIdxDiff.Removed, idxPath)
			}
		}
	}
	return idxDiffMap, nil
}

func (s *CmtIndexChanges) getCmtRecords(after *Record.Record, idxDiffMap map[string]*AutoIdxCompare) (map[string]map[string]*Record.Record, *Http.HttpError) {
	s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: collect unindexed id", after.Type, after.Id))
	cache := map[string]map[string]*Record.Record{}
	for _, idxCmp := range idxDiffMap {
		if len(idxCmp.New) == 0 {
			s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: no new idxPath on template[%s]", after.Type, after.Id, idxCmp.Key))
			continue
		}
		idList, err := s.Data.Inventory.List(idxCmp.Idx.ContentType)
		if err != nil {
			s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: failed to query type of [%s], Error:%s", after.Type, after.Id, idxCmp.Idx.ContentType, err))
			return nil, err
		}
		if len(idList) == 0 {
			s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: no record from type of [%s]", after.Type, after.Id, idxCmp.Idx.ContentType))
			continue
		}
		typeCache, ok := cache[idxCmp.Idx.ContentType]
		if !ok {
			typeCache = map[string]*Record.Record{}
		}
		for _, id := range idList {
			if _, ok := typeCache[id.(string)]; ok {
				continue
			}
			foundId := false
			for _, idMap := range idxCmp.New {
				if _, ok := idMap[id.(string)]; ok {
					foundId = true
					break
				}
			}
			if foundId {
				continue
			}
			typeCache[id.(string)] = nil
		}
		cache[idxCmp.Idx.ContentType] = typeCache
	}
	s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: retrieve records for cache", after.Type, after.Id))
	for dataType, typeCache := range cache {
		s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: retrieve records of type[%s]", after.Type, after.Id, dataType))
		for dataId := range typeCache {
			s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: retrieve record[%s/%s]", after.Type, after.Id, dataType, dataId))
			record, err := s.Data.Inventory.Get(dataType, dataId)
			if err != nil {
				s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: retrieve record[%s/%s] failed. Error:%s", after.Type, after.Id, dataType, dataId, err))
				return nil, err
			}
			typeCache[dataId] = record
		}
	}
	return cache, nil
}

func (s *CmtIndexChanges) fillForNewIdxPath(after *Record.Record, compare map[string]*AutoIdxCompare) *Http.HttpError {
	cache, err := s.getCmtRecords(after, compare)
	if err != nil {
		return err
	}
	for _, idxCmp := range compare {
		if len(idxCmp.New) == 0 {
			continue
		}
		typeCache, ok := cache[idxCmp.Idx.ContentType]
		if !ok {
			continue
		}
		strTemp, ex := Template.ParseStr(idxCmp.Idx.IndexTemplate, "{", "}")
		if ex != nil {
			msg := fmt.Sprintf("failed to parse indexTemplate[%s]", idxCmp.Idx.IndexTemplate)
			s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: %s, Error:%s", after.Type, after.Id, msg, ex))
			return Http.WrapError(ex, msg, http.StatusInternalServerError)
		}
		for id, record := range typeCache {
			idxPath, ex := strTemp.BuildValue(record.Data)
			if ex != nil {
				continue
			}
			idMap, pathExists := idxCmp.New[idxPath]
			if !pathExists {
				continue
			}
			if _, ok := idMap[id]; ok {
				continue
			}
			_, idPath := Util.ParsePath(idxPath)
			err = s.setIndex(after.Type, after.Version, idPath, id)
			if err != nil && err.Status != http.StatusNotModified {
				return err
			}
		}
	}
	return nil
}

func (s *CmtIndexChanges) fillBackGoodIdx(after *Record.Record, compare map[string]*AutoIdxCompare) *Http.HttpError {
	for _, idxCmp := range compare {
		if len(idxCmp.Removed) == 0 {
			continue
		}
		strTemp, ex := Template.ParseStr(idxCmp.Idx.IndexTemplate, "{", "}")
		if ex != nil {
			msg := fmt.Sprintf("failed to parse indexTemplate[%s]", idxCmp.Idx.IndexTemplate)
			s.Log(fmt.Sprintf("SubscriberChange[%s/%s]: %s, Error:%s", after.Type, after.Id, msg, ex))
			return Http.WrapError(ex, msg, http.StatusInternalServerError)
		}
		for idxPath, idMap := range idxCmp.Removed {
			for id := range idMap {
				record, err := s.Data.Inventory.Get(idxCmp.Idx.ContentType, id)
				if err != nil {
					if err.Status == http.StatusNotFound {
						continue
					}
				}
				iPath, ex := strTemp.BuildValue(record.Data)
				if ex != nil {
					continue
				}
				if iPath != idxPath {
					continue
				}
				_, idPath := Util.ParsePath(idxPath)
				err = s.setIndex(after.Type, after.Version, idPath, id)
				if err != nil && err.Status != http.StatusNotModified {
					return err
				}
			}
		}
	}
	return nil
}
