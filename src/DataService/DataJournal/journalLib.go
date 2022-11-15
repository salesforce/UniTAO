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
	"DataService/DataJournal/ProcessIface"
	"fmt"
	"net/http"
	"sort"
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

type JournalCache struct {
	Sort  []int
	Pages map[int]*ProcessIface.JournalPage
}

func NewCache() *JournalCache {
	return &JournalCache{
		Sort:  []int{},
		Pages: map[int]*ProcessIface.JournalPage{},
	}
}

type JournalLib struct {
	db    DbIface.Database
	table string
	Cache map[string]map[string]*JournalCache
	lock  sync.Mutex
}

func NewJournalLib(db DbIface.Database, table string) (*JournalLib, *Http.HttpError) {
	lib := JournalLib{
		db:    db,
		table: table,
		Cache: map[string]map[string]*JournalCache{},
	}
	err := lib.retrieveJournals()
	if err != nil {
		return nil, err
	}
	return &lib, nil
}

func PageId(dataType string, dataId string, idx int) string {
	pageId := fmt.Sprintf("%s:%s_%s:%s_%s:%d", KeyDataType, dataType, KeyDataId, dataId, KeyPage, idx)
	return pageId
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
		return j.ListJournalTypes(), nil
	case dataId == "":
		return j.ListJournalIds(dataType), nil
	case idx == 0:
		return j.ListJournalPages(dataType, dataId), nil
	default:
		return j.GetJournalPage(dataType, dataId, idx)
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

func (j *JournalLib) ListJournalPages(dataType string, dataId string) []*ProcessIface.JournalPage {
	if _, ok := j.Cache[dataType]; !ok {
		return []*ProcessIface.JournalPage{}
	}

	if _, ok := j.Cache[dataType][dataId]; !ok {
		return []*ProcessIface.JournalPage{}
	}
	pageList := make([]*ProcessIface.JournalPage, 0, len(j.Cache[dataType][dataId].Sort))
	for _, idx := range j.Cache[dataType][dataId].Sort {
		page := j.Cache[dataType][dataId].Pages[idx]
		pageList = append(pageList, page)
	}
	return pageList
}

func (j *JournalLib) CleanArchivedPages(dataType string, dataId string) *Http.HttpError {
	j.lock.Lock()
	defer j.lock.Unlock()
	err := j.retrieveJournals()
	if err != nil {
		return err
	}
	if _, ok := j.Cache[dataType]; !ok {
		return nil
	}
	cache, ok := j.Cache[dataType][dataId]
	if !ok {
		return nil
	}
	for idx, pageId := range cache.Sort {
		page := cache.Pages[pageId]
		if page.LastEntry >= MaxEntryPerPage && len(page.Active) == 0 {
			j.removeJournal(page)
			delete(cache.Pages, pageId)
			continue
		}
		cache.Sort = cache.Sort[idx:]
		return nil
	}
	// all page removed
	cache.Sort = []int{}
	return nil
}

func (j *JournalLib) NextJournalEntry(dataType string, dataId string) *ProcessIface.JournalEntry {
	if _, ok := j.Cache[dataType]; !ok {
		return nil
	}
	if _, ok := j.Cache[dataType][dataId]; !ok {
		return nil
	}
	if len(j.Cache[dataType][dataId].Sort) == 0 {
		return nil
	}
	for _, pageId := range j.Cache[dataType][dataId].Sort {
		page := j.Cache[dataType][dataId].Pages[pageId]
		if len(page.Active) > 0 {
			return page.Active[0]
		}
	}
	return nil
}

func (j *JournalLib) GetJournalPage(dataType string, dataId string, idx int) (*ProcessIface.JournalPage, *Http.HttpError) {
	if _, ok := j.Cache[dataType]; !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("journal page, type=[%s] does not exists.", dataType), http.StatusNotFound)
	}
	if _, ok := j.Cache[dataType][dataId]; !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("journal page, type=[%s] id=[%s] does not exists.", dataType, dataId), http.StatusNotFound)
	}
	page, ok := j.Cache[dataType][dataId].Pages[idx]
	if !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("journal page, type=[%s] id=[%s] idx=[%d] does not exists.", dataType, dataId, idx), http.StatusNotFound)
	}
	return page, nil
}

func (j *JournalLib) AddJournal(dataType string, dataId string, before map[string]interface{}, after map[string]interface{}) *Http.HttpError {
	j.lock.Lock()
	defer j.lock.Unlock()
	err := j.retrieveJournals()
	if err != nil {
		return err
	}
	err = j.addJournalEntry(dataType, dataId, before, after)
	if err != nil {
		return err
	}
	return nil
}

