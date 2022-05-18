Webix Reports backend
=====================

### How to start

- configure DB connection strings in ```config.yml```
There are two databases there, ```appdb``` where report's configuration will be stored
and ```datadb``` where data to analise is stored.
It can be the same database.   

- build the backend
```shell script
go build
```

- start the service
```shell script
# generate test data, optional
./metadb --demodata
# run the service
./metadb --scheme ./demodata/meta.yml
```

- update client side sample to use your backend (change the `url` property to  ```http://localhost:8014/```)

### Schema configuration

Schema for the demodata is stored in [demodata/meta.yml](demodata/meta.yml).

This file describes available objects, their fields and relations.
Content of the file is a serialization of DBInfo structure from the [main.go](main.go) file. 

#### Field type

Supported fields types are

- number
- date
- string
- picklist (list of hardcoded options)
- reference (key to a different model)

#### Field configuration keys

- name - name shown in report builder
- filter - true/false, allow/deny filtering by the field
- key - true/false, primary key (used for references)
- label -  true/false, mark field as object label (will be shown in place of reference)
- ref - id of referenced model/picklist (if any)

### Other CLI command

```shell script
# show cli help
./metadb --help
# generate meta.yml from DB
./metadb --save ./meta.yml
```
