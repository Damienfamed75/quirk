package quirk

import (
	"context"
	"sync"
	"unsafe"

	"github.com/damienfamed75/quirk/logging"
)

// Client is used to store enough data and help manage
// the logger when inserting nodes into Dgraph using a proper
// upsert procedure.
type Client struct {
	predicateKey string
	logger       logging.Logger
}

// setupClient returns the default states of a quirk client.
func setupClient() *Client {
	return &Client{
		logger:       NewNilLogger(),
		predicateKey: "name",
	}
}

// NewClient will setup a new client with the passed in
// configurations if so chosen to use any.
func NewClient(confs ...ClientConfiguration) *Client {
	q := setupClient()

	// Loop through the configurations and apply them to the client.
	for _, c := range confs {
		c(q)
	}

	return q
}

// InsertNode takes in an Operation to determine if multiple nodes
// will be added or a single node. Then the function will return a
// map of the returned successful UIDs with the key being the predicate
// key value. By default this will be the "name" predicate value.
func (c *Client) InsertNode(ctx context.Context, dg DgraphClient, o *Operation) (map[string]UID, error) {
	if o.SetMultiStruct != nil && o.SetSingleStruct != nil {
		return nil, &Error{
			Msg:      msgTooManyMutationFields,
			File:     "client.go",
			Function: "quirk.Client.InsertNode",
		}
	}

	var err error
	uidMap := make(map[string]string)

	switch {
	case o.SetMultiStruct != nil:
		err = c.mutateMulti(ctx, dg, o.SetMultiStruct, uidMap, c.mutateSingleStruct)
	case o.SetSingleStruct != nil:
		_, err = c.mutateSingleStruct(ctx, dg, o.SetSingleStruct, uidMap, &sync.Mutex{})
	case o.SetStringMap != nil:
		_, err = c.mutateStringMap(ctx, dg, o.SetStringMap, uidMap, &sync.Mutex{})
	case o.SetDynamicMap != nil:
		_, err = c.mutateDynamicMap(ctx, dg, o.SetDynamicMap, uidMap, &sync.Mutex{})
	case o.SetSingleDupleNode != nil:
		_, err = c.mutateSingleDupleNode(ctx, dg, o.SetSingleDupleNode, uidMap, &sync.Mutex{})
	case o.SetMultiDupleNode != nil:
		// TODO work out some way to convert the slice to []interface{} without copying.
		// This could possibly be using the "unsafe" package.
		tmp := make([]interface{}, len(o.SetMultiDupleNode))
		for i, t := range o.SetMultiDupleNode {
			tmp[i] = t
		}

		err = c.mutateMulti(ctx, dg, tmp, uidMap, c.mutateSingleDupleNode)
	}

	return *(*map[string]UID)(unsafe.Pointer(&uidMap)), err
}

// GetPredicateKey returns the name of the field(predicate) that will
// be used to label inserted nodes. By default this is "name"
func (c *Client) GetPredicateKey() string {
	return c.predicateKey
}
