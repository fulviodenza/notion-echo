package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SendAllCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "notion_echo",
		Subsystem: "commands",
		Name:      "send_all",
		Help:      "send_all command gauge",
	}, []string{"id"})
	DeauthorizeCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "notion_echo",
		Subsystem: "commands",
		Name:      "deauthorize",
		Help:      "deauthorize command gauge",
	}, []string{"id"})
	DefaultPageCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "notion_echo",
		Subsystem: "commands",
		Name:      "defaultpage",
		Help:      "defaultpage command gauge",
	}, []string{"id", "page"})
	GetDefaultPageCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "notion_echo",
		Subsystem: "commands",
		Name:      "getdefaultpage",
		Help:      "getdefaultpage command gauge",
	}, []string{"id"})
	HelpCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "notion_echo",
		Subsystem: "commands",
		Name:      "help",
		Help:      "help command gauge",
	}, []string{"id"})
	NoteCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "notion_echo",
		Subsystem: "commands",
		Name:      "note",
		Help:      "note command gauge",
	}, []string{"id"})
	RegisterCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "notion_echo",
		Subsystem: "commands",
		Name:      "register",
		Help:      "register command gauge",
	}, []string{"id"})
)
