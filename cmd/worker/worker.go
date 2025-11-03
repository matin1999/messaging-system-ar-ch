package worker

import (
	"context"
	"encoding/json"
	"errors"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
	"postchi/internal/metrics"
	"postchi/internal/sms"
	"postchi/pkg/db"
	"postchi/pkg/env"
	"postchi/pkg/kafka"
	"postchi/pkg/logger"
)

type Worker struct {
	Envs        *env.Envs
	Metrics     *metrics.Metrics
	Logger      logger.LoggerInterface
	KafkaClinet kafka.KafkaInterface
	Db          db.DataBaseInterface
}

func WorkerHandlerInit(l logger.LoggerInterface, e *env.Envs, m *metrics.Metrics, k kafka.KafkaInterface, d db.DataBaseInterface) *Worker {
	return &Worker{Envs: e, Logger: l, Metrics: m, KafkaClinet: k, Db: d}
}

func (w *Worker) Start() {

	const queueSize = 1000

	if err := w.KafkaClinet.UseReader(w.Envs.KAFKA_CONSUMER_GROUP); err != nil {
		w.Logger.StdLog("error", "[worker] kafka reader init failed: "+err.Error())
		return
	}
	defer w.KafkaClinet.Close()

	workers := w.Envs.SMS_WORKER_COUNT
	jobs := make(chan kafka.SmsKafkaMessage, queueSize)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go w.workerLoop(jobs, &wg)
	}
	w.Logger.StdLog("info", "[worker] started workers")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		defer close(jobs)
		for {
			msg, err := w.KafkaClinet.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					w.Logger.StdLog("info", "[worker] shutdown requested")
					return
				}
				w.Logger.StdLog("warn", "[worker] read error: "+err.Error())
				time.Sleep(200 * time.Millisecond)
				continue
			}

			var j kafka.SmsKafkaMessage
			if err := json.Unmarshal(msg.Value, &j); err != nil {
				w.Logger.StdLog("error", "[worker] json decode failed: "+err.Error())
				continue
			}

			select {
			case jobs <- j:
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	w.Logger.StdLog("info", "[worker] waiting for in-flight jobs...")
	wg.Wait()
	w.Logger.StdLog("info", "[worker] exit")
}

func (w *Worker) workerLoop(jobs <-chan kafka.SmsKafkaMessage, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		prov, err := sms.NewProvider(w.Envs, j.Provider)
		if err != nil {
			w.Logger.StdLog("error", "[worker] init provider failed: "+err.Error())
			continue
		}
		svc := sms.NewService(prov)

		start := time.Now()
		status, msgID, sendErr := svc.Send(j.To, j.Content)
		elapsed := time.Since(start)

		if sendErr != nil {
			w.Logger.StdLog("error", "[worker] send failed: "+sendErr.Error())
			continue
		}
		if err := w.Db.MarkSmsSent(j.UserId, j.ServiceId, j.SmsId, prov.GetName(), msgID); err != nil {
			w.Logger.StdLog("error", "[worker] failed to update SMS and deduct credit: "+err.Error())
		}
		w.Logger.StdLog("info",
			"[worker] sent OK to="+j.To+
				" provider="+prov.GetName()+
				" status="+strconv.Itoa(status)+
				" msgID="+strconv.Itoa(msgID)+
				" elapsed="+elapsed.String())
	}
}
