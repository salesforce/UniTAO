"use strict"

class callbackQueue {

    constructor (callback) {
        this.elements = {}
        this.callback = callback;
    }

    add(item) {
        this.elements[item] = true;
    }

    remove(item) {
        this.elements[item];
    }
}


