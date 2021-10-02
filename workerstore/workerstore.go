//Simple in-memory sync safe store for Task workers.
package workerstore

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"
)

type Status string

const (
	Scheduled Status = "scheduled"
	Working          = "working"
	Done             = "done"
)

//Data we get from user
type Task struct {
	N   uint    `json:"n" validate:"min=0"`   // Number of elements in sequence
	D   float64 `json:"d"`                    // Delta = N(i + 1) - N(i)
	N1  float64 `json:"n1"`                   // First element
	I   float64 `json:"I" validate:"min=0"`   // Interval (seconds)
	TTL float64 `json:"TTL" validate:"min=0"` // Result store time (seconds)
}

//Additional fields
type Worker struct {
	Task
	NumInQueue   uint   `json:"numInQueue,omitempty"`
	CurrIter     uint   `json:"currIteration,omitempty"`
	ScheduleTime string `json:"scheduledTime"`
	StartTime    string `json:"startTime,omitempty"`
	EndTime      string `json:"endTime,omitempty"`
	Status       Status `json:"status,omitempty"`
}

type WorkerStore struct {
	mx sync.RWMutex

	workers       map[int]*Worker // store for all workers
	next_id       int             // next key for next worker added to workers map
	len_scheduled uint            // scheduled workers count
}

func NewWorkerStore() *WorkerStore {
	store := &WorkerStore{}
	store.workers = make(map[int]*Worker)
	store.next_id = 0
	store.len_scheduled = 0
	return store
}

func (w *WorkerStore) AddTask(task Task) Worker {
	w.mx.Lock()
	defer w.mx.Unlock()
	worker := Worker{
		Task:         task,
		NumInQueue:   w.len_scheduled + 1,
		ScheduleTime: time.Now().UTC().Format(time.RFC3339Nano),
		Status:       Scheduled,
	}
	w.workers[w.next_id] = &worker
	w.next_id++
	w.len_scheduled++
	return worker
}

func (w *WorkerStore) executeWorker(ctx context.Context, wg *sync.WaitGroup) {

	var ttl_wg sync.WaitGroup
	var key int

	defer wg.Done()
	defer log.Println("Cancelling worker")
	log.Println("Staring worker")

	for {
		select {
		case <-ctx.Done():
			ttl_wg.Wait()
			return
		default:
			key = -1
			w.mx.Lock()
			for k, wrk := range w.workers {
				if wrk.Status == Scheduled {
					wrk.NumInQueue--
					if wrk.NumInQueue == uint(0) {
						key = k
						w.len_scheduled--
					}
				}
			}
			w.mx.Unlock()
			if key != -1 {
				worker := w.workers[key]
				worker.StartTime = time.Now().UTC().Format(time.RFC3339Nano)
				worker.Status = Working
				for i := 0; i < int(worker.N); i++ {
					worker.CurrIter += 1
					iter := time.After(time.Millisecond * time.Duration(worker.I*1000))
					select {
					case <-ctx.Done():
						ttl_wg.Wait()
						return
					case <-iter:
						worker.N1 += worker.D
					}
				}
				worker.EndTime = time.Now().UTC().Format(time.RFC3339Nano)
				worker.Status = Done
				worker.CurrIter = 0

				ttl_wg.Add(1)
				go func(key int, ctx context.Context) {
					defer ttl_wg.Done()
					ttl_timeout := time.After(time.Millisecond * time.Duration(w.workers[key].TTL*1000))
					select {
					case <-ctx.Done():
						return
					case <-ttl_timeout:
						w.mx.Lock()
						delete(w.workers, key)
						w.mx.Unlock()
					}
				}(key, ctx)
			}
		}
	}
}

func (w *WorkerStore) StartWorkers(ctx context.Context, max_working uint) {
	var wg sync.WaitGroup
	wg.Add(int(max_working))
	for i := uint(0); i < max_working; i++ {
		go w.executeWorker(ctx, &wg)
	}
	select {
	case <-ctx.Done():
		wg.Wait()
		log.Println("Cancelling StartWorkers()")
		return
	}
}

func (w *WorkerStore) GetSortedTasks() []*Worker {
	var res []*Worker
	var keys []int
	w.mx.RLock()
	defer w.mx.RUnlock()
	for key := range w.workers {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	for _, key := range keys {
		res = append(res, w.workers[key])
	}
	return res
}
