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

func ProcessNextEntry(lib *JournalLib, processes []ProcessIface.JournalProcess, dataType string, dataId string, log *log.Logger) *Http.HttpError {
	entry := lib.NextJournalEntry(dataType, dataId)
	if entry == nil {
		err := lib.CleanArchivedPages(dataType, dataId)
		if err != nil {
			return err
		}
		return Http.NewHttpError(fmt.Sprintf("failed to get journal of %s/%s", dataType, dataId), http.StatusNotFound)
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
	for _, p := range processes {
		canHandle := false
		for _, ver := range versionList {
			can, err := p.HandleType(dataType, ver)
			if err != nil {
				return Http.WrapError(err, fmt.Sprintf("not able to determine [%s] can handle type=[%s] & ver=[%s]", p.Name(), dataType, ver), http.StatusInternalServerError)
			}
			if can {
				canHandle = can
				break
			}
		}
		if canHandle {
			err := p.ProcessEntry(dataType, dataId, entry)
			if err != nil {
				if err.Status != http.StatusNotModified {
					return Http.WrapError(err, fmt.Sprintf("%s: failed to process entry", p.Name()), http.StatusInternalServerError)
				}
				log.Printf("%s: entry processed [%d-%d] with no modification.\n Message: %s", p.Name(), entry.Page, entry.Idx, err)
			}
		}
	}
	err := lib.ArchiveJournalEntry(dataType, dataId, entry)
	if err != nil {
		return err
	}
	return nil
}

func (w *JournalWorker) Run(notify chan interface{}) {
	for {
		select {
		case event := <-notify:
			signal, ok := event.(os.Signal)
			if ok && signal == syscall.SIGINT {
				return
			}
		default:
			err := ProcessNextEntry(w.lib, w.processes, w.dataType, w.dataId, w.log)
			if err != nil {
				if err.Status != http.StatusNotFound {
					w.log.Print(err)
				}
				w.log.Printf("no more entry for %s/%s, sleep for 5 seconds", w.dataType, w.dataId)
				time.Sleep(5 * time.Second)
			}
		}
	}
}
