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

import "github.com/salesforce/UniTAO/lib/Schema/JsonKey"

type JournalWorker struct {
	lib      *JournalLib
	dataType string
	dataId   string
}

func (w *JournalWorker) Run(event chan interface{}) {
	entry, err := w.lib.NextJournalEntry(w.dataType, w.dataId)
	if err != nil {
		// TODO: log error and exit without throw anything.
		return
	}
	if entry == nil {
		w.lib.CleanArchivedPages(w.dataType, w.dataId)
		return
	}
	w.ProcessJournalEntry(entry)
}

func (w *JournalWorker) ProcessJournalEntry(entry map[string]interface{}) {
	if w.dataType == JsonKey.Schema {
		// data change on schema. trace CmtSubscription
		// undo old subscription and redo new ones
		w.ProcessSchemaChange(entry)
		return
	}
	w.ProcessDataChange(entry)
}

func (w *JournalWorker) ProcessSchemaChange(entry map[string]interface{}) {
	before, beforeOk := entry[KeyBefore]
	after, afterOk := entry[KeyAfter]
	if beforeOk && afterOk {
		// log error of schema change not allowed
		return
	}
	if afterOk {
		err := w.SubScribeCmtChanges(after.(map[string]interface{}))
		if err != nil {
			return
		}
	}
	if beforeOk {
		err := w.UnSubscribeCMTChanges(before.(map[string]interface{}))
		if err != nil {
			return
		}
	}
	// log empty entry of schema change
	w.lib.ArchiveJournalEntry(w.dataType, w.dataId, entry)
}

func (w *JournalWorker) UnSubscribeCMTChanges(schema map[string]interface{}) error {
	return nil
}

func (w *JournalWorker) SubScribeCmtChanges(schema map[string]interface{}) error {
	return nil
}

func (w *JournalWorker) ProcessDataChange(entry map[string]interface{}) error {
	return nil
}
