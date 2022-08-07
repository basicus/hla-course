package queue

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type qStatCounters struct {
	readyCount      prometheus.Gauge
	rejectedCount   prometheus.Gauge
	connectionCount prometheus.Gauge
	consumerCount   prometheus.Gauge
	unackedCount    prometheus.Gauge
}

// TODO добавить в экспортер метрик
func (s *Service) setupCollectMetrics() {
	go func() {
		for {
			queues, _ := s.connection.GetOpenQueues()
			stats, _ := s.connection.CollectStats(queues)
			for queue, queueStats := range stats.QueueStats {
				var counter *qStatCounters
				if c, ok := s.counters[queue]; ok {
					counter = c
				} else {
					counter = getPromCounter(queue)
					s.counters[queue] = counter
					prometheus.MustRegister(counter.readyCount)
					prometheus.MustRegister(counter.rejectedCount)
					prometheus.MustRegister(counter.connectionCount)
					prometheus.MustRegister(counter.consumerCount)
					prometheus.MustRegister(counter.unackedCount)
				}

				counter.readyCount.Set(float64(queueStats.ReadyCount))
				counter.rejectedCount.Set(float64(queueStats.RejectedCount))
				counter.connectionCount.Set(float64(queueStats.ConnectionCount()))
				counter.consumerCount.Set(float64(queueStats.ConsumerCount()))
				counter.unackedCount.Set(float64(queueStats.UnackedCount()))
			}

			time.Sleep(30 * time.Second)
		}
	}()
}

func getPromCounter(queue string) *qStatCounters {
	return &qStatCounters{
		readyCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "queue",
			Name:        "new",
			Help:        "Number of ready messages on queue",
			ConstLabels: prometheus.Labels{"queue": queue},
		}),
		rejectedCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "queue",
			Name:        "rejected",
			Help:        "Number of rejected messages on queue",
			ConstLabels: prometheus.Labels{"queue": queue},
		}),
		connectionCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "queue",
			Name:        "connection",
			Help:        "Number of connections consuming a queue",
			ConstLabels: prometheus.Labels{"queue": queue},
		}),
		consumerCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "queue",
			Name:        "consumer",
			Help:        "Number of consumers consuming messages for a queue",
			ConstLabels: prometheus.Labels{"queue": queue},
		}),
		unackedCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "queue",
			Name:        "in_work",
			Help:        "Number of unacked (in work) messages on a consumer",
			ConstLabels: prometheus.Labels{"queue": queue},
		}),
	}
}
