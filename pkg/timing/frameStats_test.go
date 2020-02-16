package timing

import (
	"math"
	"testing"
	"time"
)

func TestFrameStats_Add(t *testing.T) {

	stats := NewFrameStats(time.Millisecond * 500)
	if math.Abs(stats.Quality()) > 0 {
		t.Errorf("Quality %v != 0.0 with zero samples", stats.Quality())
	}
	tStart := time.Now()
	stats.Add(1, tStart.Add(time.Millisecond*100))

	if stats.Quality() < 1 {
		t.Errorf("Quality %v != 1.0 with one sample", stats.Quality())
	}

	stats.Add(2, tStart.Add(time.Millisecond*200))

	if stats.Quality() < 1 {
		t.Errorf("Quality %v != 1.0 with two samples", stats.Quality())
	}

	stats.Add(4, tStart.Add(time.Millisecond*400))

	if math.Abs(stats.Quality()-0.75) > 1e-10 {
		t.Errorf("Quality %v != 0.75 with 3 out of 4 samples", stats.Quality())
	}

	stats.Prune(tStart.Add(time.Millisecond * 100))

	if math.Abs(stats.Quality()-0.75) > 1e-10 {
		t.Errorf("Quality %v != 0.75 after pruning without deleting any samples", stats.Quality())
	}

	stats.Prune(tStart.Add(time.Millisecond * 101))

	if math.Abs(stats.Quality()-2.0/3) > 1e-10 {
		t.Errorf("Quality %v != 2/3 after pruning and deleting oldest sample", stats.Quality())
	}
}
