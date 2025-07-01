package client

import (
	"bytes"
	"errors"
	"fmt"
	"time"
)

type RingWait struct {
	buffer []WaitingSlot
	head   int
	tail   int
	cap    int
	len    int
}

func NewRingWait(size int) *RingWait {
	return &RingWait{
		buffer: make([]WaitingSlot, size),
		head:   0,
		tail:   0,
		len:    0,
		cap:    size,
	}
}

func (rw *RingWait) GenInfor(roomName string) string {
	if rw.len == 0 {
		return fmt.Sprintf("%s Waiting:[]\n", roomName)
	}

	buffer := bytes.NewBufferString(fmt.Sprintf("%s Waiting:[", roomName))
	h, t := rw.head, rw.tail

	for h != t {
		buffer.WriteString(fmt.Sprintf("%s,", rw.buffer[h].client.ID()))
		h = (h + 1) % rw.cap
	}
	buffer.WriteString("]\n")
	return buffer.String()
}

var ErrBufferFull = errors.New("buffer full")

func (r *RingWait) Push(client *Client, expTime time.Time) error {
	if r.len == r.cap {
		return ErrBufferFull
	}
	r.buffer[r.tail] = WaitingSlot{
		expTime: expTime,
		client:  client,
	}
	r.len++
	r.tail = (r.tail + 1) % r.cap
	return nil
}

func (r *RingWait) Peek() WaitingSlot {
	if r.len == 0 {
		return WaitingSlot{client: nil}
	}
	return r.buffer[r.head]
}

func (r *RingWait) Pop() {
	if r.len == 0 {
		return
	}

	r.buffer[r.head].client = nil
	r.head = (r.head + 1) % r.cap
	r.len--
}

func (r *RingWait) Len() int {
	return r.len
}

func (r *RingWait) Cap() int {
	return r.cap
}

func (r *RingWait) IsEmpty() bool {
	return r.len == 0
}

func (r *RingWait) IsFull() bool {
	return r.len == r.cap
}
