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
	"DataService/Common"
	"DataService/DataJournal/ProcessIface"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

const (
	CurrentVer      = "0.0.1"
	MaxEntryPerPage = 10
	KeyActive       = "active"
	KeyAfter        = "after"
	KeyArchived     = "archived"
	KeyBefore       = "before"
	KeyId           = "id"
	KeyIdx          = "idx"
	KeyTime         = "time"
)

type JournalLib struct {
	db            DbIface.Database
	table         string
	Cache         map[string]map[string]*JournalCache
	Logger        *log.Logger
	HandlerNotify func(event interface{})
}

func NewJournalLib(db DbIface.Database, table string, logger *log.Logger) (*JournalLib, *Http.HttpError) {
	if logger == nil {
		logger = log.Default()
	}
	lib := JournalLib{
		db:     db,
		table:  table,
		Cache:  map[string]map[string]*JournalCache{},
		Logger: logger,
	}
	err := lib.initCache()
	if err != nil {
		return nil, err
	}
	return &lib, nil
}

func (j *JournalLib) initCache() *Http.HttpError {
	recordList, err := j.QueryJournal("")
	if err != nil {
		return err
	}
	pageMap := map[string]*Record.Record{}
	for _, record := range recordList.([]*Record.Record) {
		dataType, dataId, idx, ex := ProcessIface.ParseJournalId(record.Id)
		if ex != nil {
			return Http.WrapError(ex, fmt.Sprintf("failed to parse journal id=[%s]", record.Id), http.StatusInternalServerError)
		}
		pageMap[record.Id] = record
		if _, ok := j.Cache[dataType]; !ok {
			j.Cache[dataType] = map[string]*JournalCache{}
		}
		if _, ok := j.Cache[dataType][dataId]; !ok {
			j.Cache[dataType][dataId] = NewCache(dataType, dataId, j.Logger)
		}
		cache := j.Cache[dataType][dataId]
		if cache.Head.Idx == -1 || cache.Head.Idx > idx {
			cache.Head.Idx = idx
		}
		if cache.Tail.Idx < idx {
			cache.Tail.Idx = idx
		}
	}
	for dataType := range j.Cache {
		for dataId := range j.Cache[dataType] {
			cache := j.Cache[dataType][dataId]
			headRecord := pageMap[ProcessIface.PageId(dataType, dataId, cache.Head.Idx)]
			cache.Head.LoadMap(headRecord.Data)
			if cache.Head.Idx == cache.Tail.Idx {
				cache.Tail = cache.Head
				continue
			}
			tailRecord := pageMap[ProcessIface.PageId(dataType, dataId, cache.Tail.Idx)]
			cache.Tail.LoadMap(tailRecord.Data)
		}
	}
	return nil
}

func (j *JournalLib) GetJournal(journalId string) (interface{}, *Http.HttpError) {
	dataType, dataId, idx, err := ProcessIface.ParseJournalId(journalId)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusBadRequest)
	}
	if idx < 0 {
		return nil, Http.NewHttpError(fmt.Sprintf("invalid page idx:[%d] in id=[%s], expect 0..., 0 means list all pages", idx, journalId), http.StatusNotFound)
	}
	switch {
	case dataType == "":
		return j.ListJournalTypes(), nil
	case dataId == "":
		return j.ListJournalIds(dataType), nil
	case idx == 0:
		return j.ListJournalPages(dataType, dataId)
	default:
		return j.QueryJournal(journalId)
	}
}

func (j *JournalLib) ListJournalTypes() []string {
	typeList := make([]string, 0, len(j.Cache))
	for name := range j.Cache {
		typeList = append(typeList, name)
	}
	return typeList
}

func (j *JournalLib) ListJournalIds(dataType string) []string {
	if _, ok := j.Cache[dataType]; !ok {
		return []string{}
	}
	idList := make([]string, 0, len(j.Cache[dataType]))
	for id := range j.Cache[dataType] {
		idList = append(idList, id)
	}
	return idList
}

func (j *JournalLib) ListJournalPages(dataType string, dataId string) ([]string, *Http.HttpError) {
	if _, ok := j.Cache[dataType]; !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("journal page of [%s] does not exists", dataType), http.StatusNotFound)
	}
	cache, ok := j.Cache[dataType][dataId]
	if !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("journal page of [%s/%s] does not exists", dataType, dataId), http.StatusNotFound)
	}
	return cache.ListPages(), nil
}

func (j *JournalLib) NextJournalEntry(dataType string, dataId string) *ProcessIface.JournalEntry {
	if _, ok := j.Cache[dataType]; !ok {
		return nil
	}
	cache, ok := j.Cache[dataType][dataId]
	if !ok {
		return nil
	}
	if len(cache.Head.Active) == 0 && cache.Head.Idx < cache.Tail.Idx {
		currentId := cache.Head.Id()
		j.Logger.Printf("[%s]: page[%s] complete, move to next page.", WorkId(cache.DataType, cache.DataId), currentId)
		err := j.JournalHeadForward(cache)
		if err != nil {
			j.Logger.Printf("[%s]: page move forward failed. current page[%s], Error:%s", WorkId(cache.DataType, cache.DataId), cache.Head.Id(), err)
		}
	}
	j.Logger.Printf("[%s] get active entry from page[%s]", WorkId(cache.DataType, cache.DataId), cache.Head.Id())
	if len(cache.Head.Active) == 0 {
		j.Logger.Printf("[%s] no active entry from page[%s]", WorkId(cache.DataType, cache.DataId), cache.Head.Id())
		return nil
	}
	return cache.Head.Active[0]
}

