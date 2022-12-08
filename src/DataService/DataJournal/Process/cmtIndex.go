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
	"net/url"

	"github.com/salesforce/UniTAO/lib/Schema/CmtIndex"
	"github.com/salesforce/UniTAO/lib/Schema/Compare"
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
	if dataType == CmtIndex.KeyCmtIdx {
		return true, nil
	}
	_, err := s.Data.Get(CmtIndex.KeyCmtIdx, dataType)
	if err != nil {
		if err.Status == http.StatusNotFound {
			return false, nil
		}
		s.log.Printf("[%s]:%s", s.Name(), err)
		return false, err
	}
	return true, nil
}

func (s *CmtIndexChanges) Log(message string) {
	s.log.Printf("%s: %s", s.Name(), message)
}

func (s *CmtIndexChanges) ProcessEntry(dataType string, dataId string, entry *ProcessIface.JournalEntry) *Http.HttpError {
	switch dataType {
	case CmtIndex.KeyCmtIdx:
		return s.processCmtIndexChange(dataId, entry)
	default:
		return s.processDataChange(dataType, dataId, entry)
	}
}

func (s *CmtIndexChanges) processCmtIndexChange(dataType string, entry *ProcessIface.JournalEntry) *Http.HttpError {
	// if data remove on the CmtIndex record, It means the Schema with IndexTemplate is being removed.
	// Remov the empty subscription by level. record or dataType
	if entry.After == nil {
		return Http.NewHttpError(fmt.Sprintf("deletion on %s/%s, do nothing", CmtIndex.KeyCmtIdx, dataType), http.StatusNotModified)
	}
	if entry.Before == nil {
		return s.processNewSubscription(dataType)
	}
	// if data add on the CmtIndex record then add journal to all data record
	schema, err := s.Data.LocalSchema(dataType, "")
	if err != nil {
		return err
	}
	beforeRec, ex := Record.LoadMap(entry.Before)
	if ex != nil {
		// do not recognize entry as record.
		s.log.Printf("failed to load entry Before. @type=[%s], page=[%d], entry=[%d]", dataType, entry.Page, entry.Idx)
		return nil
	}
	afterRec, ex := Record.LoadMap(entry.After)
	if ex != nil {
		// do not recognize entry as record.
		return Http.NewHttpError(fmt.Sprintf("failed to load entry After. @type=[%s], page=[%d], entry=[%d]", dataType, entry.Page, entry.Idx), http.StatusNotModified)
	}
	cmp := Compare.SchemaCompare{
		Schema: schema.Schema,
	}
	diffs := cmp.CompareObj(beforeRec.Data, afterRec.Data, "")
	for _, diff := range diffs {
		if diff.Action == Compare.Add {
			return s.processNewSubscription(dataType)
		}
	}
	return nil
}

func (s *CmtIndexChanges) processNewSubscription(dataType string) *Http.HttpError {
	if s.Data.AddJournal == nil {
		return Http.NewHttpError("AddJournal=nil, do nothing", http.StatusNotModified)
	}
	allRecords, err := s.Data.QueryDb(dataType, "")
	if err != nil {
		return err
	}
	for _, data := range allRecords {
		record, ex := Record.LoadMap(data)
		if ex != nil {
			s.log.Printf("failed to load record. Error:%s", ex)
			continue
		}
		err = s.Data.AddJournal(dataType, record.Id, nil, record.Map())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CmtIndexChanges) processDataChange(dataType string, dataId string, entry *ProcessIface.JournalEntry) *Http.HttpError {
	if entry.Before == nil && entry.After == nil {
		return Http.NewHttpError(fmt.Sprintf("[%s/%s] entry page/idx=[%d/%d], have both Before and After empty.", dataType, dataId, entry.Page, entry.Idx), http.StatusNotModified)
	}
	cmtIdxData, err := s.Data.GetData(CmtIndex.KeyCmtIdx, dataType)
	if err != nil {
		return err
	}
	cmtIdxRec, ex := Record.LoadMap(cmtIdxData)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to load CmtIdx Data=[%s] as record", dataType), http.StatusInternalServerError)
	}
	cmtIdx, ex := CmtIndex.LoadMap(cmtIdxRec.Data)
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
		afterRec = rec
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
					idPath, ex := template.BuildValue(beforeRec.Data)
					if ex != nil {
						s.log.Printf("failed to build id from template on entry.Before. [%s/%s]", dataType, dataId)
						continue
					}
					beforePath = fmt.Sprintf("%s[%s]", idPath, url.QueryEscape(dataId))
				}
				afterPath := ""
				if afterRec != nil {
					idPath, ex := template.BuildValue(afterRec.Data)
					if ex != nil {
						return Http.WrapError(ex, "not able to build a good idPath, no change", http.StatusNotModified)
					}
					afterPath = fmt.Sprintf("%s[%s]", idPath, url.QueryEscape(dataId))
				}
				if beforePath == afterPath {
					continue
				}
				hasChange = true
				if beforePath != "" {
					s.setIndex(dataType, version, beforePath, "")
				}
				if afterPath != "" {
					s.setIndex(dataType, version, afterPath, dataId)
				}
			}
		}
	}
	if !hasChange {
		return Http.NewHttpError("no change made", http.StatusNotModified)
	}
	return nil
}

func (s *CmtIndexChanges) setIndex(dataType string, version string, dataPath string, idxId string) error {
	dataId, nextPath := Util.ParsePath(dataPath)
	if dataId == "" {
		s.log.Printf("empty dataPath, not able to get data Path to write the index into. @path=[%s]", dataPath)
		return nil
	}
	pathRecord, err := s.Data.Inventory.Get(dataType, dataId)
	if err != nil {
		s.log.Printf("failed to get index record. [%s/%s]", dataType, dataPath)
		return nil
	}
	if pathRecord.Version != version {
		s.log.Printf("index record version=[%s] != expected version[%s]", pathRecord.Version, version)
		return nil
	}
	if idxId == "" {
		err = s.Data.Inventory.Patch(dataType, dataId, nextPath, nil, nil)
	} else {
		err = s.Data.Inventory.Patch(dataType, dataId, nextPath, nil, idxId)
	}
	if err != nil {
		return err
	}
	return nil
}
