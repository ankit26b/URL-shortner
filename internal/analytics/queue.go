package analytics

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"url-shortener/internal/model"
	"url-shortener/internal/repository"
)

var ErrQueueFull = errors.New("analytics queue is full")

type Queue struct {
	events     chan model.ClickEvent
	repo       repository.ClickStore
	workerWG   sync.WaitGroup
	workerOnce sync.Once
}

func NewQueue(bufferSize int, repo repository.ClickStore) *Queue {
	if bufferSize <= 0 {
		bufferSize = 1024
	}

	q := &Queue{
		events: make(chan model.ClickEvent, bufferSize),
		repo:   repo,
	}
	q.startWorker()
	return q
}

func (q *Queue) Publish(event model.ClickEvent) error {
	select {
	case q.events <- event:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) startWorker() {
	q.workerWG.Add(1)
	go func() {
		defer q.workerWG.Done()
		for event := range q.events {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err := q.repo.Insert(ctx, event)
			cancel()
			if err != nil {
				log.Printf("failed to persist click event for short_code=%s: %v", event.ShortCode, err)
			}
		}
	}()
}

func (q *Queue) Shutdown() {
	q.workerOnce.Do(func() {
		close(q.events)
		q.workerWG.Wait()
	})
}
