var configs = {
    "comments": 
    `THIS FILE MAY BE GENERATED AUTOMATICALLY as download from the UI. Feel free to change it and rename it "config.js" to have the UI use it.
     
     BASE parameters. 
     serverURL: base inventory URL (excluding type/id). put "" to disable server gets
     testSchema: additional schemas not from server. of {} to disable. When name overlap exists, the testSchema wins over server
     visual: screan location (x/y in percentages) for each schema. If none is found, random locations are used.
     
     `,
    "serverURL": "http://localhost:8004",
    //"serverURL": "https://decwfdxbqa.execute-api.us-west-2.amazonaws.com/v1/inventory",   
    "visual": {}
/*   
    "visual": {
        "netsec.acl.Payload": {
           "locX": 88,
           "locY": 22
        },
        "netsec.acl.Order": {
           "locX": 29,
           "locY": 28
        },
        "netsec.acl.Addresspool": {
           "locX": 40,
           "locY": 60
        },
        "netsec.acl.Locationpath": {
           "locX": 17,
           "locY": 32
        },
        "netsec.raven.Policy": {
           "locX": 53,
           "locY": 2
        },
        "netsec.raven.Node": {
           "locX": 70,
           "locY": 23
        },
        "netsec.raven.Service": {
           "locX": 20,
           "locY": 15
        },
        "netsec.raven.Location": {
           "locX": 77,
           "locY": 44
        },
        "netsec.raven.Provider": {
           "locX": 28,
           "locY": 36
        },
        "netsec.raven.HAProfile": {
           "locX": 40,
           "locY": 23
        },
        "netsec.raven.Computed": {
           "locX": 3,
           "locY": 11
        },
        "netsec.raven.Tunnel": {
           "locX": 1,
           "locY": 74
        },
        "netsec.raven.Orders": {
           "locX": 73,
           "locY": 81
        },
        "netsec.raven.AddressPool": {
           "locX": 61,
           "locY": 61
        },
        "netsec.raven.LocationPath": {
           "locX": 84,
           "locY": 62
        },
        "_schema": {
           "locX": 57,
           "locY": 82
        },
        "netsec.raven.UserSchenario": {
           "locX": 33,
           "locY": 73
        }
     },
     "testSchema_OFFLINE": {
        "_schema": {
           "type": "_schema",
           "id": "_schema",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              },
              "properties": {
                 "type": "array",
                 "items": {
                    "type": "object"
                 }
              }
           }
        },
        "netsec.raven.UserSchenario": {
           "type": "_schema",
           "id": "netsec.raven.UserSchenario",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              },
              "keywords": {
                 "type": "array",
                 "items": {
                    "type": "string"
                 }
              },
              "rootSchema": {
                 "type": "object"
              },
              "rootSchemaType": {
                 "type": "string",
                 "contentMediaType": "inventory/_schema"
              }
           }
        },
        "netsec.raven.Orders": {
           "type": "_schema",
           "id": "netsec.raven.Orders",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              }
           }
        },
        "netsec.raven.Policy": {
           "type": "_schema",
           "id": "netsec.raven.Policy",
           "description": "describes a desired raven functionality",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              },
              "computed": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.Computed"
              },
              "node": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.Node"
              },
              "service": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.Service"
              },
              "haprofile": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.HAProfile"
              },
              "status": {
                 "type": "string"
              }
           }
        },
        "netsec.raven.Node": {
           "type": "_schema",
           "id": "netsec.raven.Node",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              },
              "location": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.Location"
              },
              "address_pool": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.AddressPool"
              }
           }
        },
        "netsec.raven.Service": {
           "type": "_schema",
           "id": "netsec.raven.Service",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              },
              "providers": {
                 "type": "array",
                 "items": {
                    "type": "string",
                    "contentMediaType": "inventory/netsec.raven.Provider"
                 }
              }
           }
        },
        "netsec.raven.AddressPool": {
           "type": "_schema",
           "id": "netsec.raven.AddressPool",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              }
           }
        },
        "netsec.raven.Location": {
           "type": "_schema",
           "id": "netsec.raven.Location",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              },
              "location_paths": {
                 "type": "array",
                 "items": {
                    "type": "string",
                    "contentMediaType": "inventory/netsec.raven.LocationPath"
                 }
              }
           }
        },
        "netsec.raven.LocationPath": {
           "type": "_schema",
           "id": "netsec.raven.LocationPath",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              }
           }
        },
        "netsec.raven.Provider": {
           "type": "_schema",
           "id": "netsec.raven.Provider",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              },
              "location": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.Location"
              }
           }
        },
        "netsec.raven.HAProfile": {
           "type": "_schema",
           "id": "netsec.raven.HAProfiles",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              }
           }
        },
        "netsec.raven.Computed": {
           "type": "_schema",
           "id": "netsec.raven.Computed",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              },
              "tunnels": {
                 "type": "array",
                 "items": {
                    "type": "string",
                    "contentMediaType": "inventory/netsec.raven.Tunnel"
                 }
              }
           }
        },
        "netsec.raven.Tunnel": {
           "type": "_schema",
           "id": "netsec.raven.Tunnel",
           "description": "",
           "properties": {
              "id": {
                 "type": "string"
              },
              "description": {
                 "type": "string"
              },
              "service": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.Service"
              },
              "provider": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.Provider"
              },
              "from": {
                 "type": "string",
                 "contentMediaType": "inventory/netsec.raven.Node"
              }
           }
        }
     }
*/
}

/*

        "netsec.acl.Provider": { 
            "type": "_schema",
            "id": "netsec.acl.Provider",
            "description": "askjdfasjfd",
            "properties": { 
                "another": { "contentMediaType": "inventory/obj3" },
                "final": { "contentMediaType": "inventory/obj4" },
                "inner": {
                    "readOnly": true,
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/netsec.acl.subProvider",
                        "type": "object"
                    }
                }    
            },
            "definitions": {
                "netsec.acl.subProvider": {    
                    "properties": {
                        "toTWO": { "contentMediaType": "inventory/obj2" },     
                    }
                }
            }  
        },
        "obj2": {
            "properties": {
                "O3": { "contentMediaType": "inventory/obj3" },
                "O4": { "contentMediaType": "inventory/obj4" },
                "nothing": {},
                "O1": { "contentMediaType": "inventory/netsec.acl.Provider" }     
            }
        },
        "obj3": {
            "properties": {
                "O4": { "contentMediaType": "inventory/obj4" },
                "O1": { "contentMediaType": "inventory/netsec.acl.Provider" }  
            }
        },
        "obj4": {
            "properties": {
                "O3": { "contentMediaType": "inventory/obj3" }
            }
        }


*/