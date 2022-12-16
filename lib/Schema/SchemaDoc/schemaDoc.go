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
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Json"
	"github.com/salesforce/UniTAO/lib/Util/Template"
)

type SchemaDoc struct {
	Id          string
	Version     string
	Parent      *SchemaDoc
	KeyTemplate *Template.StrTemp
	Data        map[string]interface{}
	Definitions map[string]*SchemaDoc
	CmtRefs     map[string]*CMTDocRef
	SubDocs     map[string]*SchemaDoc
	RAW         map[string]interface{}
}

type CMTDocRef struct {
	Doc           *SchemaDoc
	Name          string
	CmtType       string
	ContentType   string
	IndexTemplate string
}

func create(data map[string]interface{}, id string, parent *SchemaDoc) (*SchemaDoc, error) {
	version, vreExists := data[JsonKey.Version].(string)
	schemaName, nameExists := data[JsonKey.Name]
	if parent == nil {
		if !nameExists || schemaName.(string) == "" {
			return nil, fmt.Errorf("missing key=[%s] in schema data", JsonKey.Name)
		}
		if id != "" && id != schemaName.(string) {
			return nil, fmt.Errorf("given id=[%s] not match schema name=[%s]", id, schemaName)
		}
		id = schemaName.(string)
		if !vreExists {
			return nil, fmt.Errorf("missing key=[%s] at schema document root", JsonKey.Version)
		}
	} else {
		if nameExists && id != schemaName.(string) {
			return nil, fmt.Errorf("def key=[%s] not match def name=[%s]", id, schemaName)
		}
		if !vreExists {
			version = ""
		}
	}
	parentPath := ""
	if parent != nil {
		parentPath = parent.Path()
	}
	_, ok := data[JsonKey.Properties].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing key [%s], failed to create doc path=[%s/%s]", JsonKey.Properties, parentPath, schemaName)
	}
	keyTemplate, ok := data[JsonKey.Key].(string)
	if !ok {
		keyTemplate = ""
	}
	template, err := Template.ParseStr(keyTemplate, "{", "}")
	if err != nil {
		return nil, fmt.Errorf("invalid template=[%s], Error:%s", keyTemplate, err)
	}
	doc := SchemaDoc{
		Id:          id,
		Version:     version,
		Parent:      parent,
		Data:        data,
		KeyTemplate: template,
		CmtRefs:     map[string]*CMTDocRef{},
		SubDocs:     map[string]*SchemaDoc{},
	}
	if parent == nil {
		rawDataIface, err := Json.Copy(data)
		if err != nil {
			return nil, fmt.Errorf("failed to copy SchemaDoc Data. @path=[%s], Error:%s", parentPath, err)
		}
		doc.RAW = rawDataIface.(map[string]interface{})
	} else {
		rawData, err := parent.GetDefinitionRaw(doc.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get Raw Schema for definition dataType=[%s]", doc.Id)
		}
		doc.RAW = rawData
	}

	docDefs, ok := data[JsonKey.Definitions]
	if ok {
		defMap, ok := docDefs.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to parse field=[%s], path=[%s/%s]", JsonKey.Definitions, parentPath, doc.Id)
		}
		doc.Definitions = make(map[string]*SchemaDoc, len(defMap))
		for key, def := range defMap {
			defObj, ok := def.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("failed to parse definition to object. path=[%s/%s/%s/%s]", parentPath, doc.Id, JsonKey.Definitions, key)
			}
			defDoc, err := create(defObj, key, &doc)
			if err != nil {
				return nil, err
			}
			doc.Definitions[key] = defDoc
		}
	}
	return &doc, nil
}

func FromString(data string) (*SchemaDoc, error) {
	dataObj := map[string]interface{}{}
	err := json.Unmarshal([]byte(data), &dataObj)
	if err != nil {
		return nil, fmt.Errorf("failed to parse string as map[string]interface{}, Error:%s", err)
	}
	return New(dataObj)
}

