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
	"time"
)

type HashLock struct {
	log  *log.Logger
	lock *ChanLock
	hash map[string]*lockCtrl
}

func NewHashLock(logger *log.Logger) *HashLock {
	if logger == nil {
		logger = log.Default()
	}
	return &HashLock{
		log:  logger,
		lock: NewChanLock(logger),
		hash: map[string]*lockCtrl{},
	}
}

func (l *HashLock) getLockCtrl(key string) *lockCtrl {
	_, ok := l.hash[key]
	if ok {
		return l.hash[key]
	}
	_, ok = l.hash[key]
	if !ok {
		lock := NewLockCtrl(key, l.log)
		l.hash[key] = lock
	}
	return l.hash[key]
}

func (l *HashLock) Aquire(key string, userId string) error {
	l.log.Printf("HashLock: aquire lock on [%s] for [%s]", key, userId)
	timeout := 10 * time.Second
	err := l.lock.Lock(timeout)
	if err != nil {
		l.log.Printf("HashLock: aquire timeout[10 seconds] lock on [%s] for [%s]", key, userId)
		return err
	}
	lockCtrl := l.getLockCtrl(key)
	l.lock.Unlock()
	return lockCtrl.Aquire(userId, timeout)
}

func (l *HashLock) Release(key string, userId string) {
	lc, ok := l.hash[key]
	if !ok {
		l.log.Printf("HashLock: lock key [%s] does not exists", key)
		return
	}
	currentUser := lc.LockedBy()
	if currentUser == userId {
		lc.Release()
	} else {
		l.log.Printf("HashLock: lock user [%s]!= release user [%s]", currentUser, userId)
	}
	l.log.Printf("HashLock: check lock waiters on key[%s]", key)
	l.lock.Lock(10 * time.Second)
	if lc.Count() == 0 {
		l.log.Printf("HashLock: no waiter, remove lock key [%s]", key)
		delete(l.hash, key)
	}
	l.lock.Unlock()
}
