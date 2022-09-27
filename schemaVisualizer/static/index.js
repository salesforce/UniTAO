'use strict';

function main() {
    console.log("started");
    let e=document.getElementById("main");
    let newEl = new schemaVisualizer(e);

    newEl.init("", configs.visual);  // all from config
    // newEl.init("server", null, config.visual);       // all server
}

