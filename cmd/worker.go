package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"postchi/internal/metrics"
	"postchi/internal/sms"
	"postchi/pkg/env"
	"postchi/pkg/kafka"

	"postchi/pkg/logger"
)

type Worker struct {
	Envs        *env.Envs
	Metrics     *metrics.Metrics
	Logger      logger.LoggerInterface
	KafkaClinet kafka.KafkaInterface
}

func WorkerHandlerInit(l logger.LoggerInterface, e *env.Envs, m *metrics.Metrics, k kafka.KafkaInterface) *Worker {
	return &Worker{Envs: e, Logger: l, Metrics: m, KafkaClinet: k}
}

func (w *Worker) startWorker() {

	workers := w.Envs.SMS_WORKER_COUNT
	queueSize := 1000
	jobs := make(chan kafka.SmsKafkaMessage, queueSize)

	kafkaClient, err := kafka.Init(w.Envs.KAFKA_BROKERS, w.Envs.KAFKA_TOPIC_SMS)
	if err != nil {
		w.Logger.StdLog("error", fmt.Sprintf("[worker] kafka init failed: %v", err))
		panic("worker cannot run kafka not initialized with err " + err.Error())
	}
	defer kafkaClient.Close()

	if err := kafkaClient.UseReader(w.Envs.KAFKA_CONSUMER_GROUP); err != nil {
		w.Logger.StdLog("error", fmt.Sprintf("[worker] kafka reader init failed: %v", err))
		panic("worker cannot read kafka messages with err " + err.Error())
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go w.workerLoop(i+1, jobs, &wg)
	}
	w.Logger.StdLog("info", fmt.Sprintf("[worker] started %d workers; listening on topic=%s group=%s", workers, w.Envs.KAFKA_TOPIC_SMS, w.Envs.KAFKA_CONSUMER_GROUP))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		defer close(jobs)
		for {
			msg, err := kafkaClient.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					w.Logger.StdLog("info", "[worker] context canceled, stopping consumer")
					return
				}
				w.Logger.StdLog("warn", fmt.Sprintf("[worker] read error: %v", err))
				time.Sleep(200 * time.Millisecond)
				continue
			}

			var j kafka.SmsKafkaMessage
			if err := json.Unmarshal(msg.Value, &j); err != nil {
				w.Logger.StdLog("error", fmt.Sprintf("[worker] json decode failed (offset=%d, key=%s): %v", msg.Offset, string(msg.Key), err))
				continue
			}

			select {
			case jobs <- j:
			case <-ctx.Done():
				w.Logger.StdLog("info", "[worker] shutdown while enqueue; dropping remaining messages")
				return
			}
		}
	}()

	<-ctx.Done()
	w.Logger.StdLog("info", "[worker] shutdown signal received; waiting workers to finish in-flight jobs...")
	wg.Wait()
	w.Logger.StdLog("info", "[worker] exit clean")
}

func (w *Worker) workerLoop(id int, jobs <-chan kafka.SmsKafkaMessage, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		providerName := j.Provider

		prov, err := sms.NewProvider(w.Envs, providerName)
		if err != nil {
			w.Logger.StdLog("error", fmt.Sprintf("[worker:%d] init provider(%s) failed: %v", id, providerName, err))
			continue
		}
		svc := sms.NewService(prov)

		start := time.Now()
		status, msgID, sendErr := svc.Send(j.To, j.Content)
		elapsed := time.Since(start)

		if sendErr != nil {
			w.Logger.StdLog("error", fmt.Sprintf("[worker:%d] send fail to=%s provider=%s err=%v elapsed=%s",
				id, j.To, prov.GetName(), sendErr, elapsed))
			continue
		}

		w.Logger.StdLog("info", fmt.Sprintf("[worker:%d] sent OK to=%s provider=%s status=%d msgID=%d elapsed=%s",
			id, j.To, prov.GetName(), status, msgID, elapsed))
	}
}
