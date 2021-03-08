package testdata

import (
	"time"

	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/params"
	"github.com/riposo/riposo/pkg/schema"
)

func mapObjectIDs(objs []*schema.Object, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(objs))
	for i, o := range objs {
		ids[i] = o.ID
	}
	return ids, nil
}

// StdSeeds seeds objects.
// 	O1 {id: EPR.ID}
// 	O2 {id: ITR.ID}
func StdSeeds(tx storage.Transaction) (*schema.Object, *schema.Object, error) {
	o1 := &schema.Object{Extra: []byte(`{
		"ary": ["x", 7, null, false, {"z": 8}],
		"mix": "val",
		"non": false,
		"num": 33,
		"str": "k",
		"sub": {"num": 11, "ok": true},
		"yes": true
	}`)}
	o2 := &schema.Object{Extra: []byte(`{
		"mix": true,
		"num": 66.0,
		"sub": {"ok": true}
	}`)}

	if err := tx.Create("/objects/*", o1); err != nil {
		return nil, nil, err
	}
	time.Sleep(3 * time.Millisecond)
	if err := tx.Create("/objects/*", o2); err != nil {
		return nil, nil, err
	}
	return o1, o2, nil
}

// FilterScope applies a filter to a ListAll query and returns the IDs.
func FilterScope(tx storage.Transaction, field, value string) ([]string, error) {
	return ListScope(tx, storage.ListOptions{Condition: params.Condition{
		params.ParseFilter(field, value),
	}})
}

// SortScope applies sorting to a ListAll query and returns the IDs.
func SortScope(tx storage.Transaction, order string) ([]string, error) {
	return ListScope(tx, storage.ListOptions{Sort: params.ParseSort(order)})
}

// ListScope applies options to a ListAll query and returns the IDs.
func ListScope(tx storage.Transaction, opt storage.ListOptions) ([]string, error) {
	return mapObjectIDs(tx.ListAll(nil, "/objects/*", opt))
}
