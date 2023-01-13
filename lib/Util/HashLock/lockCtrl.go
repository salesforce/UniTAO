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

package HashLock

import (
	"log"
	"sync/atomic"
	"time"
)

type lockCtrl struct {
	log    *log.Logger
	key    string
	lock   *ChanLock
	count  int32
	userId string
}

func NewLockCtrl(key string, logger *log.Logger) *lockCtrl {
	if logger == nil {
		logger = log.Default()
	}
	return &lockCtrl{
		log:    logger,
		key:    key,
		lock:   NewChanLock(logger),
		count:  0,
		userId: "",
	}
}

func (lc *lockCtrl) inc() {
	atomic.AddInt32(&lc.count, 1)
}

func (lc *lockCtrl) dec() {
	atomic.AddInt32(&lc.count, -1)
}

func (lc *lockCtrl) Key() string {
	return lc.key
}

func (lc *lockCtrl) Count() int32 {
	return lc.count
}

func (lc *lockCtrl) LockedBy() string {
	return lc.userId
}

func (lc *lockCtrl) Aquire(userId string, timeout time.Duration) error {
	lc.log.Printf("LockCtrl: acquire lock on [%s] for [%s], current waiting [%d]", lc.key, userId, lc.count)
	lc.inc()
	err := lc.lock.Lock(timeout)
	lc.dec()
	if err != nil {
		return err
	}
	lc.userId = userId
	return nil
}

func (lc *lockCtrl) Release() {
	lc.log.Printf("LockCtrl: release lock on [%s] for [%s], current waiting [%d]", lc.key, lc.userId, lc.count)
	lc.userId = ""
	lc.lock.Unlock()
}
