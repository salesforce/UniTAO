class schemaVisualizer {
    constructor(element) {
        this.box = element.getBoundingClientRect();
        this.box.width -= 2;
        this.box.height -= 2;
        element.innerHTML = `<div id="display" style="height:${this.box.height}; width:${this.box.width}"></div>
                            <svg id="canvas" height=${this.box.height} width=${this.box.width}></svg>`;
        this.display = element.children[0];
        this.canvas = element.children[1];
        this.display.addEventListener("mousedown",this.mousedown.bind(this));
        this.display.addEventListener("mouseup",this.mouseup.bind(this));
        this.moveMethod = this.mousemove.bind(this);
        this.scmStore = new schemaStore();
    }

    async init(schemaForest, visual) {
        this.placeIndex = 0;
        this.colors = new colorPalette();
        if (visual)
            this.visual = visual;  
        this.items = {};
        this.state = {};
        await this.scmStore.init(schemaForest, this.schemaFetched.bind(this));
        this.render();
    }

    schemaFetched(schema, id) {
        this.items[id]= {};
        this.items[id].id = id;
        this.items[id].placeIndex = this.placeIndex++;
        if ( !(id in this.visual) ) {
            this.visual[id] = {
                "locX": Math.floor(Math.random() * 80) + 10,
                "locY": Math.floor(Math.random() * 80) + 10
            }
        };
        Object.assign(this.items[id], this.visual[id]);
        this.items[id].color = this.colors.nextColor(); 

        let root = this.resolveRef (schema);
        let links = this.getNestedMediaTypes(schema, root);
        if (links && Object.keys(links).length > 0) {    
            this.items[id].properties = links;
            let destinations = [];
            for (let l in links)
                destinations.push(links[l].destination);
            this.scmStore.fetchSchema(destinations);
        }   
    }

    resolveRef (schema) {
        let target = null;
        if ("$ref" in schema) {
            let refLink = schema.$ref.split("/")[2];    // remove the #/definitions/
            if (refLink in schema.definitions) 
                target = schema.definitions[refLink];
        }    
        else
            target = schema;   // no ref, the object is left ASIS    
        return target;    
    }

    getNestedMediaTypes(fullSchema, subSchema) {
        let links = {};
        if ( !("properties" in subSchema) )
            return results;   

        for (let id in subSchema.properties) {
            let linkDetails = {"array": false};
            let property = subSchema.properties[id];
            if ("items" in property) {
                property = property.items;        // dig into the items portion (array type is secondary)
                linkDetails.array = true;
            }    
            if ("$ref" in property) {
                let refLink = property.$ref.split("/")[2];
                if (refLink in fullSchema.definitions) 
                    Object.assign(links, this.getNestedMediaTypes(fullSchema, fullSchema.definitions[refLink]));
            }    
            else if ("contentMediaType" in property) {
                linkDetails.destination = property.contentMediaType.split("/")[1];
                links[id] = linkDetails;
            }    
        };
        return links;
    }        

    render () {
        this.display.innerHTML = `<button id="download" class="iconButton"><img src="img/download.png" /></button>`;
        this.canvas.innerHTML = "";
        this.display.firstElementChild.addEventListener("click", this.downloadState.bind(this));

        for (let id in this.items) {
            this.renderRect(this.items[id]);
        }
        this.renderAllLinks();
    }

    renderRect(item) {
        let x = item.locX;
        let y = item.locY;
        if (this.state.mousemove && this.state.mousemove.id == item.id) {
            x += this.state.mousemove.diffX;
            y += this.state.mousemove.diffY;
        }
        let tmpEl = document.createElement("div");
        let propArray = [];
        for (let id in item.properties) {
            propArray.push(`<div id="${item.id}:${id}" class="itemProperty"><a>${id}</a></div>`);
        }
        let z = (this.state.mousemove && this.state.mousemove.id === item.id) ? 900 : 0;
        

        tmpEl.innerHTML = 
            `<div id="itemWrapper:${item.id}" class="itemWrapper" style="left:${this.box.width * x / 100}; top:${this.box.height * y / 100}; z-index:${z};">
                <div class="itemHeader" style="background-color:${item.color}"><a>${item.id}</a></div>
                ${propArray.join("\n")}
            </div>`;
        this.display.appendChild(tmpEl);
    }    

    renderAllLinks() {
        for (let id in this.items) {
            let item = this.items[id];
            for (let property in item.properties) {
                let to = item.properties[property];
                this.renderArrow(id, property, to, this.items[to.destination].color);
            }    
        }    
    }
    
    renderArrow(from, property, to, color) {

        let fromBox = document.getElementById(`${from}:${property}`).getBoundingClientRect();
        let toBox   = document.getElementById(`itemWrapper:${to.destination}`).getBoundingClientRect();
        let direction = -1; // line goes to left
        let fromX = fromBox.left
        let fromY = fromBox.top + fromBox.height / 2;
        let toX = toBox.left + toBox.width / 2;
        let toY = toBox.top;
        let offsetY = 30;
        let offsetX = (toBox.width * 3 / 5);
        if (fromBox.left + fromBox.width / 2 < toX) {
            direction = 1;
            offsetX *= -1;
            fromX += fromBox.width;
        }
        let midPoint = "";
        if (fromY > toY ) {
            midPoint = `L ${toX + offsetX} ${toY-offsetY}`;
        }
    
        let arrayIcon = to.array ? this.drawMultipleRect(fromX + 30 * direction, fromY, color) : ""; 
        let gEl = document.createElementNS('http://www.w3.org/2000/svg','g');
        gEl.style.zIndex = -1000;
        gEl.innerHTML = `
                        <path d="M ${fromX+4*direction} ${fromY} l ${direction * 30} 0 ${midPoint} L ${toX} ${toY-offsetY} L ${toX} ${toY-15}" fill="transparent" stroke="${color}" stroke-width="3"/>
                        <path d="M ${toX} ${toY} l -10 -15 l 20 0 l -10 15" fill="red" stroke="black" stroke-width="2"/>
                        <circle cx="${fromX}" cy="${fromY}" r="4" stroke="black" fill="red" />
                        ${arrayIcon}`;
        this.canvas.appendChild(gEl);                         
    }

    drawMultipleRect(x, y, color) { 
        return `<rect x="${x-2}" y="${y-2}" height="9" width="8" stroke="black" fill=${color} />
                <rect x="${x-4}" y="${y-4}" height="9" width="8" stroke="black" fill=${color} />
                <rect x="${x-6}" y="${y-6}" height="9" width="8" stroke="black" fill=${color} />`;
    }

    mousedown (e) {
        let element = this.eToId(e);
        let ids = element.id.split(":");
        if (ids[0] !== "itemWrapper")
           return;     // not moving element
        element.style.zIndex = 900;   
        this.state.mousedown = {
            type: ids[0],
            id: ids[1],
            x: e.clientX,
            y: e.clientY
        }
        document.addEventListener("mousemove",this.moveMethod);

    }

    mouseup (e) {
        if (!this.state.mousemove) {
            this.resetState();
            return;
        }
        document.removeEventListener("mousemove", this.moveMethod);  
        let id = this.state.mousemove.id;
        this.items[id].locX += this.state.mousemove.diffX; 
        this.items[id].locY += this.state.mousemove.diffY;     
  
        console.log(`up`);
        this.resetState();
        this.render();
    }

    mousemove (e) {
        if (this.state.mousedown) {
            this.state.mousemove = this.state.mousedown;
            delete this.state.mousedown;
        }
        if (!this.state.mousemove) 
            return;
        let id = this.state.mousemove.id;
        this.state.mousemove.diffX =  100 * (e.clientX - this.state.mousemove.x ) / this.box.width;
        this.state.mousemove.diffY =  100 * (e.clientY - this.state.mousemove.y) / this.box.height;        
        this.render();
        console.log("move");
    }

    resetState() {
        this.state = {};
    }

    eToId(e) {
        return e.target.closest("[id]");
    }

    downloadState() {
        let data = JSON.parse(JSON.stringify(configs)); //deep copy
        for (let id in this.items) {
            data.visual[id] = {
                "locX": Math.round(this.items[id].locX),
                "locY": Math.round(this.items[id].locY)
            }
        }
        this.save(JSON.stringify(data, null, 3), "schema.json", "text/plain");
    }

    save(content, fileName, contentType) {
        const a = document.createElement("a");
        const file = new Blob([content], { type: contentType });
        a.href = URL.createObjectURL(file);
        a.download = fileName;
        a.click();
    }     
}