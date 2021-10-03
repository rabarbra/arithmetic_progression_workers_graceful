package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"workers_server/workerstore"
)

func TestAdd(t *testing.T) {
	var task workerstore.Task
	var worker workerstore.Worker
	var ws *workerstore.WorkerStore

	ws = workerstore.NewWorkerStore()
	ws.GetSortedTasks()
	task = workerstore.Task{
		N:   10,
		D:   1,
		N1:  0,
		I:   2,
		TTL: 5,
	}
	//b, _ := json.Marshal(task)
	s := "{\"n\":10,\"d\":1,\"n1\":0,\"I\":2,\"TTL\":5}"
	r := httptest.NewRequest("POST", "http://127.0.0.1/add", strings.NewReader(s))
	w := httptest.NewRecorder()
	NewTask(w, r)
	fmt.Println(worker)
	correct_worker := workerstore.Worker{
		Task:         task,
		NumInQueue:   1,
		ScheduleTime: time.Now().UTC().Format(time.RFC3339Nano),
		Status:       workerstore.Scheduled,
	}
	json.Unmarshal(w.Body.Bytes(), &worker)
	if correct_worker.Task != worker.Task {
		t.Error("Tasks differ.")
	}
}