func New(data map[string]interface{}) (*SchemaDoc, error) {
	doc, err := create(data, "", nil)
	if err != nil {
		return nil, fmt.Errorf("faile to create doc tree. Error: %s", err)
	}
	err = doc.preprocess()
	if err != nil {
		return nil, fmt.Errorf("failed @preprocess, Err:\n%s", err)
	}
	err = doc.validateKeyDefs()
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func ArchivedSchemaId(dataType string, typeVersion string) string {
	return fmt.Sprintf("%s%s%s", dataType, JsonKey.ArchivedSchemaIdDiv, typeVersion)
}

func (d *SchemaDoc) Path() string {
	if d.Parent == nil {
		return d.Id
	}
	return path.Join(d.Parent.Path(), d.Id)
}

func (d *SchemaDoc) Properties() map[string]interface{} {
	return d.Data[JsonKey.Properties].(map[string]interface{})
}

func (d *SchemaDoc) preprocess() error {
	err := d.processRequired()
	if err != nil {
		return fmt.Errorf("preprocess failed @processRequired, [path]=[%s], Error:%s", d.Path(), err)
	}
	err = d.processAttrInvalidKeyChar()
	if err != nil {
		return err
	}
	err = d.processMap()
	if err != nil {
		return fmt.Errorf("preprocess failed @processRequired, [path]=[%s], Error:%s", d.Path(), err)
	}
	err = d.processRefs()
	if err != nil {
		return fmt.Errorf("preprocess failed @processInvRefs, [path]=[%s], Error:%s", d.Path(), err)
	}
	err = d.validateKeyAttrs()
	if err != nil {
		return fmt.Errorf("validate Key Attributes failed. [path]=[%s] Error: %s", d.Path(), err)
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

func (d *SchemaDoc) processAttrInvalidKeyChar() error {
	propPath := path.Join(d.Path(), JsonKey.Properties)
	propMap := d.Data[JsonKey.Properties].(map[string]interface{})
	invalidChars := Util.CheckInvalidKeys(JsonKey.InvalidKeyChars, propMap)
	if len(invalidChars) > 0 {
		return fmt.Errorf("invalid property names found. path:[%s], invalid names: %s", propPath, invalidChars)
	}
	return nil
}

func (d *SchemaDoc) processRequired() error {
	propPath := path.Join(d.Path(), JsonKey.Properties)
	propMap := d.Data[JsonKey.Properties].(map[string]interface{})
	requiredList := make([]interface{}, 0, len(propMap))
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

func (d *SchemaDoc) validateKeyAttrs() error {
	requiredAttrs, ok := d.Data[JsonKey.Required].([]interface{})
	if !ok {
		requiredAttrs = []interface{}{}
	}
	reqMap := map[string]interface{}{}
	for _, attr := range requiredAttrs {
		attrStr := attr.(string)
		reqMap[attrStr] = d.Data[JsonKey.Properties].(map[string]interface{})[attrStr]
	}
	keyMap, err := d.KeyTemplate.BuildVarMap(reqMap)
	if err != nil {
		return fmt.Errorf("required key attr definition validaton failed. Error: %s", err)
	}
	testValue := d.KeyTemplate.TestValue()
	invalidChars := make([]string, 0, len(JsonKey.InvalidKeyChars))
	for _, invalidC := range JsonKey.InvalidKeyChars {
		if strings.Contains(testValue, invalidC) {
			invalidChars = append(invalidChars, invalidC)
		}
	}
	if len(invalidChars) > 0 {
		return fmt.Errorf("invalid schema key definition [%s], found illegal chars [%s] @[%s]", d.KeyTemplate.Template, invalidChars, d.Path())
	}
	for attr, attrDef := range keyMap {
		attrType := attrDef.(map[string]interface{})[JsonKey.Type].(string)
		if attrType != JsonKey.String {
			return fmt.Errorf("only string attribute can be key attribute. [attr]=[%s] [type]=[%s]", attr, attrType)
		}
	}
	return nil
}

// all hash type, array/map of object, the schema for sub-doc should define key
func (d *SchemaDoc) validateKeyDefs() error {
	propDef := d.Data[JsonKey.Properties].(map[string]interface{})
	for attr := range propDef {
		attrDef := propDef[attr].(map[string]interface{})
		switch attrDef[JsonKey.Type].(string) {
		case JsonKey.Array:
			itemType := attrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
			if itemType == JsonKey.Object {
				subDoc := d.SubDocs[attr]
				if len(subDoc.KeyTemplate.Vars) == 0 {
					return fmt.Errorf("no key definition @path=[%s/%s]", d.Path(), attr)
				}
				err := subDoc.validateKeyDefs()
				if err != nil {
					return err
				}
			}
		case JsonKey.Object:
			if IsMap(attrDef) {
				itemType := attrDef[JsonKey.AdditionalProperties].(map[string]interface{})[JsonKey.Type].(string)
				if itemType == JsonKey.Object {
					subDoc := d.SubDocs[attr]
					err := subDoc.validateKeyDefs()
					if err != nil {
						return err
					}
				}
				return nil
			}
			subDoc := d.SubDocs[attr]
			if !d.IsAncestor(subDoc.Id) {
				err := subDoc.validateKeyDefs()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (d *SchemaDoc) processItemDef(pType string, pname string, itemDef map[string]interface{}) error {
	itemType := itemDef[JsonKey.Type].(string)
	switch itemType {
	case JsonKey.String:
		err := d.getCmtRef(pname, itemDef)
		if err != nil {
			return fmt.Errorf("failed to get CmtRef @[path]=[%s/%s]. Error: %s", d.Path(), pname, err)
		}
	case JsonKey.Object:
		if IsMap(itemDef) {
			return fmt.Errorf("invalid schema. %s of map not supported, @[path]=[%s/%s]", pType, d.Path(), pname)
		}
		err := d.getRefDoc(pname, itemDef)
		if err != nil {
			return fmt.Errorf("failed to get ref doc @processItemDef @[path]=[%s/%s]. Error: %s", d.Path(), pname, err)
		}
	}
	return nil
}

func (d *SchemaDoc) processRefs() error {
	for pname, prop := range d.Data[JsonKey.Properties].(map[string]interface{}) {
		propDef := prop.(map[string]interface{})
		switch propDef[JsonKey.Type].(string) {
		case JsonKey.Array:
			itemDef, ok := propDef[JsonKey.Items].(map[string]interface{})
			if !ok {
				return fmt.Errorf("missing key=[%s] for %s @[path]=[%s/%s]", JsonKey.Items, JsonKey.Array, d.Path(), pname)
			}
			err := d.processItemDef(JsonKey.Array, pname, itemDef)
			if err != nil {
				return err
			}
		case JsonKey.Object:
			if IsMap(propDef) {
				itemDef := propDef[JsonKey.AdditionalProperties].(map[string]interface{})
				err := d.processItemDef(JsonKey.Map, pname, itemDef)
				if err != nil {
					return err
				}
				continue
			}
			err := d.getRefDoc(pname, propDef)
			if err != nil {
				return fmt.Errorf("failed to get ref doc @processRefs @[path]=[%s/%s]. Error: %s", d.Path(), pname, err)
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
	if !IsCmtRef(prop) {
		return nil
	}
	cmt := prop[JsonKey.ContentMediaType].(string)
	cmtType, dataType := Util.ParsePath(cmt)
	switch cmtType {
	case JsonKey.Inventory:
		ref := CMTDocRef{
			Doc:           d,
			Name:          pname,
			CmtType:       JsonKey.Inventory,
			ContentType:   dataType,
			IndexTemplate: "",
		}
		idxTemp, ok := prop[JsonKey.IndexTemplate]
		if ok {
			ref.IndexTemplate = idxTemp.(string)
		}
		d.CmtRefs[ref.Name] = &ref
		return nil
	default:
		return fmt.Errorf("[%s]=[%s] not supported. @[path]=[%s/%s]", JsonKey.ContentMediaType, cmtType, d.Path(), pname)
	}
}

func ParseRefName(prop map[string]interface{}) (string, error) {
	ref, ok := prop[JsonKey.Ref].(string)
	if !ok {
		return "", nil
	}
	if ref == JsonKey.DocRoot {
		return ref, nil
	}
	if !strings.HasPrefix(ref, JsonKey.DefinitionPrefix) {
		return "", fmt.Errorf("unknown ref value=[%s]", ref)
	}
	refName := ref[len(JsonKey.DefinitionPrefix):]
	return refName, nil
}

// get referenced sub definition
func (d *SchemaDoc) getRefDoc(pname string, prop map[string]interface{}) error {
	refType, err := ParseRefName(prop)
	if err != nil {
		return fmt.Errorf("failed to parse ref. @path=[%s/%s/%s], Error: %s", d.Path(), pname, JsonKey.Ref, err)
	}
	if refType == "" {
		// if no ref, then do nothing
		return nil
	}
	doc, err := d.GetDefinition(refType)
	if err != nil {
		return fmt.Errorf("failed to get Definition=[%s], path=[%s/%s/%s], Error:%s", refType, d.Path(), pname, JsonKey.Ref, err)
	}
	if doc == nil {
		return fmt.Errorf("cannot find definition=[%s], path=[%s/%s/%s], no error", refType, d.Path(), pname, JsonKey.Ref)
	}
	d.SubDocs[pname] = doc
	return nil
}

func (d *SchemaDoc) GetDefinition(dataType string) (*SchemaDoc, error) {
	if dataType == JsonKey.DocRoot {
		if d.Parent == nil {
			return d, nil
		}
		doc, err := d.Parent.GetDefinition(dataType)
		if err != nil {
			return nil, err
		}
		return doc, nil
	}
	if d.Definitions != nil {
		doc, ok := d.Definitions[dataType]
		if ok {
			return doc, nil
		}
	}
	if d.Parent != nil {
		doc, err := d.Parent.GetDefinition(dataType)
		if err != nil {
			return nil, err
		}
		return doc, nil
	}
	return nil, nil
}

func (d *SchemaDoc) GetDefinitionRaw(dataType string) (map[string]interface{}, error) {
	if dataType == JsonKey.DocRoot {
		if d.Parent == nil {
			return d.RAW, nil
		}
		doc, err := d.Parent.GetDefinitionRaw(dataType)
		if err != nil {
			return nil, err
		}
		return doc, nil
	}
	if d.Definitions != nil {
		doc, ok := d.RAW[JsonKey.Definitions].(map[string]interface{})[dataType].(map[string]interface{})
		if ok {
			return doc, nil
		}
	}
	if d.Parent != nil {
		doc, err := d.Parent.GetDefinitionRaw(dataType)
		if err != nil {
			return nil, err
		}
		return doc, nil
	}
	return nil, nil
}

func (d *SchemaDoc) BuildKey(data map[string]interface{}) (string, error) {
	return d.KeyTemplate.BuildValue(data)
}

func (d *SchemaDoc) IsAncestor(docId string) bool {
	if d.Id == docId {
		return true
	}
	if d.Parent != nil {
		return d.Parent.IsAncestor(docId)
	}
	return false
}

func IsMap(attrDef map[string]interface{}) bool {
	addProps, ok := attrDef[JsonKey.AdditionalProperties]
	if !ok || reflect.TypeOf(addProps).Kind() == reflect.Bool {
		return false
	}
	return true
}

func IsCmtRef(attrDef map[string]interface{}) bool {
	if attrDef[JsonKey.Type] != JsonKey.String {
		return false
	}
	cmt, ok := attrDef[JsonKey.ContentMediaType].(string)
	if !ok {
		return false
	}
	if cmt == "" {
		return false
	}
	return true
}
