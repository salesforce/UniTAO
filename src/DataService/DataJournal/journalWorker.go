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
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type JournalWorker struct {
	lib       *JournalLib
	dataType  string
	dataId    string
	log       *log.Logger
	processes []ProcessIface.JournalProcess
}

func NewJournalWorker(lib *JournalLib, dataType string, dataId string, logger *log.Logger, processes []ProcessIface.JournalProcess) *JournalWorker {
	if logger == nil {
		logger = log.Default()
	}
	if processes == nil {
		processes = []ProcessIface.JournalProcess{}
	}
	return &JournalWorker{
		lib:       lib,
		dataType:  dataType,
		dataId:    dataId,
		log:       logger,
		processes: processes,
	}
}

func (w *JournalWorker) ProcessNextEntry() *Http.HttpError {
	entry := w.lib.NextJournalEntry(w.dataType, w.dataId)
	if entry == nil {
		errMsg := "no new journal"
		w.Log(errMsg)
		return Http.NewHttpError(errMsg, http.StatusNotFound)
	}
	versionList := make([]string, 0, 2)
	if entry.Before != nil {
		versionList = append(versionList, entry.Before[Record.Version].(string))
	}
	if entry.After != nil {
		afterVer := entry.After[Record.Version].(string)
		if len(versionList) == 0 || versionList[0] != afterVer {
			versionList = append(versionList, afterVer)
		}
	}
	for _, p := range w.processes {
		canHandle := false
		for _, ver := range versionList {
			can, err := p.HandleType(w.dataType, ver)
			if err != nil {
				return Http.WrapError(err, fmt.Sprintf("not able to determine [%s] can handle type=[%s] & ver=[%s]", p.Name(), w.dataType, ver), http.StatusInternalServerError)
			}
			if can {
				canHandle = can
				break
			}
		}
		if canHandle {
			err := p.ProcessEntry(w.dataType, w.dataId, entry)
			if err != nil {
				if err.Status != http.StatusNotModified {
					return Http.WrapError(err, fmt.Sprintf("%s: failed to process entry", p.Name()), http.StatusInternalServerError)
				}
				w.Log(fmt.Sprintf("%s: entry processed [%d-%d] with no modification.\n Message: %s", p.Name(), entry.Page, entry.Idx, err))
			}
		}
	}
	err := w.lib.ArchiveJournalEntry(w.dataType, w.dataId, entry)
	if err != nil {
		w.Log(fmt.Sprintf("failed to archive Journal entry [%d] @[%s]", entry.Idx, ProcessIface.PageId(w.dataType, w.dataId, entry.Page)))
		return err
	}
	return nil
}

func (w *JournalWorker) Log(message string) {
	w.log.Printf("JournalWorker[%s/%s]: %s", w.dataType, w.dataId, message)
}

func (w *JournalWorker) Run(notify chan interface{}) error {
	interval := 0
	for {
		select {
		case event := <-notify:
			w.Log("got an event")
			signal, ok := event.(os.Signal)
			if ok && signal == syscall.SIGINT {
				w.Log("exit event")
				return nil
			}
			_, isJournalEvent := event.(ProcessIface.JournalEvent)
			if isJournalEvent {
				w.Log(fmt.Sprintf("got event of new journal: [%s]", WorkId(w.dataType, w.dataId)))
				interval = w.RunAndCalcNextSleep(0)
			}
		case <-time.After(time.Duration(interval) * time.Second):
			interval = w.RunAndCalcNextSleep(interval)
		}
		if interval > 0 {
			w.Log(fmt.Sprintf("sleep %d seconds", interval))
		}
	}
}

func (w *JournalWorker) RunAndCalcNextSleep(interval int) int {
	w.Log("process next entry")
	err := w.ProcessNextEntry()
	if err == nil {
		// no error, means no sleep. keep going
		return 0
	}
	if err.Status != http.StatusNotFound {
		// error, sleep minimum time
		w.Log(fmt.Sprintf("process error: %s", err))
		return intervalStep
	}
	w.Log("no more entries")
	if interval < maxInterval {
		// no entry, step up interval
		return interval + intervalStep
	}
	return interval
}