func (j *JournalLib) QueryJournal(journalId string) (interface{}, *Http.HttpError) {
	args := make(map[string]interface{})
	args[DbIface.Table] = j.table
	args[Record.DataType] = Common.KeyJournal
	if journalId != "" {
		args[Record.DataId] = journalId
	}
	dataList, err := j.db.Get(args)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	if journalId != "" {
		if len(dataList) == 0 {
			return nil, Http.NewHttpError(fmt.Sprintf("journal page:[%s] does not exists", journalId), http.StatusNotFound)
		}
		if len(dataList) > 1 {
			return nil, Http.NewHttpError(fmt.Sprintf("found %d pages of id:[%s]", len(dataList), journalId), http.StatusInternalServerError)
		}
		record, err := Record.LoadMap(dataList[0])
		if err != nil {
			return nil, Http.WrapError(err, fmt.Sprintf("journal:[%s] failed to be load as Record", journalId), http.StatusInternalServerError)
		}
		return record, nil
	}
	result := []*Record.Record{}
	for idx, data := range dataList {
		record, err := Record.LoadMap(data)
		if err != nil {
			return nil, Http.WrapError(err, fmt.Sprintf("failed to load %d journal as record", idx), http.StatusInternalServerError)
		}
		result = append(result, record)
	}
	return result, nil
}

func (j *JournalLib) AddJournal(dataType string, dataId string, before map[string]interface{}, after map[string]interface{}) *Http.HttpError {
	if _, ok := j.Cache[dataType]; !ok {
		j.Cache[dataType] = map[string]*JournalCache{}
	}
	if _, ok := j.Cache[dataType][dataId]; !ok {
		c := NewCache(dataType, dataId, j.Logger)
		c.Head = c.Tail
		j.Cache[dataType][dataId] = c
	}
	j.Logger.Printf("AddJournal: [%s/%s] adding Journal", dataType, dataId)
	err := j.addJournalEntry(dataType, dataId, before, after)
	if err != nil {
		j.Logger.Printf("AddJournal: error while addJournalEntry. Error:%s", err)
		return err
	}
	j.Logger.Printf("AddJournal: [%s/%s] journal added ", dataType, dataId)
	if j.HandlerNotify == nil {
		j.Logger.Printf("AddJournal: [%s/%s] no JournalHandler Channel define", dataType, dataId)
		return nil
	}
	j.Logger.Printf("AddJournal: [%s/%s] there is Handler Channel defined, submit event to Journal Handler", dataType, dataId)
	event := ProcessIface.JournalEvent{
		DataType: dataType,
		DataId:   dataId,
	}
	j.HandlerNotify(event)
	j.Logger.Printf("AddJournal: [%s/%s] add journal process completed", dataType, dataId)
	return nil
}

func (j *JournalLib) ArchiveJournalEntry(dataType string, dataId string, entry *ProcessIface.JournalEntry) *Http.HttpError {
	if _, ok := j.Cache[dataType]; !ok {
		j.Logger.Printf("Archive: no journal for type=[%s]", dataType)
		return nil
	}
	if _, ok := j.Cache[dataType][dataId]; !ok {
		j.Logger.Printf("Archive: no journal for data=[%s/%s]", dataType, dataId)
		return nil
	}
	cache := j.Cache[dataType][dataId]
	if cache.Head.Idx != entry.Page {
		j.Logger.Printf("Archive: entry page [%d]!= head page [%d]", entry.Page, cache.Head.Idx)
		return nil
	}
	if len(cache.Head.Active) == 0 {
		j.Logger.Printf("Archive: no entry to archive @[%s]", cache.Head.Id())
		return nil
	}
	if cache.Head.Active[0].Idx != entry.Idx {
		j.Logger.Printf("Archive: entry idx [%d]!= current entry idx [%d]", entry.Idx, cache.Head.Active[0].Idx)
		return nil
	}
	if cache.Head.Idx == cache.Tail.Idx {
		j.Logger.Printf("Archive: acquire Lock for [%s/%s]", dataType, dataId)
		cache.Lock.Lock(10 * time.Second)
		defer cache.Lock.Unlock()
	}
	cache.Head.Active = cache.Head.Active[1:]
	cache.Head.Archived = append(cache.Head.Archived, entry)
	if len(cache.Head.Active) > 0 || cache.Head.Idx == cache.Tail.Idx {
		if len(cache.Head.Active) > 0 {
			j.Logger.Printf("Archive[%s]: still [%d] active entries", WorkId(dataType, dataId), len(cache.Head.Active))
		} else {
			j.Logger.Printf("Archive[%s]: no more entries, wait for new", WorkId(dataType, dataId))
		}
		return j.updateJournal(cache.Head)
	}
	return j.JournalHeadForward(cache)
}

