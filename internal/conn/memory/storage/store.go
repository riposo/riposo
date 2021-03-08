package storage

import (
	"github.com/riposo/riposo/pkg/params"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

func copyObject(o *schema.Object) *schema.Object {
	co := &schema.Object{
		ID:      o.ID,
		ModTime: o.ModTime,
		Deleted: o.Deleted,
	}
	if o.Extra != nil {
		co.Extra = make([]byte, len(o.Extra))
		copy(co.Extra, o.Extra)
	}
	return co
}

// --------------------------------------------------------------------

type objectNode struct {
	objects map[string]*schema.Object
	modTime riposo.Epoch
}

func (n *objectNode) Len() int {
	if n != nil {
		return len(n.objects)
	}
	return 0
}

func (n *objectNode) Get(objID string) *schema.Object {
	if n != nil {
		return n.objects[objID]
	}
	return nil
}

func (n *objectNode) Put(obj *schema.Object, modTime riposo.Epoch) {
	if n.modTime >= modTime {
		modTime = n.modTime + 1
	}

	obj.ModTime = modTime
	n.modTime = modTime
	n.objects[obj.ID] = copyObject(obj)
}

func (n *objectNode) ForcePut(obj *schema.Object) {
	n.objects[obj.ID] = copyObject(obj)
}

func (n *objectNode) Del(objID string, modTime riposo.Epoch) *schema.Object {
	if o, ok := n.objects[objID]; ok {
		if n.modTime >= modTime {
			modTime = n.modTime + 1
		}

		delete(n.objects, objID)
		n.modTime = modTime
		o.ModTime = modTime
		o.Deleted = true
		return o
	}
	return nil
}

// --------------------------------------------------------------------

type objectTree map[string]*objectNode

func (t objectTree) Len() (n int) {
	for _, node := range t {
		n += node.Len()
	}
	return
}

func (t objectTree) GetNode(ns string) *objectNode {
	return t[ns]
}

func (t objectTree) FetchNode(ns string, modTime riposo.Epoch) *objectNode {
	if node, ok := t[ns]; ok {
		return node
	}

	node := &objectNode{
		objects: make(map[string]*schema.Object),
		modTime: modTime,
	}
	t[ns] = node
	return node
}

func (t objectTree) Get(ns, objID string) *schema.Object {
	return t.GetNode(ns).Get(objID)
}

func (t objectTree) Unlink(ns, objID string) {
	if node := t.GetNode(ns); node != nil {
		delete(node.objects, objID)
		if len(node.objects) == 0 {
			delete(t, ns)
		}
	}
}

func (t objectTree) Each(ns string, cond params.Condition, cb func(*schema.Object)) {
	if node := t.GetNode(ns); node != nil {
		for _, obj := range node.objects {
			if conditionMatch(obj, cond) {
				cb(obj)
			}
		}
	}
}
