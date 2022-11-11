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

// functions to record all data changes
package DataJournal

import (
	"Data/DbIface"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

const (
	KeyJournal      = "journal"
	KeyCmtIdx       = "cmtIdx"
	CurrentVer      = "0.0.1"
	MaxEntryPerPage = 10
	KeyActive       = "active"
	KeyAfter        = "after"
	KeyArchived     = "archived"
	KeyBefore       = "before"
	KeyDataId       = "dataId"
	KeyDataType     = "dataType"
	KeyId           = "id"
	KeyIdx          = "idx"
	KeyPage         = "page"
	KeyTime         = "time"
)

type JournalLib struct {
	db    DbIface.Database
	table string
	lock  sync.Mutex
}

func NewJournalLib(db DbIface.Database, table string) *JournalLib {
	return &JournalLib{
		db:    db,
		table: table,
	}
}

func PageId(dataType string, dataId string, idx int) string {
	pageId := fmt.Sprintf("%s:%s_%s:%s_%s:%d", KeyDataType, dataType, KeyDataId, dataId, KeyPage, idx)
	return pageId
}

func NewJournalPage(dataType string, dataId string, idx int) *Record.Record {

	page := map[string]interface{}{
		KeyDataType: dataType,
		KeyDataId:   dataId,
		KeyIdx:      idx,
		KeyActive:   []interface{}{},
		KeyArchived: []interface{}{},
	}
	record := Record.NewRecord(KeyJournal, CurrentVer, PageId(dataType, dataId, idx), page)
	return record
}

func NewJournalEntry(before map[string]interface{}, after map[string]interface{}) map[string]interface{} {
	currentTime := time.Now().String()
	entry := map[string]interface{}{
		KeyTime: currentTime,
	}
	if before != nil {
		entry[KeyBefore] = before
	}
	if after != nil {
		entry[KeyAfter] = after
	}
	return entry
}

func ParseJournalId(journalId string) (string, string, int, error) {
	typeTag := fmt.Sprintf("%s:", KeyDataType)
	idTag := fmt.Sprintf("_%s:", KeyDataId)
	pageTag := fmt.Sprintf("_%s:", KeyPage)
	if !strings.HasPrefix(journalId, typeTag) {
		return "", "", 0, fmt.Errorf("invalid journalId=[%s], missing prefix=[%s]", journalId, KeyDataType)
	}
	typeStr, idStr := Util.ParseCustomPath(journalId, idTag)
	dataType := typeStr[len(typeTag):]
	if idStr == "" {
		if strings.Contains(dataType, pageTag) {
			return "", "", 0, fmt.Errorf("invalid journalId=[%s], missing tag=[%s]", journalId, idTag)
		}
		return dataType, "", 0, nil
	}
	dataId, pageStr := Util.ParseCustomPath(idStr, pageTag)
	if pageStr == "" {
		return dataType, dataId, 0, nil
	}
	idx, err := strconv.Atoi(pageStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid journalId=[%s], failed to convert pageNo=[%s] to int, Error:%s", journalId, pageStr, err)
	}
	return dataType, dataId, idx, nil
}

func (j *JournalLib) GetJournal(journalId string) (interface{}, *Http.HttpError) {
	dataType, dataId, idx, err := ParseJournalId(journalId)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusBadRequest)
	}
	if idx < 0 {
		return nil, Http.NewHttpError(fmt.Sprintf("invalid page idx:[%d] in id=[%s], expect 0..., 0 means list all pages", idx, journalId), http.StatusNotFound)
	}
	switch {
	case dataType == "":
		return j.ListJournalTypes()
	case dataId == "":
		return j.ListJournalIds(dataType)
	case idx == 0:
		return j.ListJournalPages(dataType, dataId)
	default:
		return j.GetJournalPage(dataType, dataId, idx)
	}
}

func (j *JournalLib) ListJournalTypes() ([]string, *Http.HttpError) {
	journalList, err := j.retrieveJournals("", "")
	if err != nil {
		return nil, err
	}
	typeHash := map[string]int{}
	for _, journal := range journalList {
		activeCount := len(journal[Record.Data].(map[string]interface{})[KeyActive].([]interface{}))
		if activeCount == 0 {
			continue
		}
		dataType := journal[Record.Data].(map[string]interface{})[KeyDataType].(string)
		if _, ok := typeHash[dataType]; !ok {

			typeHash[dataType] = 1
		} else {
			typeHash[dataType] += 1
		}
	}
	typeList := make([]string, 0, len(typeHash))
	for key := range typeHash {
		typeList = append(typeList, key)
	}
	return typeList, nil
}

