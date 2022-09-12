'use strict';

class schemaStore {

    async init(idList, callback) {
        this.callback = callback;
        this.schemaList = configs.testSchema ? configs.testSchema : {};
        for (let id in this.schemaList)
            this.callback(this.schemaList[id], id);
     
        if (configs.serverURL) {
            if (! (idList)) {
                idList = await this.getRest("schema", null);
            }   
            this.lock = new promiseLock();
            this.fetchSchema(idList);
            return this.lock.promise;
        }
        else
            return null;
    }

    get(id) {
        return this.schemaList[id];
    }

    getAll() {
        return this.schemaList;
    } 
    
    async fetchSchema(idList) {
        for (let id of idList) {
            if (id in this.schemaList)
                continue;
            console.log("missing schema: "+id);    
            this.lock.start(1);
            this.getRest(this.buildUrl("_schema",id)).then((resolve, reject) => {  
                this.schemaList[id] = resolve;
                this.callback(this.schemaList[id], id);
                this.lock.complete();
            })
        }    
    }

    buildUrl(type, id) {
        return `${configs.serverURL}${type?"/"+type:""}${id?"/"+id:""}`;
    }

    async getRest(url) {
        let response = await fetch(url);
        if (response.status !== 200) {
            console.log('Looks like there was a problem. Status Code: ' +
            response.status);
        }
        let data = await response.json();
        return (data);
    } 
}
