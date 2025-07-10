package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/remiehneppo/be-task-management/internal/logger"
	"github.com/robfig/cron/v3"
)

type Do func() error

type IntervalJob struct {
	IntervalTime int64
	Do           Do
}

type ScheduleJob struct {
	Do   Do
	Cron string
}

type Worker struct {
	intervalJob []*IntervalJob
	scheduleJob []*ScheduleJob
	logger      *logger.Logger
	c           *cron.Cron
}

func NewWorker(logger *logger.Logger) *Worker {
	return &Worker{
		intervalJob: make([]*IntervalJob, 0),
		scheduleJob: make([]*ScheduleJob, 0),
		c:           cron.New(),
		logger:      logger,
	}
}

func (w *Worker) Start() {
	for _, job := range w.intervalJob {

		go func(job *IntervalJob) {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ticker := time.NewTicker(time.Duration(job.IntervalTime) * time.Second)
			for {
				select {
				case <-ctx.Done():
					w.logger.Info("Stopping interval job")
					return
				case <-ticker.C:
					if err := job.Do(); err != nil {
						w.logger.Errorf("Error executing interval job: %v", err)
					}
				}
			}
		}(job)
	}

	for _, job := range w.scheduleJob {
		_, err := w.c.AddFunc(job.Cron, func() {
			if err := job.Do(); err != nil {
				w.logger.Errorf("Error executing scheduled job: %v", err)
			}
		})
		if err != nil {
			w.logger.Errorf("Error adding scheduled job: %v", err)
		}
	}
}

func (w *Worker) RegisterIntervalJob(intervalTime int64, do Do) {
	w.intervalJob = append(w.intervalJob, &IntervalJob{
		IntervalTime: intervalTime,
		Do:           do,
	})
}

func (w *Worker) RegisterScheduleJob(cron string, do Do) {
	w.scheduleJob = append(w.scheduleJob, &ScheduleJob{
		Cron: cron,
		Do:   do,
	})
}
