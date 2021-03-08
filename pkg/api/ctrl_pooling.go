package api

import (
	"sync"

	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

var stringSlicePool sync.Pool

type stringSlice struct {
	S []string
}

func poolStringSlice() *stringSlice {
	if v := stringSlicePool.Get(); v != nil {
		s := v.(*stringSlice)
		s.Reset()
		return s
	}
	return &stringSlice{S: make([]string, 0, 26)}
}

func (s *stringSlice) Reset()   { s.S = s.S[:0] }
func (s *stringSlice) Release() { stringSlicePool.Put(s) }

// --------------------------------------------------------------------

var schemaValueSlicePool sync.Pool

type schemaValueSlice struct {
	S []schema.Value
}

func poolSchemaValueSlice() *schemaValueSlice {
	if v := schemaValueSlicePool.Get(); v != nil {
		s := v.(*schemaValueSlice)
		s.Reset()
		return s
	}
	return &schemaValueSlice{}
}

func (s *schemaValueSlice) Reset()   { s.S = s.S[:0] }
func (s *schemaValueSlice) Release() { schemaValueSlicePool.Put(s) }

// --------------------------------------------------------------------

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
