package message

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"sync"
)

// maxLine is the longest line a reader accepts: a state with a long list.
const maxLine = 8 << 20

// Writer writes messages, one per line. It is safe for concurrent use.
type Writer struct {
	mu sync.Mutex
	w  io.Writer
}

// NewWriter writes to w.
func NewWriter(w io.Writer) *Writer { return &Writer{w: w} }

// Write sends one message.
func (w *Writer) Write(m Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err = w.w.Write(append(data, '\n'))
	return err
}

// Reader reads messages, one per line.
type Reader struct{ r *bufio.Reader }

// NewReader reads from r.
func NewReader(r io.Reader) *Reader { return &Reader{r: bufio.NewReader(r)} }

// Next is the next message. A line that is not a message is skipped: an
// application and a designer of different versions still talk. It returns the
// error of the stream (io.EOF at the end).
func (r *Reader) Next() (Message, error) {
	for {
		line, err := r.r.ReadBytes('\n')
		if len(line) > maxLine {
			return Message{}, errors.New("message: a line is longer than the limit")
		}
		var m Message
		if len(line) > 0 && json.Unmarshal(line, &m) == nil && m.Type != "" {
			return m, nil
		}
		if err != nil {
			return Message{}, err
		}
	}
}

// Values turns the values of a state into the raw JSON a message carries.
func Values(values map[string]any) (map[string]json.RawMessage, error) {
	out := make(map[string]json.RawMessage, len(values))
	for name, v := range values {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		out[name] = data
	}
	return out, nil
}
