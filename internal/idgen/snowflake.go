package idgen

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
)

type IDGenerator struct {
	gen *snowflake.Node
}

func New(nodeId int64) (*IDGenerator, error) {
	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		return nil, fmt.Errorf("unable to make id generator: %w", err)
	}
	return &IDGenerator{node}, nil
}

func (g *IDGenerator) Next() int64 {
	return g.gen.Generate().Int64()
}
