# gomongo

## Install

```bash
go get -u -v github.com/danteay/gomongo
```

And import in your files whit the next lines:

```go
import (
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "github.com/danteay/gomongo"
)
```

## Configure

Setup config for circut and recover strategies

```go
conf := gomongo.MongoOptions{
  Host:     "MONGO_HOST",
  User:     "MONGO_USER",
  Pass:     "MONGO_PASS",
  Dbas:     "MONGO_DBAS",
  Poolsize: 10,
  FailRate: 0.25,
  Universe: 4,
  TimeOut:  time.Second * 1,
}
```

Init connection pool

```go
pool, err := gomongo.InitPool(conf)

if err != nil {
  fmt.Println(err)
}
```

Execute querys inside of the circuit breaker

```go
type Person struct {
    Name string
    Phone string
}

errQuery := pool.Execute(func(db *mgo.Database) error {
  return db.C("people").Insert(
    &Person{"Ale", "+55 53 8116 9639"},
    &Person{"Cla", "+55 53 8402 8510"},
  )
})

if errQuery != nil {
  fmt.Println(errQuery)
}
```

Helt check of the pool connection

```go
fmt.Println("==>> State: ", pool.State())
```