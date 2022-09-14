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

// Schema keywords and helper static function

class Schema {
    static Key = Object.freeze({
        "additionalProperties": "additionalProperties",
        "cmt": "ContentMediaType",
        "description": "description",
        "definitions": "definitions",
        "inventory": "inventory",
        "items": "items",
        "key": "key",
        "name": "name",
        "properties": "properties",
        "ref": "$ref",
        "required": "required",
        "root": "#",
        "type": "type"
    })

    static Type = Object.freeze({
        "array": "array",
        "bool": "bool",
        "float": "float",
        "int": "int",
        "map": "map",
        "object": "object",
        "string": "string"
    })

    static Schema = Object.freeze({
        "name": "schemaOfSchema",
        "description": "schema of Schema Definition",
        "additionalProperties": false,
        "properties": {
            "name": {
                "type": "string",
                "required": false
            },
            "description": {
                "type": "string",
                "required": false
            },
            "properties": {
                "type": "map",
                "items": {
                    "type": "object",
                    "$ref": "#/definitions/properties"
                }
            },
            "definitions": {
                "type": "map",
                "items": {
                    "type": "object",
                    "$ref": "#"
                },
                "required": false
            }
        },
        "definitions": {
            "properties": {
                "additionalProperties": false,
                "properties": {
                    "type": {
                        "type": "string"
                    },
                    "$ref": {
                        "type": "string",
                        "required": false
                    },
                    "items": {
                        "type": "object",
                        "$ref": "#/definitions/properties",
                        "required": false
                    }
                }
            }
        }
    })

    static PrefixRef(){
        return Schema.Key.root + "/" + Schema.Key.definitions + "/"
    }

    static PrefixCMT(){
        return Schema.Key.inventory + "/"
    }

    static GetDef(data, defName){
        let defMap = data[Schema.Key.definitions]
        if (defMap[defName]){
            return defMap[defName]
        }
        return null
    }

}

export {Schema}

