package ring

import "errors"

type Data interface {
	Clear()
}

type RingBuffer[T Data] struct {
	buffer []T
	head   int
	tail   int
	size   int
	cap    int
}

// NewRingBuffer creates a new RingBuffer with the given capacity.
func New[T Data](capacity int) *RingBuffer[T] {
	if capacity <= 0 {
		panic("capacity must be greater than zero")
	}
	return &RingBuffer[T]{
		buffer: make([]T, capacity),
		head:   0,
		tail:   0,
		size:   0,
		cap:    capacity,
	}
}

// Push adds an item to the ring buffer. Returns an error if the buffer is full.
func (rb *RingBuffer[T]) Push(item T) error {
	if rb.size == rb.cap {
		return errors.New("ring buffer is full")
	}
	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.cap
	rb.size++
	return nil
}

// Pop removes and returns the oldest item from the ring buffer. Returns an error if the buffer is empty.
func (rb *RingBuffer[T]) Pop() (T, error) {
	var zero T
	if rb.size == 0 {
		return zero, errors.New("ring buffer is empty")
	}

	item := rb.buffer[rb.head]

	rb.head = (rb.head + 1) % rb.cap
	rb.size--
	return item, nil
}

// IsFull returns true if the buffer is full.
func (rb *RingBuffer[T]) IsFull() bool {
	return rb.size == rb.cap
}

// IsEmpty returns true if the buffer is empty.
func (rb *RingBuffer[T]) IsEmpty() bool {
	return rb.size == 0
}

// Size returns the number of elements currently in the buffer.
func (rb *RingBuffer[T]) Size() int {
	return rb.size
}

// Capacity returns the total capacity of the buffer.
func (rb *RingBuffer[T]) Capacity() int {
	return rb.cap
}
