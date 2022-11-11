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

package SchemaPath

import (
	"fmt"
	"net/http"

	"github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

const (
	QueryPath      = "queryPath"
	QueryResults   = "queryResults"
	QueryData      = "data"
	QueryIterators = "iterators"
)

type QueryResult struct {
	Data      interface{} `json:"data"`
	Iterators []string    `json:"iterators"`
}

type IteratorResult struct {
	Path         string        `json:"queryPath"`
	QueryResults []QueryResult `json:"queryResults"`
}

type CmdPathIterator struct {
	path string
	p    *Node.PathNode
}

func NewIteratorQuery(conn *Data.Connection, dataType string, dataId string, path string) (*CmdPathIterator, *Http.HttpError) {
	node, err := BuildNodePath(conn, dataType, dataId, path)
	if err != nil {
		return nil, err
	}
	qPath := dataId
	if path != "" {
		qPath = fmt.Sprintf("%s/%s", qPath, path)
	}
	return &CmdPathIterator{
		path: qPath,
		p:    node,
	}, nil
}

func (c *CmdPathIterator) Name() string {
	return PathCmd.CmdIter
}

func (c *CmdPathIterator) WalkValue() (interface{}, *Http.HttpError) {
	/*
		TODO:
		walk as the path but only record the * Idx on Array or Map.
		record the format as following:

	*/
	queryResult, err := c.GetNodeValue(c.p)
	if err != nil {
		return nil, err
	}
	iterResult := IteratorResult{
		Path:         c.path,
		QueryResults: queryResult,
	}
	result, cErr := Util.JsonCopy(iterResult)
	if cErr != nil {
		return nil, Http.WrapError(cErr, "failed to convert struct [IteratorResult] to json", http.StatusInternalServerError)
	}
	return result, nil
}

func (c *CmdPathIterator) GetNodeValue(node *Node.PathNode) ([]QueryResult, *Http.HttpError) {
	if len(node.Next) > 0 {
		resultList := []QueryResult{}
		for _, next := range node.Next {
			nextList, err := c.GetNodeValue(next)
			if err != nil {
				return nil, err
			}
			for _, result := range nextList {
				if node.Idx == Node.All {
					result.Iterators = append([]string{next.Idx}, result.Iterators...)
				}
				resultList = append(resultList, result)
			}
		}
		return resultList, nil
	}
	return []QueryResult{QueryResult{Data: node.Data, Iterators: []string{}}}, nil
}
