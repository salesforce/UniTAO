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

package Thread

import (
	"sync"
	"syscall"
)

type Worker struct {
	Id      string
	wg      *sync.WaitGroup
	stopped bool
	event   chan interface{}
	run     func(chan interface{})
}

func NewWorker(workerId string, wg *sync.WaitGroup, run func(chan interface{})) *Worker {
	worker := Worker{
		Id:      workerId,
		wg:      wg,
		stopped: true,
		event:   make(chan interface{}),
		run:     run,
	}
	return &worker
}

func (w *Worker) setup() {
	w.stopped = false
	w.wg.Add(1)
}

func (w *Worker) postRun() {
	w.wg.Done()
	w.stopped = true
}

func (w *Worker) workerRoutine() {
	w.setup()
	defer w.postRun()
	w.run(w.event)
}

func (w *Worker) Run() {
	go w.workerRoutine()
}

func (w *Worker) Stopped() bool {
	return w.stopped
}

func (w *Worker) Notify(event interface{}) {
	w.event <- event
}

func (w *Worker) Stop() {
	w.event <- syscall.SIGINT
}