func (j *JournalLib) JournalHeadForward(cache *JournalCache) *Http.HttpError {
	j.Logger.Printf("Archive[%s]: Page[%s] done, move HEAD forward", WorkId(cache.DataType, cache.DataId), cache.Head.Id())
	err := j.removeJournal(cache.Head.Id())
	if err != nil {
		return err
	}
	nextHeadIdx := cache.Head.Idx + 1
	for nextHeadIdx < cache.Tail.Idx {
		nextId := ProcessIface.PageId(cache.DataType, cache.DataId, nextHeadIdx)
		j.Logger.Printf("Archive[%s]: Query next Journal Page: [%s]", WorkId(cache.DataType, cache.DataId), nextId)
		data, err := j.QueryJournal(nextId)
		if err != nil {
			if err.Status == http.StatusNotFound {
				nextHeadIdx += 1
				j.Logger.Printf("Archive[%s]: journal page [%s] does not exists. skip: %d", WorkId(cache.DataType, cache.DataId), nextId, nextHeadIdx)
				continue
			}
			return err
		}
		cache.Head = ProcessIface.NewPage(cache.DataType, cache.DataId, nextHeadIdx)
		ex := cache.Head.LoadMap(data.(*Record.Record).Data)
		if ex != nil {
			return Http.WrapError(ex, fmt.Sprintf("failed to load journalPage from record [%s/%s]", Common.KeyJournal, nextId), http.StatusInternalServerError)
		}
		j.Logger.Printf("Archive[%s]: set Head Journal to [%s]", WorkId(cache.DataType, cache.DataId), cache.Head.Id())
		return nil
	}
	j.Logger.Printf("Archive[%s]: Processing Last Page.[%s]", WorkId(cache.DataType, cache.DataId), cache.Tail.Id())
	cache.Head = cache.Tail
	return nil
}

func (j *JournalLib) removeJournal(journalId string) *Http.HttpError {
	keys := make(map[string]interface{})
	keys[Record.DataType] = Common.KeyJournal
	keys[Record.DataId] = journalId
	ex := j.db.Delete(j.table, keys)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to delete record [type/id]=[%s/%s]", Common.KeyJournal, journalId), http.StatusInternalServerError)
	}
	return nil
}

func (j *JournalLib) updateJournal(page *ProcessIface.JournalPage) *Http.HttpError {
	pageData := map[string]interface{}{}
	err := Json.CopyTo(page, &pageData)
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to create Record Data. Error:%s", err), http.StatusBadRequest)
	}
	pageRecord := Record.NewRecord(Common.KeyJournal, CurrentVer, page.Id(), pageData)
	err = j.db.Replace(j.table, map[string]interface{}{
		Record.DataType: Common.KeyJournal,
		Record.DataId:   page.Id(),
	}, pageRecord.Map())
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to create record [{type}/{id}]=[%s/%s]", pageRecord.Type, pageRecord.Id), http.StatusInternalServerError)
	}
	return nil
}

func (j *JournalLib) addJournalEntry(dataType string, dataId string, before map[string]interface{}, after map[string]interface{}) *Http.HttpError {
	cache := j.Cache[dataType][dataId]
	j.Logger.Printf("AddJournal: acquire lock for [%s/%s]", dataType, dataId)
	ex := cache.Lock.Lock(10 * time.Second)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("AddJournal: lock acquire time out after 10 seconds. [%s/%s]", dataType, dataId), http.StatusInternalServerError)
	}
	defer cache.Lock.Unlock()
	defer j.Logger.Printf("AddJournal: lock for [%s/%s] released", dataType, dataId)
	if j.Cache[dataType][dataId].Tail.LastEntry() >= MaxEntryPerPage {
		nextIdx := j.Cache[dataType][dataId].Tail.Idx + 1
		j.Cache[dataType][dataId].Tail = ProcessIface.NewPage(dataType, dataId, nextIdx)
		j.Logger.Printf("[%s]: create new journal page[%s]", WorkId(dataType, dataId), j.Cache[dataType][dataId].Tail.Id())
		j.Logger.Printf("[%s]: head page[%s]", WorkId(dataType, dataId), j.Cache[dataType][dataId].Head.Id())
	}
	tail := j.Cache[dataType][dataId].Tail
	if tail.Idx == -1 {
		tail.Idx = 1
	}
	entryIdx := tail.LastEntry() + 1
	j.Logger.Printf("[%s]: add Journal[%d] to page %d", WorkId(dataType, dataId), entryIdx, tail.Idx)
	entry := ProcessIface.JournalEntry{
		Time:   time.Now().String(),
		Page:   tail.Idx,
		Idx:    entryIdx,
		Before: before,
		After:  after,
	}
	tail.Active = append(tail.Active, &entry)
	err := j.updateJournal(tail)
	if err != nil {
		j.Logger.Printf("[%s]: update Journal page [%s] error:%s", WorkId(dataType, dataId), tail.Id(), err)
		return err
	}
	j.Logger.Printf("[%s]: update Journal page [%s] saved", WorkId(dataType, dataId), tail.Id())
	return nil
}
