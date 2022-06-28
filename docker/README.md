# build docker images

### Image List

 - DataService
 - InventoryService

### Local Build
 - build image into local registry
 ```
 cd docker/{service name}
 ./build.sh
 ```
 - this will generate a image in local docker registry
   - REPOSITORY=unitao/{service name} TAG=localbuild
 - list images
 ```
 docker images
 ```
 - to check content of the image
 ```
 docker run -it --rm -v "$(pwd)"/DataService01:/opt/UniTAO/config unitao/dataservice:localbuild 
 ```

 - run DataServiceAdmin
 ```
 docker run -it --rm --network {network_name} -v "$(pwd)"/DataService01:/opt/UniTAO/config unitao/dataservice:localbuild DataServiceAdmin {data service admin arguments}
 ```
   - example - (init table):
   ```
   docker run -it --rm --network 2data1inv_default -v "$(pwd)"/DataService01:/opt/UniTAO/config unitao/dataservice:localbuild initTable.sh 
   ```

  - reset data for data service
  ```
   docker container restart {data service admin container name}
   example:
   docker container restart unitao-data-service01-admin
  ```

- run InventoryServiceAdmin manually
```
docker run -it --rm --network 2data1inv_default -v "$(pwd)"/InventoryService/config:/opt/UniTAO/config -v "$(pwd)"/InventoryService/data:/opt/UniTAO/data unitao/inventoryservice:localbuild InventoryServiceAdmin sync -config /opt/UniTAO/config/config.json
```
  - Sync schema from all Data Sources
  ```
  docker container restart unitao-inv-service-admin
  ```


### Clean Up
 - Clean up stopped Containers:
 ```
 docker container prune -f
 ```




