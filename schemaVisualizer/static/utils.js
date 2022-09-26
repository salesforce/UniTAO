class colorPalette {

    constructor(palette) {
        this.palette = palette ? palette : ["darkOrchid",  "hotPink", "MediumSeaGreen", "orange", "RosyBrown",  "cyan",  "cadetBlue", "olive", "gold", "crimson", "coral", "lightpink"];  
        this.idx = 0;
    }

    nextColor() {
        this.idx = this.idx % this.palette.length;
        return this.palette[this.idx++];
    }
}

class promiseLock {

    constructor () {
        this.started = 0;
        this.completed = 0;
        this.outsideResolve;
        this.promise = new Promise((resolve, reject) => {
            this.outsideResolve = resolve; 
        });
    }

    start (n) {
        this.started += n;
    }

    complete () {
        this.completed++;
        if (this.started <= this.completed)
        this.outsideResolve(); 
    }

    promise () {
        return this.outsideResolve;
    }
}

class utils {

    static isEmpty(object) {
        for (const property in object) {
          return false;
        }
        return true;
      }
}

