package batch

import "sync"

var recorderPool sync.Pool

func poolRecorder() *ResponseRecorder {
	if v := recorderPool.Get(); v != nil {
		w := v.(*ResponseRecorder)
		w.Reset()
		return w
	}
	return new(ResponseRecorder)
}

func releaseRecorder(w *ResponseRecorder) {
	recorderPool.Put(w)
}