func (j *JournalLib) ArchiveJournalEntry(dataType string, dataId string, entry *ProcessIface.JournalEntry) *Http.HttpError {
	j.lock.Lock()
	defer j.lock.Unlock()
	err := j.retrieveJournals()
	if err != nil {
		return err
	}
	if _, ok := j.Cache[dataType]; !ok {
		return Http.NewHttpError(fmt.Sprintf("entry type=[%s] not found", dataType), http.StatusBadRequest)
	}
	if _, ok := j.Cache[dataType][dataId]; !ok {
		return Http.NewHttpError(fmt.Sprintf("entry type=[%s], id=[%s] not found", dataType, dataId), http.StatusBadRequest)
	}
	if _, ok := j.Cache[dataType][dataId].Pages[entry.Page]; !ok {
		return Http.NewHttpError(fmt.Sprintf("entry type=[%s], id=[%s], idx=[%d] not found", dataType, dataId, entry.Page), http.StatusBadRequest)
	}
	page := j.Cache[dataType][dataId].Pages[entry.Page]
	if len(page.Active) == 0 || page.Active[0].Idx != entry.Idx {
		return Http.NewHttpError("entry not active", http.StatusInternalServerError)
	}
	page.Active = page.Active[1:]
	page.Archived = append([]*ProcessIface.JournalEntry{entry}, page.Archived...)
	return j.updateJournal(page)
}

func (j *JournalLib) removeJournal(page *ProcessIface.JournalPage) *Http.HttpError {
	recId := PageId(page.DataType, page.DataId, page.Idx)
	keys := make(map[string]interface{})
	keys[Record.DataType] = KeyJournal
	keys[Record.DataId] = recId
	ex := j.db.Delete(j.table, keys)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to delete record [type/id]=[%s/%s]", KeyJournal, recId), http.StatusInternalServerError)
	}
	return nil
}

func (j *JournalLib) retrieveJournals() *Http.HttpError {
	args := make(map[string]interface{})
	args[DbIface.Table] = j.table
	args[Record.DataType] = KeyJournal
	journalList, err := j.db.Get(args)
	if err != nil {
		return Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	for _, page := range journalList {
		record, err := Record.LoadMap(page)
		if err != nil {
			return Http.WrapError(err, fmt.Sprintf("failed to load data as record. Error: %s", err), http.StatusInternalServerError)
		}
		journal := ProcessIface.JournalPage{}
		err = Util.ObjCopy(record.Data, &journal)
		if err != nil {
			return Http.WrapError(err, fmt.Sprintf("failed to load data as journal page. Error:%s", err), http.StatusInternalServerError)
		}
	}
	return nil
}

func (j *JournalLib) addPageToCache(journal *ProcessIface.JournalPage) {
	if _, ok := j.Cache[journal.DataType]; !ok {
		j.Cache[journal.DataType] = map[string]*JournalCache{}
	}
	typeMap := j.Cache[journal.DataType]
	if _, ok := typeMap[journal.DataId]; !ok {
		typeMap[journal.DataId] = NewCache()
	}
	cache := typeMap[journal.DataId]
	if _, ok := cache.Pages[journal.Idx]; !ok {
		cache.Pages[journal.Idx] = journal
		cache.Sort = append(cache.Sort, journal.Idx)
		sort.Ints(cache.Sort)
	}
}

func (j *JournalLib) updateJournal(page *ProcessIface.JournalPage) *Http.HttpError {
	pageData := map[string]interface{}{}
	err := Util.ObjCopy(page, &pageData)
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to create Record Data. Error:%s", err), http.StatusBadRequest)
	}
	pageId := PageId(page.DataType, page.DataId, page.Idx)
	pageRecord := Record.NewRecord(KeyJournal, CurrentVer, pageId, pageData)
	err = j.db.Create(j.table, pageRecord.Map())
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to create record [{type}/{id}]=[%s]/%s", pageRecord.Type, pageRecord.Id), http.StatusInternalServerError)
	}
	return nil
}

func (j *JournalLib) newJournalPage(dataType string, dataId string, idx int) *ProcessIface.JournalPage {
	page := ProcessIface.JournalPage{
		DataType:  dataType,
		DataId:    dataId,
		Idx:       idx,
		LastEntry: 0,
		Active:    []*ProcessIface.JournalEntry{},
		Archived:  []*ProcessIface.JournalEntry{},
	}
	j.addPageToCache(&page)
	return &page
}

func (j *JournalLib) addJournalEntry(dataType string, dataId string, before map[string]interface{}, after map[string]interface{}) *Http.HttpError {
	if _, ok := j.Cache[dataType]; !ok {
		j.Cache[dataType] = map[string]*JournalCache{}
	}
	if _, ok := j.Cache[dataType][dataId]; !ok {
		j.Cache[dataType][dataId] = NewCache()
	}
	cache := j.Cache[dataType][dataId]
	var page *ProcessIface.JournalPage
	if len(cache.Pages) == 0 {
		page = j.newJournalPage(dataType, dataId, 0)
	} else {
		lastPageIdx := cache.Sort[len(cache.Sort)-1]
		page = cache.Pages[lastPageIdx]
		if page.LastEntry >= MaxEntryPerPage {
			page = j.newJournalPage(dataType, dataId, lastPageIdx+1)
		}
	}
	entry := ProcessIface.JournalEntry{
		Time:   time.Now().String(),
		Page:   page.Idx,
		Idx:    page.LastEntry + 1,
		Before: before,
		After:  after,
	}
	page.Active = append(page.Active, &entry)
	page.LastEntry = entry.Idx
	err := j.updateJournal(page)
	if err != nil {
		return err
	}
	return nil
}
