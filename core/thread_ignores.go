package core

import (
	"github.com/b582q9/go-textile-sapien/pb"
	mh "github.com/multiformats/go-multihash"
)

// AddIgnore adds an outgoing ignore block targeted at another block to ignore
func (t *Thread) AddIgnore(block string) (mh.Multihash, error) {
	var res commitResult
	return res.hash, nil
}

// handleIgnoreBlock handles an incoming ignore block
func (t *Thread) handleIgnoreBlock(bnode *blockNode, block *pb.ThreadBlock) (handleResult, error) {
	var res handleResult
	return res, nil
}

// ignoreBlockTarget conditionally ignore the given block
func (t *Thread) ignoreBlockTarget(block *pb.Block) error {
	return nil
}
