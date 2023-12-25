package queue

import "sync"

type URLWithDepth struct {
	URL   string
	Depth int
}

type URLQueue struct {
	mu    sync.Mutex
	queue []URLWithDepth
}

func (q *URLQueue) Enqueue(url string, depth int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, URLWithDepth{URL: url, Depth: depth})
}

func (q *URLQueue) Dequeue() (URLWithDepth, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.queue) == 0 {
		return URLWithDepth{}, false
	}
	url := q.queue[0]
	q.queue = q.queue[1:]
	return url, true
}
