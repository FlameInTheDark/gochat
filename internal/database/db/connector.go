package db

import (
	"fmt"

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
	if logger == nil {
		c.Logger = logger
	}

	s, err := c.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("Error creating CQL connection: %w", err)
	}

	return &CQLCon{s: s}, nil
}

func (con *CQLCon) Session() *gocql.Session {
	return con.s
}

func (con *CQLCon) Close() error {
	con.s.Close()
	return nil
}
