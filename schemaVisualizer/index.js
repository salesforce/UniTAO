const express = require('express');
const process = require('process')
process.chdir(__dirname)

var cors = require('cors')
const fs = require('fs')

const app = express();
const PORT = 3000;

// app.use(cors())

app.use(express.static('static'));
//app.use(express.static('/Users/sherzog/code/schemaVisualizer/static'));
//app.use(express.static('/Users/sherzog/documents/schemaVisualizer'));
/*app.get('/v1/inventory/:type/:id', (req, res) => {
    // req.query.xxx for query parameters
    // res.send('Hello Worldfff! type= '+req.params.type+' id= '+req.params.id);
    try {
        const data = fs.readFileSync(`/Users/sherzog/Documents/UAGInventoryCg/schema/${req.params.id}-schema.json`, 'utf8');
        res.send(data);
      } catch (err) {
        console.error(err)
      }
});
*/
app.listen(PORT, () => console.log(`Server listening on port: ${PORT}`));


