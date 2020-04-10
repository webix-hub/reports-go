Data Query engine
=======


### How to use

Generate schema

```bash
metadb -save ./scheme.yml
```

Run query engine

```bash
metadb -config ./scheme.yml
```


### REST API

#### Get all data from the tablesave

```
POST /api/data/{table}
```

Body can contain a filtering query

#### Get unique field values

```
GET /api/data/{table}/{field}/suggest
```