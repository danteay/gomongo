package gomongo

import (
	"errors"
	"log"
	"time"

	circuit "github.com/rubyist/circuitbreaker"
	"gopkg.in/mgo.v2"
)

// Regenerate status for circuitbreaker
var Regenerate = "regenerate"

// Success status for circuitbreaker
var Success = "success"

// Fail status for circuitbreaker
var Fail = "fail"

// MongoOptions conection options structre
type MongoOptions struct {
	Host       string
	User       string
	Pass       string
	Dbas       string
	Poolsize   int64
	FailRate   float64
	Universe   int64
	TimeOut    time.Duration
	Regenerate time.Duration
}

// MongoPool pool structure
type MongoPool struct {
	cb         *circuit.Breaker
	conn       chan *mgo.Database
	state      string
	trippedAt  int64
	failCount  int64
	regenTryes int64
	Configs    MongoOptions
}

// InitPool sets all connections to pool
func InitPool(opts MongoOptions) (*MongoPool, error) {
	pool := new(MongoPool)
	pool.Configs = opts

	configValidate(&opts)
	pool.cb = circuit.NewRateBreaker(pool.Configs.FailRate, pool.Configs.Universe)
	pool.subscribe()

	err := generatePool(pool, false)
	if err != nil {
		pool.state = Fail
	}

	return pool, err
}

// Execute wrapper for manage failures in circuit breaker
func (pool *MongoPool) Execute(callback func(*mgo.Database) error) error {
	if pool.State() == Fail {
		pool.regenerate()
		return errors.New("unavailable service")
	}

	if pool.State() == Regenerate {
		return errors.New("unavailable service")
	}

	conn := pool.popConx()
	if conn == nil {
		pool.cb.Fail()
		return errors.New("empty connection")
	}

	var err error

	pool.cb.Call(func() error {
		err = callback(conn)
		return nil
	}, pool.Configs.TimeOut)
	pool.pushConx(conn)

	return err
}

// Subscribe events for control reset
func (pool *MongoPool) subscribe() {
	events := pool.cb.Subscribe()

	go func() {
		for {
			e := <-events
			switch e {
			case circuit.BreakerTripped:
				pool.state = Fail
			case circuit.BreakerReset:
				pool.state = Regenerate
			case circuit.BreakerFail:
				log.Println(":::::: breaker fail ::::::")
			case circuit.BreakerReady:
				pool.state = Success
			}
		}
	}()
}

// PopConx return a conection of the pool
func (pool *MongoPool) popConx() *mgo.Database {
	return <-pool.conn
}

// PushConx restore a conection into the pool
func (pool *MongoPool) pushConx(conx *mgo.Database) {
	pool.conn <- conx
}

func configValidate(options *MongoOptions) {
	if options.FailRate < 0.0 {
		options.FailRate = 0.0
	}
	if options.FailRate > 1.0 {
		options.FailRate = 1.0
	}
	if options.Poolsize <= 0 {
		options.Poolsize = 5
	}
	if options.Universe <= options.Poolsize {
		options.Universe = options.Poolsize
	}
	if options.TimeOut < 0 {
		options.TimeOut = 0
	}
	if options.Regenerate <= 0 {
		options.Regenerate = time.Second * 3
	}
}

func (pool *MongoPool) connect() (*mgo.Database, error) {
	var err error
	var db *mgo.Database

	if pool.cb.Tripped() {
		return nil, errors.New("unavailable service")
	}

	err = pool.cb.Call(func() error {
		log.Println("mongo://" + pool.Configs.User + ":" + pool.Configs.Pass + "@" + pool.Configs.Host + "/" + pool.Configs.Dbas)

		c, err := mgo.Dial(pool.Configs.Host)
		if err != nil {
			return err
		}

		c.SetMode(mgo.Monotonic, true)

		db = c.DB(pool.Configs.Dbas)
		err = db.Login(pool.Configs.User, pool.Configs.Pass)
		if err != nil {
			return err
		}

		return nil
	}, pool.Configs.TimeOut)

	return db, err
}

func generatePool(pool *MongoPool, failFirst bool) error {
	pool.conn = make(chan *mgo.Database, pool.Configs.Poolsize)

	for x := int64(0); x < pool.Configs.Poolsize; x++ {
		var conn *mgo.Database
		var err error

		if conn, err = pool.connect(); err != nil {
			if failFirst {
				pool.setTrippedTime()
				return err
			}

			pool.failCount++
		}
		pool.conn <- conn
	}

	if pool.cb.Tripped() {
		pool.setTrippedTime()
		return errors.New("failed to create connection pool")
	}

	pool.state = Success
	return nil
}

func (pool *MongoPool) regenerate() {
	epoch := time.Now().Unix()
	diference := epoch - pool.trippedAt
	regentime := int64(pool.Configs.Regenerate / 1000000000)

	if diference >= regentime && pool.State() == Fail {
		pool.regenTryes++
		pool.reset()

		if err := generatePool(pool, true); err != nil {
			pool.cb.Trip()
			pool.setTrippedTime()
		} else {
			pool.regenTryes = 0
		}
	}
}

func (pool *MongoPool) reset() {
	pool.clean()
	pool.cb.Reset()
	pool.failCount = 0
	pool.trippedAt = 0
}

func (pool *MongoPool) clean() {
	if pool.regenTryes == 0 {
		for x := int64(0); x < pool.Configs.Poolsize; x++ {
			if aux := <-pool.conn; aux != nil {
				aux.Session.Close()
			}
		}
	}
	close(pool.conn)
}

func (pool *MongoPool) setTrippedTime() {
	if pool.trippedAt == 0 {
		trip := time.Now().Unix()
		pool.trippedAt = trip
	}
}

// GetURL connection url
func (pool *MongoPool) GetURL() string {
	return "mongo://" + pool.Configs.User + ":" + pool.Configs.Pass + "@" + pool.Configs.Host + "/" + pool.Configs.Dbas
}

// State returns actual state of the pool
func (pool *MongoPool) State() string {
	return pool.state
}