func (j *JournalLib) ListJournalIds(dataType string) ([]string, *Http.HttpError) {
	journalList, err := j.retrieveJournals(dataType, "")
	if err != nil {
		return nil, err
	}
	idHash := map[string]int{}
	for _, journal := range journalList {
		dataId := journal[Record.Data].(map[string]interface{})[KeyDataId].(string)
		if _, ok := idHash[dataId]; !ok {
			idHash[dataId] = 1
		} else {
			idHash[dataId] += 1
		}
	}
	idList := make([]string, 0, len(idHash))
	for key := range idHash {
		idList = append(idList, key)
	}
	return idList, nil
}

func (j *JournalLib) ListJournalPages(dataType string, dataId string) ([]map[string]interface{}, *Http.HttpError) {
	journalList, err := j.retrieveJournals(dataType, dataId)
	if err != nil {
		return nil, err
	}
	minIdx := 0
	maxIdx := 0
	pageHash := map[int]map[string]interface{}{}
	for _, journal := range journalList {
		pageIdx := int(journal[Record.Data].(map[string]interface{})[KeyIdx].(float64))
		pageHash[pageIdx] = journal
		if minIdx == 0 || pageIdx < minIdx {
			minIdx = pageIdx
		}
		if maxIdx == 0 || pageIdx > maxIdx {
			maxIdx = pageIdx
		}
	}
	pageList := make([]map[string]interface{}, 0, len(pageHash))
	idx := minIdx
	for idx <= maxIdx {
		if journal, ok := pageHash[idx]; ok {
			pageList = append(pageList, journal)
		}
		idx += 1
	}
	return pageList, nil
}

func (j *JournalLib) CleanArchivedPages(dataType string, dataId string) *Http.HttpError {
	pageList, err := j.ListJournalPages(dataType, dataId)
	if err != nil {
		return err
	}
	if len(pageList) == 0 {
		return nil
	}
	for idx, page := range pageList {
		rec, ex := Record.LoadMap(page)
		if ex != nil {
			return Http.WrapError(ex, fmt.Sprintf("failed to load page[%d] as record. [%s/%s], Error:%s", idx, dataType, dataId, ex), http.StatusInternalServerError)
		}
		activeLen := len(rec.Data[KeyActive].([]interface{}))
		if activeLen != 0 {
			return nil
		}
		archivedLen := len(rec.Data[KeyArchived].([]interface{}))
		if archivedLen < MaxEntryPerPage {
			return nil
		}
		keys := make(map[string]interface{})
		keys[Record.DataType] = rec.Type
		keys[Record.DataId] = rec.Id
		ex = j.db.Delete(j.table, keys)
		if ex != nil {
			return Http.WrapError(ex, fmt.Sprintf("failed to delete record [type/id]=[%s/%s]", dataType, dataId), http.StatusInternalServerError)
		}
	}
	return nil
}

func (j *JournalLib) NextJournalEntry(dataType string, dataId string) (map[string]interface{}, *Http.HttpError) {
	pageList, err := j.ListJournalPages(dataType, dataId)
	if err != nil {
		return nil, err
	}
	for idx, page := range pageList {
		record, e := Record.LoadMap(page)
		if e != nil {
			return nil, Http.WrapError(e, fmt.Sprintf("failed to load page[%d] as record", idx), http.StatusInternalServerError)
		}
		if len(record.Data[KeyActive].([]interface{})) > 0 {
			entry := record.Data[KeyActive].([]interface{})[0].(map[string]interface{})
			return entry, nil
		}
	}
	return nil, nil
}

func (j *JournalLib) GetJournalPage(dataType string, dataId string, idx int) (map[string]interface{}, *Http.HttpError) {
	pageId := PageId(dataType, dataId, idx)
	args := make(map[string]interface{})
	args[DbIface.Table] = j.table
	args[Record.DataType] = KeyJournal
	args[Record.DataId] = pageId
	journalList, err := j.db.Get(args)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	if len(journalList) == 0 {
		return nil, Http.NewHttpError(fmt.Sprintf("journal page=[%s] does not exists.", pageId), http.StatusNotFound)
	}
	if len(journalList) > 1 {
		return nil, Http.NewHttpError(fmt.Sprintf("journal page=[%s] conflict. fount [%d] records.", pageId, len(journalList)), http.StatusInternalServerError)
	}
	return journalList[0], nil
}

