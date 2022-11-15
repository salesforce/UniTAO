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
	"DataService/DataJournal/ProcessIface"
	"log"
)

type JournalWorker struct {
	lib       *JournalLib
	dataType  string
	dataId    string
	log       *log.Logger
	processes []ProcessIface.JournalProcess
}

func (w *JournalWorker) Run(event chan interface{}) {
	entry := w.lib.NextJournalEntry(w.dataType, w.dataId)
	if entry == nil {
		w.lib.CleanArchivedPages(w.dataType, w.dataId)
		return
	}
	for _, p := range w.processes {
		ex := p.ProcessEntry(w.dataType, w.dataId, entry)
		if ex != nil {
			w.log.Printf("%s: failed to process entry. Error:%s", p.Name(), ex)
			return
		}
	}
	w.lib.ArchiveJournalEntry(w.dataType, w.dataId, entry)
}
