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

package SchemaDoc

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
)

type SchemaDoc struct {
	Id          string
	Parent      *SchemaDoc
	Data        map[string]interface{}
	Definitions map[string]*SchemaDoc
	CmtRefs     []*SchemaDocRef
	SubDocs     map[string]*SchemaDoc
}

type SchemaDocRef struct {
	Doc         *SchemaDoc
	Name        string
	ContentType string
}

func NewSchemaDoc(data map[string]interface{}, id string, parent *SchemaDoc) (*SchemaDoc, error) {
	parentPath := ""
	if parent != nil {
		parentPath = parent.Path()
	}
	_, ok := data[JsonKey.Properties].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to create doc path=[%s/%s]", parentPath, id)
	}
	doc := SchemaDoc{
		Id:      id,
		Parent:  parent,
		Data:    data,
		CmtRefs: []*SchemaDocRef{},
		SubDocs: map[string]*SchemaDoc{},
	}
	docDefs, ok := data[JsonKey.Definitions]
	if ok {
		defMap, ok := docDefs.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to parse field=[%s], path=[%s/%s]", JsonKey.Definitions, parentPath, id)
		}
		doc.Definitions = make(map[string]*SchemaDoc, len(defMap))
		for key, def := range defMap {
			defObj, ok := def.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("failed to parse definition to object. path=[%s/%s/%s/%s]", parentPath, id, JsonKey.Definitions, key)
			}
			defDoc, err := NewSchemaDoc(defObj, key, &doc)
			if err != nil {
				return nil, err
			}
			doc.Definitions[key] = defDoc
		}
	}
	err := doc.preprocess()
	if err != nil {
		return nil, fmt.Errorf("failed @preprocess, Err:\n%s", err)
	}
	return &doc, nil
}

func (d *SchemaDoc) Path() string {
	if d.Parent == nil {
		return d.Id
	}
	return path.Join(d.Parent.Path(), d.Id)
}

