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

package DataLock

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Request struct {
	UserId   string
	DataPath string
	Length   int
}

type LockHash struct {
	mu            sync.Mutex
	Hash          map[string]*Lock
	MaxLockLength int
}

func NewHash(maxLockTime int) *LockHash {
	hash := LockHash{
		Hash:          map[string]*Lock{},
		MaxLockLength: maxLockTime,
	}
	return &hash
}

func (l *LockHash) NewRequest(userId string, dataPath string, length int) *Request {
	if userId == "" {
		userId = uuid.NewString()
	}
	if length <= 0 {
		length = l.MaxLockLength
	}
	r := Request{
		UserId:   userId,
		DataPath: dataPath,
		Length:   length,
	}
	return &r
}

func (l *LockHash) QueryLock(userId string, pathList []string) (*Lock, error) {
	nextPath := pathList[2:]
	var result *Lock
	for _, lock := range l.Hash {
		if lock.DataType != pathList[0] || lock.DataId != pathList[1] {
			continue
		}
		if ListMatch(lock.DataPath, nextPath) {
			if lock.OwnerId == userId {
				result = lock
				continue
			}
			return nil, fmt.Errorf("found lock [%s] conflict with request", lock.Handle)
		}
	}
	return result, nil
}

// return handle of new lock if succeeded.
// or error if not able to create the lock
func (l *LockHash) Lock(r *Request) (string, error) {
	pathList := ParsePathList(r.DataPath)
	if len(pathList) < 2 {
		return "", fmt.Errorf("invalid path to lock. [%s] not match [{dataType}/{dataId}/{additionalPath}]", r.DataPath)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	lock, err := l.QueryLock(r.UserId, pathList)
	if err != nil {
		return "", err
	}
	if lock == nil {
		lock = NewLock(r.UserId, r.DataPath, r.Length)
		l.Hash[lock.Handle] = lock
		return lock.Handle, nil
	}
	if len(lock.DataPath) > len(pathList[2:]) {
		lock.DataPath = pathList[2:]
		lock.Path = r.DataPath
	}
	lock.ExpireTime = time.Now().Unix() + int64(r.Length)
	return lock.Handle, nil
}

func (l *LockHash) UnLock(handle string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.Hash[handle]
	if !ok {
		return
	}
	delete(l.Hash, handle)
}
