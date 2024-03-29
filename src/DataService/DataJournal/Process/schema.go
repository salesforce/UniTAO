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

// process module for schema changes
package Process

import (
	"DataService/Common"
	"DataService/DataHandler"
	"DataService/DataJournal/ProcessIface"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/salesforce/UniTAO/lib/Schema/CmtIndex"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

type SchemaChanges struct {
	Data *DataHandler.Handler
	log  *log.Logger
}

func NewSchemaProcess(data *DataHandler.Handler, logger *log.Logger) (ProcessIface.JournalProcess, error) {
	if data == nil {
		return nil, fmt.Errorf("dataHander cannot be nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	process := SchemaChanges{
		Data: data,
		log:  logger,
	}
	return &process, nil
}

func (s *SchemaChanges) Name() string {
	return "Schema Change Process"
}

func (s *SchemaChanges) HandleType(dataType string, version string) (bool, error) {
	if dataType == JsonKey.Schema {
		return true, nil
	}
	return false, nil
}

func (s *SchemaChanges) Log(message string) {
	s.log.Printf("%s: %s", s.Name(), message)
}

func (s *SchemaChanges) ProcessEntry(dataType string, dataId string, entry *ProcessIface.JournalEntry) *Http.HttpError {
	entryId := fmt.Sprintf("%s/%s/%d-%d", dataType, dataId, entry.Page, entry.Idx)
	s.Log(fmt.Sprintf("process entry [%s]", entryId))
	if dataType == JsonKey.Schema {
		return s.processSchemaChange(entry, entryId)
	}
	return nil
}

func (s *SchemaChanges) processSchemaChange(entry *ProcessIface.JournalEntry, entryId string) *Http.HttpError {
	if entry.Before != nil && entry.After != nil {
		beforeRecord, err := Record.LoadMap(entry.Before)
		if err != nil {
			return Http.WrapError(err, fmt.Sprintf("failed to load before record on entry[%s]", entryId), http.StatusInternalServerError)
		}
		beforeVer := beforeRecord.Data[JsonKey.Version].(string)
		archiveId := SchemaDoc.ArchivedSchemaId(beforeRecord.Id, beforeVer)
		afterRecord, err := Record.LoadMap(entry.After)
		if err != nil {
			return Http.WrapError(err, fmt.Sprintf("failed to load after record on entry[%s]", entryId), http.StatusInternalServerError)
		}
		if archiveId != afterRecord.Id {
			return Http.NewHttpError(fmt.Sprintf("invalid patch schema id[%s], expect:[%s]", afterRecord.Id, archiveId), http.StatusInternalServerError)
		}
		_, ex := s.Data.LocalSchema(beforeRecord.Id, "")
		if ex != nil {
			s.Log(fmt.Sprintf("new schema of [%s] not available yet, please add or roll back manually", beforeRecord.Id))
			return ex
		}
		return nil
	}
	if entry.Before != nil {
		record, err := Record.LoadMap(entry.Before)
		if err != nil {
			return Http.WrapError(err, fmt.Sprintf("failed to load Before from entry[%s]", entryId), http.StatusInternalServerError)
		}
		return s.removeCMTSubscription(record)
	}
	reccord, err := Record.LoadMap(entry.After)
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to load After from entry[%s]", entryId), http.StatusInternalServerError)
	}
	return s.subscribeCMT(reccord)
}

func (s *SchemaChanges) removeCMTSubscription(schemaRec *Record.Record) *Http.HttpError {
	schema, err := SchemaDoc.New(schemaRec.Data)
	if err != nil {
		return Http.WrapError(err, "invalid schema of before schema, do nothing", http.StatusNotModified)
	}
	idxList := CmtIndex.FindAutoIndex(schema, "")
	for _, idx := range idxList {
		if _, ok := Common.InternalTypes[idx.ContentType]; ok {
			return Http.NewHttpError(fmt.Sprintf("invalid link Type for CMT. [%s] is internal type", idx.ContentType), http.StatusNotModified)
		}
		err := s.removeIdxSubscription(schema.Id, schema.Version, idx)
		if err != nil && err.Status != http.StatusNotModified {
			s.Log(fmt.Sprintf("failed to remove idx, [%s/%s], template[%s]", schema.Id, schema.Version, idx.IndexTemplate))
			return err
		}
	}
	return nil
}

func (s *SchemaChanges) subscribeCMT(schemaRec *Record.Record) *Http.HttpError {
	schema, err := SchemaDoc.New(schemaRec.Data)
	if err != nil {
		return Http.WrapError(err, "invalid schema of before schema, do nothing", http.StatusNotModified)
	}
	idxList := CmtIndex.FindAutoIndex(schema, "")
	if err != nil {
		return nil
	}
	for _, idx := range idxList {
		if _, ok := Common.InternalTypes[idx.ContentType]; ok {
			return Http.NewHttpError(fmt.Sprintf("invalid link Type for CMT. [%s] is internal type", idx.ContentType), http.StatusNotModified)
		}
		err := s.subscribeIndex(schema.Id, schema.Version, idx)
		if err != nil {
			s.Log(fmt.Sprintf("failed to add Idx, [%s/%s], template:[%s]", schema.Id, schema.Version, idx.IndexTemplate))
			return err
		}
	}
	return nil
}

func (s *SchemaChanges) subscribeIndex(dataType string, version string, idx *CmtIndex.AutoIndex) *Http.HttpError {
	cmtIdx := CmtIndex.CmtIndex{
		DataType: idx.ContentType,
		Subscriber: map[string]CmtIndex.CmtSubscriber{
			dataType: {
				DataType: dataType,
				VersionIndex: map[string]CmtIndex.VersionIndex{
					version: {
						Version: version,
						IndexTemplate: []interface{}{
							idx.IndexTemplate,
						},
					},
				},
			},
		},
	}
	err := s.createCmtIdx(cmtIdx)
	if err != nil {
		return err
	}
	return nil
}

func (s *SchemaChanges) removeIdxSubscription(dataType string, version string, idx *CmtIndex.AutoIndex) *Http.HttpError {
	_, err := s.Data.Inventory.Get(CmtIndex.KeyCmtIdx, idx.ContentType)
	if err != nil {
		if err.Status != http.StatusNotFound {
			return err
		}
		return Http.WrapError(err, fmt.Sprintf("%s/%s not found, do nothing", CmtIndex.KeyCmtIdx, idx.ContentType), http.StatusNotModified)
	}
	verPath := fmt.Sprintf("%s/%s/%s/%s", CmtIndex.KeyCmtSubscriber, dataType, JsonKey.IndexTemplate, version)
	return s.Data.Inventory.Patch(CmtIndex.KeyCmtIdx, idx.ContentType, verPath, nil, nil)
}

func (s *SchemaChanges) createCmtIdx(idx CmtIndex.CmtIndex) *Http.HttpError {
	idxRec, err := s.Data.Inventory.Get(CmtIndex.KeyCmtIdx, idx.DataType)
	if err != nil {
		if err.Status != http.StatusNotFound {
			return Http.WrapError(err, fmt.Sprintf("failed to get %s/%s", CmtIndex.KeyCmtIdx, idx.DataType), http.StatusInternalServerError)
		}
		s.Log(fmt.Sprintf("inventory post record %s/%s", CmtIndex.KeyCmtIdx, idx.DataType))
		idxRec := idx.Record()
		idxStr, _ := json.MarshalIndent(idxRec, "", "    ")
		s.Log(string(idxStr))
		err = s.Data.Inventory.Post(idx.Record())
		if err != nil {
			s.Log(fmt.Sprintf("failed to crate cmtIdx. [%s/%s], Error:%s", CmtIndex.KeyCmtIdx, idx.DataType, err))
			return err
		}
		return nil
	}
	currentIdx, ex := CmtIndex.LoadMap(idxRec.Data)
	if ex != nil {
		return Http.NewHttpError("invalid data, failed to load data as CmtIndex", http.StatusInternalServerError)
	}
	hasChange := false
	for dataType, subscriber := range idx.Subscriber {
		typePath := fmt.Sprintf("%s[%s]", CmtIndex.KeyCmtSubscriber, url.QueryEscape(dataType))
		cSubscriber, ok := currentIdx.Subscriber[dataType]
		if !ok {
			// only 1 data type in the new subscriber
			hasChange = true
			data, _ := Json.Copy(subscriber)
			err = s.Data.Inventory.Patch(CmtIndex.KeyCmtIdx, idx.DataType, typePath, nil, data)
			if err != nil {
				s.Log(fmt.Sprintf("failed add dataType[%s] to %s/%s, \nError:%s", idx.DataType, CmtIndex.KeyCmtIdx, currentIdx.DataType, err))
			}
			return err
		}
		for version, idxTemp := range subscriber.VersionIndex {
			versionPath := fmt.Sprintf("%s/%s[%s]", typePath, CmtIndex.KeyVersionIndex, url.QueryEscape(version))
			cVerIdx, ok := cSubscriber.VersionIndex[version]
			if !ok {
				hasChange = true
				data, _ := Json.Copy(idxTemp)
				err = s.Data.Inventory.Patch(CmtIndex.KeyCmtIdx, idx.DataType, versionPath, nil, data)
				if err != nil {
					s.Log(fmt.Sprintf("failed add version [%s/%s] to %s/%s, \nError:%s", idx.DataType, version, CmtIndex.KeyCmtIdx, currentIdx.DataType, err))
				}
				return err
			}
			cPathTempMap := Util.IdxList(cVerIdx.IndexTemplate)
			for _, pathTemplate := range idxTemp.IndexTemplate {
				if _, ok := cPathTempMap[pathTemplate]; !ok {
					tempPath := fmt.Sprintf("%s/%s[%s]", versionPath, JsonKey.IndexTemplate, url.QueryEscape(pathTemplate.(string)))
					hasChange = true
					err = s.Data.Inventory.Patch(CmtIndex.KeyCmtIdx, idx.DataType, tempPath, nil, pathTemplate)
					if err != nil {
						s.Log(fmt.Sprintf("failed add version [%s/%s] with template[%s] to %s/%s, \nError:%s", idx.DataType, version, url.QueryEscape(pathTemplate.(string)), CmtIndex.KeyCmtIdx, currentIdx.DataType, err))
						return err
					}
				}
			}
		}
	}
	if !hasChange {
		s.Log("no change made")
		return Http.NewHttpError("no change made", http.StatusNotModified)
	}
	return nil
}
