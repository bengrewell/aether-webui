package taskrunner

import (
	"sync"
	"testing"
)

func TestOutputBuffer_WriteAndString(t *testing.T) {
	var buf OutputBuffer
	buf.Write([]byte("hello "))
	buf.Write([]byte("world"))

	if got := buf.String(); got != "hello world" {
		t.Fatalf("String() = %q, want %q", got, "hello world")
	}
	if got := buf.Len(); got != 11 {
		t.Fatalf("Len() = %d, want 11", got)
	}
}

func TestOutputBuffer_ReadFromZero(t *testing.T) {
	var buf OutputBuffer
	buf.Write([]byte("abc"))

	data, newOff := buf.ReadFrom(0)
	if string(data) != "abc" {
		t.Fatalf("ReadFrom(0) data = %q, want %q", data, "abc")
	}
	if newOff != 3 {
		t.Fatalf("ReadFrom(0) newOffset = %d, want 3", newOff)
	}
}

func TestOutputBuffer_ReadFromMiddle(t *testing.T) {
	var buf OutputBuffer
	buf.Write([]byte("abcdef"))

	data, newOff := buf.ReadFrom(3)
	if string(data) != "def" {
		t.Fatalf("ReadFrom(3) data = %q, want %q", data, "def")
	}
	if newOff != 6 {
		t.Fatalf("ReadFrom(3) newOffset = %d, want 6", newOff)
	}
}

func TestOutputBuffer_ReadFromAtEnd(t *testing.T) {
	var buf OutputBuffer
	buf.Write([]byte("abc"))

	data, newOff := buf.ReadFrom(3)
	if len(data) != 0 {
		t.Fatalf("ReadFrom(3) data = %q, want empty", data)
	}
	if newOff != 3 {
		t.Fatalf("ReadFrom(3) newOffset = %d, want 3", newOff)
	}
}

func TestOutputBuffer_ReadFromBeyondEnd(t *testing.T) {
	var buf OutputBuffer
	buf.Write([]byte("abc"))

	data, newOff := buf.ReadFrom(100)
	if len(data) != 0 {
		t.Fatalf("ReadFrom(100) data = %q, want empty", data)
	}
	if newOff != 3 {
		t.Fatalf("ReadFrom(100) newOffset = %d, want 3", newOff)
	}
}

func TestOutputBuffer_IncrementalReads(t *testing.T) {
	var buf OutputBuffer
	buf.Write([]byte("chunk1"))

	data, off := buf.ReadFrom(0)
	if string(data) != "chunk1" {
		t.Fatalf("first read data = %q, want %q", data, "chunk1")
	}

	buf.Write([]byte("chunk2"))

	data, off = buf.ReadFrom(off)
	if string(data) != "chunk2" {
		t.Fatalf("second read data = %q, want %q", data, "chunk2")
	}

	// No new data.
	data, _ = buf.ReadFrom(off)
	if len(data) != 0 {
		t.Fatalf("third read data = %q, want empty", data)
	}
}

func TestOutputBuffer_ConcurrentWrites(t *testing.T) {
	var buf OutputBuffer
	var wg sync.WaitGroup

	const goroutines = 10
	const writes = 100

	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range writes {
				buf.Write([]byte("x"))
			}
		}()
	}
	wg.Wait()

	if got := buf.Len(); got != goroutines*writes {
		t.Fatalf("Len() = %d, want %d", got, goroutines*writes)
	}
}

func TestOutputBuffer_Empty(t *testing.T) {
	var buf OutputBuffer

	if got := buf.String(); got != "" {
		t.Fatalf("String() on empty buffer = %q, want empty", got)
	}
	if got := buf.Len(); got != 0 {
		t.Fatalf("Len() on empty buffer = %d, want 0", got)
	}

	data, off := buf.ReadFrom(0)
	if len(data) != 0 || off != 0 {
		t.Fatalf("ReadFrom(0) on empty = (%q, %d), want (empty, 0)", data, off)
	}
}