func (d *SchemaDoc) preprocess() error {
	err := d.processRequired()
	if err != nil {
		err := fmt.Errorf("preprocess failed @processRequired, [path]=[%s], Error:%s", d.Path(), err)
		return err
	}
	err = d.processMap()
	if err != nil {
		err := fmt.Errorf("preprocess failed @processRequired, [path]=[%s], Error:%s", d.Path(), err)
		return err
	}
	err = d.processCmtRefs()
	if err != nil {
		err := fmt.Errorf("preprocess failed @processInvRefs, [path]=[%s], Error:%s", d.Path(), err)
		return err
	}
	if d.Definitions != nil {
		for _, defDoc := range d.Definitions {
			err = defDoc.preprocess()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *SchemaDoc) processRequired() error {
	propPath := path.Join(d.Path(), JsonKey.Properties)
	propMap := d.Data[JsonKey.Properties].(map[string]interface{})
	requiredList := make([]string, 0, len(propMap))
	for pname, prop := range propMap {
		propDef, ok := prop.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid property value !=[map[string]interface{}], path:[%s/%s]", propPath, pname)
		}
		required, ok := propDef[JsonKey.Required]
		if !ok {
			requiredList = append(requiredList, pname)
			continue
		}
		requiredBool, ok := required.(bool)
		if !ok {
			return fmt.Errorf("invalid data type, [type] != [bool], [path]=[%s/%s/%s]", propPath, pname, JsonKey.Required)
		}
		if requiredBool {
			requiredList = append(requiredList, pname)
		}
		delete(propDef, JsonKey.Required)
	}
	if len(requiredList) > 0 {
		d.Data[JsonKey.Required] = requiredList
	}
	return nil
}

// add new custom type=[map], to represent a hash
// JSONSchema definition for map is confusing.
// here we want to use type=[map] and items=hash valud definition for easy understanding
func (d *SchemaDoc) processMap() error {
	for _, prop := range d.Data[JsonKey.Properties].(map[string]interface{}) {
		propDef := prop.(map[string]interface{})
		if propDef[JsonKey.Type] == JsonKey.Map {
			propDef[JsonKey.Type] = JsonKey.Object
			itemDef, ok := propDef[JsonKey.Items].(map[string]interface{})
			if !ok {
				// when no items defined in map, then freeform hash
				propDef[JsonKey.AdditionalProperties] = true
				continue
			}
			propDef[JsonKey.AdditionalProperties] = itemDef
		}
	}
	return nil
}

func (d *SchemaDoc) processCmtRefs() error {
	for pname, prop := range d.Data[JsonKey.Properties].(map[string]interface{}) {
		propDef := prop.(map[string]interface{})
		switch propDef[JsonKey.Type].(string) {
		case JsonKey.Array:
			if itemDef, ok := propDef[JsonKey.Items].(map[string]interface{}); ok {
				err := d.getCmtRef(pname, itemDef)
				if err != nil {
					return err
				}
			}
		case JsonKey.Object:
			if addProps, ok := propDef[JsonKey.AdditionalProperties]; ok {
				// if additionalProperties exists, then collect CMT from it
				if reflect.TypeOf(addProps).Kind() == reflect.Bool {
					// if additionalProperties is a bool, then this is a free style map
					continue
				}
				itemDef := addProps.(map[string]interface{})
				err := d.getCmtRef(pname, itemDef)
				if err != nil {
					return err
				}
				continue
			}
			err := d.getCmtRef(pname, propDef)
			if err != nil {
				return nil
			}
		case JsonKey.String:
			err := d.getCmtRef(pname, propDef)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *SchemaDoc) getCmtRef(pname string, prop map[string]interface{}) error {
	cmt, ok := prop[JsonKey.ContentMediaType]
	if !ok {
		return nil
	}
	ref := SchemaDocRef{
		Doc:         d,
		Name:        pname,
		ContentType: cmt.(string),
	}
	d.CmtRefs = append(d.CmtRefs, &ref)
	return nil
}

// get referenced sub definition
func (d *SchemaDoc) GetRefDoc(pname string, prop map[string]interface{}) (*SchemaDoc, error) {
	switch prop[JsonKey.Type].(string) {
	case JsonKey.Array:
		itemDef, ok := prop[JsonKey.Items].(map[string]interface{})
		if !ok {
			// no item defitinition for array
			return nil, nil
		}
		return d.GetRefDoc(pname, itemDef)
	case JsonKey.Object:
		if addProps, ok := prop[JsonKey.AdditionalProperties]; ok {
			// if additionalProperties exists, then collect CMT from it
			if reflect.TypeOf(addProps).Kind() == reflect.Bool {
				// if additionalProperties is a bool, then this is a free style map
				return nil, nil
			}
			return d.GetRefDoc(pname, addProps.(map[string]interface{}))
		}
	default:
		return nil, nil
	}
	ref, ok := prop[JsonKey.Ref].(string)
	if !ok {
		return nil, nil
	}
	if !strings.HasPrefix(ref, JsonKey.DefinitionPrefix) {
		return nil, fmt.Errorf("unknown ref value=[%s], path=[%s/%s/%s]", ref, d.Path(), pname, JsonKey.Ref)
	}
	docType := ref[len(JsonKey.DefinitionPrefix):]
	doc, err := d.getDefinition(docType)
	if err != nil {
		return nil, fmt.Errorf("failed to get Definition=[%s], path=[%s/%s/%s], Error:%s", docType, d.Path(), pname, JsonKey.Ref, err)
	}
	if doc == nil {
		return nil, fmt.Errorf("cannot find definition=[%s], path=[%s/%s/%s], no error", docType, d.Path(), pname, JsonKey.Ref)
	}
	return doc, nil
}

func (d *SchemaDoc) getDefinition(dataType string) (*SchemaDoc, error) {
	if d.Definitions != nil {
		doc, ok := d.Definitions[dataType]
		if ok {
			return doc, nil
		}
	}
	if d.Parent != nil {
		doc, err := d.Parent.getDefinition(dataType)
		if err != nil {
			return nil, err
		}
		return doc, nil
	}
	return nil, nil
}
