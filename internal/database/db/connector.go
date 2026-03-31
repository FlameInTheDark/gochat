package db

import (
	"context"
	"errors"
	"time"

	"github.com/gocql/gocql"
)

type CQLCon struct {
	s *gocql.Session
}

func NewCQLCon(keyspace string, logger gocql.StdLogger, cluster ...string) (*CQLCon, error) {
	if len(cluster) == 0 {
		cluster = []string{"127.0.0.1"}
	}

	c := gocql.NewCluster(cluster...)
	c.Keyspace = keyspace
	// Attach logger if provided (fix: check for non-nil)
	if logger != nil {
		c.Logger = logger
	}

	// Attempt to create session with retries and 5s delay to avoid restart loops
	var (
		s   *gocql.Session
		err error
		n   int
	)
	for {
		n++
		s, err = c.CreateSession()
		if err == nil {
			break
		}
		if logger != nil {
			logger.Printf("CQL connect attempt %d failed: %v", n, err)
		}
		time.Sleep(5 * time.Second)
	}

	return &CQLCon{s: s}, nil
}

func (con *CQLCon) Session() *gocql.Session {
	return con.s
}

func (con *CQLCon) Ping(ctx context.Context) error {
	if con == nil || con.s == nil {
		return errors.New("scylla session is not connected")
	}
	return con.s.Query("SELECT now() FROM system.local").WithContext(ctx).Exec()
}

func (con *CQLCon) Close() error {
	con.s.Close()
	return nil
}
