package fastpath

import "time"

type SlidingWindow struct {
	windowSize int64 	// total window duration
	bucketSize int64 	// duration of one bucket
	numBuckets int
	buckets []int64
	head int 		 	// current bucket index
	windowStart int64	// start time of head bucket (nanos)
	total int64 		// total count in window
}

func NewSlidingWindow(window time.Duration, buckets int, now int64) *SlidingWindow {
	if buckets <= 0 {
		panic("buckets must be > 0")
	}	

	bucketSize := int64(window) / int64(buckets)
	if bucketSize <= 0 {
		panic("bucket size too small")
	}

	start := now - (now % bucketSize)

	return &SlidingWindow{
		windowSize:  int64(window),
		bucketSize:  bucketSize,
		numBuckets:  buckets,
		buckets:     make([]int64, buckets),
		head:        0,
		windowStart: start,
	}
}


func (w *SlidingWindow) advance(now int64) {
	elapsed := now - w.windowStart
	if elapsed < w.bucketSize {
		return
	}

	steps := int(elapsed / w.bucketSize)

	if steps >= w.numBuckets {
		for i := range w.buckets {
			w.buckets[i] = 0
		}
		w.total = 0
		w.head = 0
		w.windowStart = now - (now % w.bucketSize)
		return
	}

	for i := 0; i < steps; i++ {
		w.head = (w.head + 1) % w.numBuckets
		w.total -= w.buckets[w.head]
		w.buckets[w.head] = 0
	}

	w.windowStart += int64(steps) * w.bucketSize
}

func (w *SlidingWindow) Add(now int64) {
	w.advance(now)
	w.buckets[w.head]++
	w.total++
}

func (w *SlidingWindow) Count(now int64) int64 {
	w.advance(now)
	return w.total
}
