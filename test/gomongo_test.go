package test

import (
	"os"
	"testing"
	"time"

	"github.com/danteay/gomongo"
	"github.com/stretchr/testify/assert"
)

func TestConfigStruct(t *testing.T) {
	conf := gomongo.MongoOptions{
		Host:     os.Getenv("AUTH_MONGO_HOST"),
		User:     os.Getenv("AUTH_MONGO_USER"),
		Pass:     os.Getenv("AUTH_MONGO_PASS"),
		Dbas:     os.Getenv("AUTH_MONGO_DBAS"),
		Poolsize: 10,
		FailRate: 0.25,
		Universe: 4,
		TimeOut:  time.Second * 1,
	}

	assert.True(t, conf.Host == os.Getenv("AUTH_MONGO_HOST"))
	assert.True(t, conf.User == os.Getenv("AUTH_MONGO_USER"))
	assert.True(t, conf.Pass == os.Getenv("AUTH_MONGO_PASS"))
	assert.True(t, conf.Dbas == os.Getenv("AUTH_MONGO_DBAS"))
	assert.True(t, conf.Poolsize == 10)
	assert.True(t, conf.FailRate == 0.25)
	assert.True(t, conf.Universe == 4)
	assert.True(t, conf.TimeOut == time.Second*1)
}

func TestEnvUrlConnetc(t *testing.T) {
	_, err := gomongo.InitPool(gomongo.MongoOptions{
		Host: os.Getenv("AUTH_MONGO_HOST"),
		User: os.Getenv("AUTH_MONGO_USER"),
		Pass: os.Getenv("AUTH_MONGO_PASS"),
		Dbas: os.Getenv("AUTH_MONGO_DBAS"),
	})
	assert.Nil(t, err)
}

func TestPoolConection(t *testing.T) {
	conf := gomongo.MongoOptions{
		Host:     os.Getenv("AUTH_MONGO_HOST"),
		User:     os.Getenv("AUTH_MONGO_USER"),
		Pass:     os.Getenv("AUTH_MONGO_PASS"),
		Dbas:     os.Getenv("AUTH_MONGO_DBAS"),
		Poolsize: 5,
		FailRate: 0.25,
		Universe: 4,
		TimeOut:  time.Second * 1,
	}
	_, err := gomongo.InitPool(conf)

	assert.Nil(t, err)
}
