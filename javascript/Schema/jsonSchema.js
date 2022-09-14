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

import {Schema} from "./schema.js"

class SchemaDoc{
    // create SchemaDoc without parent. and do schema validation on it
    static New(data){
        let schemaDoc = new SchemaDoc(Schema.Schema)
        schemaDoc.Validate(data)
        return new SchemaDoc(data)
    }

    // no validation. simply build the structure
    constructor(data, name=null, parent=null){
        this.parent = parent
        this.data = data
        if (!name){
            if (!this.data[Schema.Key.name]){
                throw Error("failed to load SchemaDoc, no name for schema")
            }
            name = this.data[Schema.Key.name]
        }
        if (this.data[Schema.Key.name] && name != this.data[Schema.Key.name]){
            throw Error("param name=[" + name + "] not match field name=[" + this.data[Schema.Key.name] + "]")
        }        
    }

    GetDef(name){
        let defData = Schema.GetDef(this.data, name)
        if (defData){
            return defData
        }
        if (this.parent){
            return this.parent.GetDef(name)
        }
        return null
    }

    Validate(data){
        let validator = JsonSchema.GetValidator(this.data)   
        validator.Validate(data)
    }
}

class JsonSchema{
    static GetValidator(data){
        let schemaStr = JSON.stringify(data)
        let schemaData = JSON.parse(schemaStr)
        let schemaDoc = new SchemaDoc(schemaData)
        let validator = new JsonSchema(schemaDoc)
        return validator
    }

    constructor(schema, path="/") {
        if (!(schema instanceof SchemaDoc)){
            throw Error("schema is not a instance of SchemaDoc")
        }
        this.schema = schema
        this.path = path
    }    

    Validate(data){
        if (!data || typeof data !== 'object'){
            throw Error("validation failed. data is null or not an object. @path=[" + this.path + "]")
        }
    }

    ValidateAttrs(data){
        for(let attr in this.schema.data[Schema.Key.properties]){
            let attrDef = this.schema.data[Schema.Key.properties][attr]
            let attrValue = data[attr]
            this.ValidateAttrValue(attrDef, attrValue)
        }
    }

    ValidateAttrValue(attrDef, attrValue, subPath=null){
        let path = this.path
        if (idx){
            path = path + subPath
        }
        if (!attrValue){
            if(attrDef[Schema.Key.required]){
                throw Error("attribute [" + attr + "] required, but not exists. @path=[" + path + "]")
            }
            return
        }
        let attrDefType = attrDef[Schema.Key.type]
        if ([Schema.Type.object, Schema.Type.map].includes(attrDefType) && typeof attrValue !== 'object'){
            throw Error("attrbute [" + attr + "] invalid type. expect type=[" + attrDefType + "] @path=[" + path + "]")
        }
        if (attrDefType == Schema.Type.string && typeof attrValue !== 'string'){
            throw Error("attrbute [" + attr + "] invalid type. expect type=[" + attrDefType + "] @path=[" + path + "]")
        }
        if (attrDefType == Schema.Type.bool && typeof attrValue !== 'boolean'){
            throw Error("attrbute [" + attr + "] invalid type. expect type=[" + attrDefType + "] @path=[" + path + "]")
        }
        if (attrDefType == Schema.Type.int && !Number.isInteger(attrValue)){
            throw Error("attrbute [" + attr + "] invalid type. expect type=[" + attrDefType + "] @path=[" + path + "]")
        }
        if (attrDefType == Schema.Type.float && (Number.isInteger(attrValue) || Number.isNaN(attrValue))){
            throw Error("attrbute [" + attr + "] invalid type. expect type=[" + attrDefType + "] @path=[" + path + "]")
        }
        if(attrDefType == Schema.Type.array){
            if (!Array.isArray(attrValue)) {
                throw Error("attrbute [" + attr + "] invalid type. expect type=[" + attrDefType + "] @path=[" + path + "]")
            }
            let itemDef = attrDef[Schema.Key.items]
            if (itemDef){
                let aryIdx = 0
                for(let itemValue of attrValue){
                    let itemPath = attr + "[" + aryIdx + "]/"
                    this.ValidateAttrValue(itemDef, itemValue, itemPath)
                    aryIdx += 1
                }
            }
        }
        if(attrDefType == Schema.Type.map) {
            let itemDef = attrDef[Schema.Key.items]
            if (itemDef){
                for(let key in attrValue){
                    let itemValue = attrValue[key]
                    let itemPath = attr + "[" + key + "]/"
                    this.ValidateAttrValue(itemDef, itemValue, itemPath)
                }
            }
        }
        if(attrDefType == Schema.Type.object){
            let refPrefix = Schema.PrefixRef()
            if(attrDef[Schema.Key.ref]){
                if(!attrDef[Schema.Key.ref].startsWith(refPrefix)){
                    throw Error("invalid ref=[" + attrDef[Schema.Key.ref] + "], expect [" + refPrefix + "{definitionName}] @path=[" + path + attr + "]")
                }
                let defName = attrDef[Schema.Key.ref].replace(refPrefix, "")
                let schemaData = this.schema.GetDef(defName)
                if(!defSchema){
                    throw Error("invalid ref=[" + ttrDef[Schema.Key.ref] + "], definition not found, @path=[" + path + attr + "]")
                }
                let schemaDoc = new SchemaDoc(schemaData, defName, this.schema) 
                let validator = new JsonSchema(schemaDoc, path + attr + "/")
                validator.Validate(attrValue)
            }
        }
    }
}

export {SchemaDoc}

