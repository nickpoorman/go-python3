package python3

/*
#include "Python.h"
#include "macro.h"
*/
import "C"
import (
	"reflect"
	"unsafe"
)

// Py_buffer layer
type Py_buffer struct {
	ptr *C.Py_buffer
}

type PyBUF_Flag int

const (
	PyBUF_SIMPLE       = PyBUF_Flag(C.PyBUF_SIMPLE)
	PyBUF_WRITABLE     = PyBUF_Flag(C.PyBUF_WRITABLE)
	PyBUF_STRIDES      = PyBUF_Flag(C.PyBUF_STRIDES)
	PyBUF_ND           = PyBUF_Flag(C.PyBUF_ND)
	PyBUF_C_CONTIGUOUS = PyBUF_Flag(C.PyBUF_C_CONTIGUOUS)
	PyBUF_INDIRECT     = PyBUF_Flag(C.PyBUF_INDIRECT)
	PyBUF_FORMAT       = PyBUF_Flag(C.PyBUF_FORMAT)
	PyBUF_STRIDED      = PyBUF_Flag(C.PyBUF_STRIDED)
	PyBUF_STRIDED_RO   = PyBUF_Flag(C.PyBUF_STRIDED_RO)
	PyBUF_RECORDS      = PyBUF_Flag(C.PyBUF_RECORDS)
	PyBUF_RECORDS_RO   = PyBUF_Flag(C.PyBUF_RECORDS_RO)
	PyBUF_FULL         = PyBUF_Flag(C.PyBUF_FULL)
	PyBUF_FULL_RO      = PyBUF_Flag(C.PyBUF_FULL_RO)
	PyBUF_CONTIG       = PyBUF_Flag(C.PyBUF_CONTIG)
	PyBUF_CONTIG_RO    = PyBUF_Flag(C.PyBUF_CONTIG_RO)
)

// int PyObject_GetBuffer(PyObject *obj, Py_buffer *view, int flags)
// Export obj into a Py_buffer, view. These arguments must never be NULL. The flags argument is a bit field indicating what kind of buffer the caller is prepared to deal with and therefore what kind of buffer the exporter is allowed to return. The buffer interface allows for complicated memory sharing possibilities, but some caller may not be able to handle all the complexity but may want to see if the exporter will let them take a simpler view to its memory.
//
// Some exporters may not be able to share memory in every possible way and may need to raise errors to signal to some consumers that something is just not possible. These errors should be a BufferError unless there is another error that is actually causing the problem. The exporter can use flags information to simplify how much of the Py_buffer structure is filled in with non-default values and/or raise an error if the object can’t support a simpler view of its memory.
//
// 0 is returned on success and -1 on error.
//
// The following table gives possible values to the flags arguments.
//
// Flag	Description
// PyBUF_SIMPLE	This is the default flag state. The returned buffer may or may not have writable memory. The format of the data will be assumed to be unsigned bytes. This is a “stand-alone” flag constant. It never needs to be ‘|’d to the others. The exporter will raise an error if it cannot provide such a contiguous buffer of bytes.
// PyBUF_WRITABLE	The returned buffer must be writable. If it is not writable, then raise an error.
// PyBUF_STRIDES	This implies PyBUF_ND. The returned buffer must provide strides information (i.e. the strides cannot be NULL). This would be used when the consumer can handle strided, discontiguous arrays. Handling strides automatically assumes you can handle shape. The exporter can raise an error if a strided representation of the data is not possible (i.e. without the suboffsets).
// PyBUF_ND	The returned buffer must provide shape information. The memory will be assumed C-style contiguous (last dimension varies the fastest). The exporter may raise an error if it cannot provide this kind of contiguous buffer. If this is not given then shape will be NULL.
// PyBUF_C_CONTIGUOUS PyBUF_F_CONTIGUOUS PyBUF_ANY_CONTIGUOUS	These flags indicate that the contiguity returned buffer must be respectively, C-contiguous (last dimension varies the fastest), Fortran contiguous (first dimension varies the fastest) or either one. All of these flags imply PyBUF_STRIDES and guarantee that the strides buffer info structure will be filled in correctly.
// PyBUF_INDIRECT	This flag indicates the returned buffer must have suboffsets information (which can be NULL if no suboffsets are needed). This can be used when the consumer can handle indirect array referencing implied by these suboffsets. This implies PyBUF_STRIDES.
// PyBUF_FORMAT	The returned buffer must have true format information if this flag is provided. This would be used when the consumer is going to be checking for what ‘kind’ of data is actually stored. An exporter should always be able to provide this information if requested. If format is not explicitly requested then the format must be returned as NULL (which means 'B', or unsigned bytes)
// PyBUF_STRIDED	This is equivalent to (PyBUF_STRIDES | PyBUF_WRITABLE).
// PyBUF_STRIDED_RO	This is equivalent to (PyBUF_STRIDES).
// PyBUF_RECORDS	This is equivalent to (PyBUF_STRIDES | PyBUF_FORMAT | PyBUF_WRITABLE).
// PyBUF_RECORDS_RO	This is equivalent to (PyBUF_STRIDES | PyBUF_FORMAT).
// PyBUF_FULL	This is equivalent to (PyBUF_INDIRECT | PyBUF_FORMAT | PyBUF_WRITABLE).
// PyBUF_FULL_RO	This is equivalent to (PyBUF_INDIRECT | PyBUF_FORMAT).
// PyBUF_CONTIG	This is equivalent to (PyBUF_ND | PyBUF_WRITABLE).
// PyBUF_CONTIG_RO	This is equivalent to (PyBUF_ND).
func PyObject_GetBuffer(self *PyObject, flags PyBUF_Flag) (buf *Py_buffer, err bool) {
	buf = &Py_buffer{}
	buf.ptr = &C.Py_buffer{}
	err = int(C.PyObject_GetBuffer(toc(self), buf.ptr, C.int(flags))) != 0
	return
}

func PyObject_GetBufferBytes(buf *Py_buffer) []byte {
	length := buf.ptr.len
	slen := int(length)

	var t byte
	v := sliceAt(reflect.TypeOf(t), unsafe.Pointer(buf.ptr.buf), slen)
	return v.Bytes()

	// This GoBytes copies the buffer into Go allocated memory... BAD!
	// return C.GoBytes(unsafe.Pointer(buf.ptr.buf), C.int(length))
}

// sliceAt returns a view of the memory at p as a slice of elem.
// The elem parameter is the element type of the slice, not the complete slice type.
func sliceAt(elem reflect.Type, p unsafe.Pointer, n int) reflect.Value {
	if p == nil && n == 0 {
		return reflect.Zero(reflect.SliceOf(elem))
	}
	return reflect.NewAt(bigArrayOf(elem), p).Elem().Slice3(0, n, n)
}

// bigArrayOf returns the type of a maximally-sized array of t.
//
// This works around the memory-stranding issue described in
// https://golang.org/issue/13656 by producing only one array type per element
// type (instead of one array type per length).
func bigArrayOf(t reflect.Type) reflect.Type {
	n := ^uintptr(0) / uintptr(t.Size())
	const maxInt = uintptr(^uint(0) >> 1)
	if n > maxInt {
		n = maxInt
	}
	return reflect.ArrayOf(int(n), t)
}
