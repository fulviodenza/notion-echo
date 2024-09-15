package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	NoteCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "notion-echo",
		Subsystem: "commands",
		Name:      "note",
		Help:      "note command gauge",
	}, []string{"id"})
)
