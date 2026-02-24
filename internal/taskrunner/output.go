package taskrunner

import "sync"

// OutputChunk is the result of an incremental read from an OutputBuffer.
type OutputChunk struct {
	Data      string `json:"data"`
	Offset    int    `json:"offset"`
	NewOffset int    `json:"new_offset"`
}

// OutputBuffer is a thread-safe, append-only byte buffer that implements
// io.Writer. It is designed to be used as both cmd.Stdout and cmd.Stderr for
// an exec.Cmd, capturing interleaved output in real-time as the subprocess
// produces it.
type OutputBuffer struct {
	mu   sync.RWMutex
	data []byte
}

// Write appends p to the buffer. It always returns len(p), nil.
func (b *OutputBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	b.data = append(b.data, p...)
	b.mu.Unlock()
	return len(p), nil
}

// ReadFrom returns the bytes from offset to the current end of the buffer,
// along with the new offset (i.e. the current length). If offset is beyond the
// end, an empty slice and the current length are returned.
func (b *OutputBuffer) ReadFrom(offset int) ([]byte, int) {
	b.mu.RLock()
	n := len(b.data)
	if offset >= n {
		b.mu.RUnlock()
		return nil, n
	}
	// Copy to avoid holding the lock while the caller processes the data.
	out := make([]byte, n-offset)
	copy(out, b.data[offset:n])
	b.mu.RUnlock()
	return out, n
}

// Len returns the current size of the buffer.
func (b *OutputBuffer) Len() int {
	b.mu.RLock()
	n := len(b.data)
	b.mu.RUnlock()
	return n
}

// String returns the full buffer contents as a string.
func (b *OutputBuffer) String() string {
	b.mu.RLock()
	s := string(b.data)
	b.mu.RUnlock()
	return s
}