func (j *JournalLib) retrieveJournals(dataType string, dataId string) ([]map[string]interface{}, *Http.HttpError) {
	args := make(map[string]interface{})
	args[DbIface.Table] = j.table
	args[Record.DataType] = KeyJournal
	journalList, err := j.db.Get(args)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	if dataType == "" {
		return journalList, nil
	}
	filterResults := []map[string]interface{}{}
	for _, journal := range journalList {
		jDataType := journal[Record.Data].(map[string]interface{})[KeyDataType].(string)
		jDataId := journal[Record.Data].(map[string]interface{})[KeyDataId].(string)
		if dataType == jDataType {
			if dataId == "" || dataId == jDataId {
				filterResults = append(filterResults, journal)
			}
		}
	}
	return filterResults, nil
}

func (j *JournalLib) updateJournal(journalPage *Record.Record) *Http.HttpError {
	e := j.db.Create(j.table, journalPage.Map())
	if e != nil {
		return Http.WrapError(e, fmt.Sprintf("failed to create record [{type}/{id}]=[%s]/%s", journalPage.Type, journalPage.Id), http.StatusInternalServerError)
	}
	return nil
}

func (j *JournalLib) AddJournal(dataType string, dataId string, before map[string]interface{}, after map[string]interface{}) (int, *Http.HttpError) {
	j.lock.Lock()
	defer j.lock.Unlock()
	journalEntry := NewJournalEntry(before, after)
	return j.AddJournalEntry(dataType, dataId, journalEntry)
}

func (j *JournalLib) AddJournalEntry(dataType string, dataId string, journalEntry map[string]interface{}) (int, *Http.HttpError) {
	pageList, err := j.ListJournalPages(dataType, dataId)
	if err != nil {
		return 0, err
	}
	var journalPage *Record.Record
	if len(pageList) == 0 {
		journalPage = NewJournalPage(dataType, dataId, 1)
		journalPage.Data[KeyActive] = []interface{}{journalEntry}
		journalEntry[KeyIdx] = 1
		journalEntry[KeyPage] = 1
	} else {
		lastPage := pageList[len(pageList)-1]
		pageRecord, e := Record.LoadMap(lastPage)
		if e != nil {
			return 0, Http.WrapError(e, fmt.Sprintf("failed to load page as record. [%s], Error:%s", lastPage[Record.DataId], e), http.StatusInternalServerError)
		}
		idx := int(pageRecord.Data[KeyIdx].(float64))
		activeList := pageRecord.Data[KeyActive].([]interface{})
		archiveList := pageRecord.Data[KeyArchived].([]interface{})
		if len(activeList)+len(archiveList) < MaxEntryPerPage {
			journalEntry[KeyIdx] = len(activeList) + len(archiveList) + 1
			journalEntry[KeyPage] = idx
			pageRecord.Data[KeyActive] = append(activeList, journalEntry)
			journalPage = pageRecord
		} else {
			journalPage = NewJournalPage(dataType, dataId, idx+1)
			journalPage.Data[KeyActive] = []interface{}{journalEntry}
			journalEntry[KeyIdx] = 1
			journalEntry[KeyPage] = idx + 1
		}
	}
	err = j.updateJournal(journalPage)
	if err != nil {
		return 0, err
	}
	return journalEntry[KeyPage].(int), nil
}

func (j *JournalLib) ArchiveJournalEntry(dataType string, dataId string, entry map[string]interface{}) *Http.HttpError {
	pageIdx := int(entry[KeyPage].(float64))
	entryIdx := int(entry[KeyIdx].(float64))
	// only move the first active one to archived if it matched.
	journalPage, err := j.GetJournalPage(dataType, dataId, pageIdx)
	if err != nil {
		return err
	}
	journalRecord, e := Record.LoadMap(journalPage)
	if e != nil {
		return Http.WrapError(e, "failed to load journal page data as record.", http.StatusInternalServerError)
	}
	activeList := journalRecord.Data[KeyActive].([]interface{})
	archiveList := journalRecord.Data[KeyArchived].([]interface{})
	firstEntry := activeList[0].(map[string]interface{})
	if int(firstEntry[KeyIdx].(float64)) != entryIdx {
		// the immediate journal entry is not the one to archive. do nothing.
		return nil
	}
	journalRecord.Data[KeyActive] = activeList[1:]
	journalRecord.Data[KeyArchived] = append(archiveList, firstEntry)
	return j.updateJournal(journalRecord)
}
