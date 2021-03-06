package api

import (
	"sync"

	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/riposo"
)

var entSlicePool sync.Pool

type entSlice struct {
	S []permission.ACE
}

func poolEntSlice() *entSlice {
	if v := entSlicePool.Get(); v != nil {
		s := v.(*entSlice)
		s.Reset()
		return s
	}
	return &entSlice{S: make([]permission.ACE, 0, 10)}
}

func (s *entSlice) Append(perm string, path riposo.Path) {
	s.S = append(s.S, permission.ACE{Perm: perm, Path: path})
}
func (s *entSlice) Reset()   { s.S = s.S[:0] }
func (s *entSlice) Release() { entSlicePool.Put(s) }

// --------------------------------------------------------------------

var pathSlicePool sync.Pool

type pathSlice struct {
	S []riposo.Path
}

func poolPathSlice() *pathSlice {
	if v := pathSlicePool.Get(); v != nil {
		s := v.(*pathSlice)
		s.Reset()
		return s
	}
	return &pathSlice{}
}

func (s *pathSlice) Reset()   { s.S = s.S[:0] }
func (s *pathSlice) Release() { pathSlicePool.Put(s) }

// --------------------------------------------------------------------

var callbacksSlicePool sync.Pool

type callbacksSlice struct {
	S []interface{}
}

func poolCallbacksSlice() *callbacksSlice {
	if v := callbacksSlicePool.Get(); v != nil {
		s := v.(*callbacksSlice)
		s.Reset()
		return s
	}
	return &callbacksSlice{}
}

func (s *callbacksSlice) Reset()   { s.S = s.S[:0] }
func (s *callbacksSlice) Release() { callbacksSlicePool.Put(s) }
