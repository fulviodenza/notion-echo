package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsInterface interface {
	IncreaseNoteCount(lvs []string)
	IncreaseRegisterCount(lvs []string)
	IncreaseDeauthorizeCount(lvs []string)
	IncreaseDefaultPageCount(lvs []string)
	IncreaseGetDefaultPageCount(lvs []string)
	IncreaseHelpCount(lvs []string)
}

type MetricsClient struct {
	NoteMetrics           prometheus.GaugeVec
	RegisterMetrics       prometheus.GaugeVec
	DeauthorizeMetrics    prometheus.GaugeVec
	DefaultPageMetrics    prometheus.GaugeVec
	GetDefaultPageMetrics prometheus.GaugeVec
	HelpMetrics           prometheus.GaugeVec
}

func NewMetricsClient() MetricsInterface {
	return &MetricsClient{
		NoteMetrics:           *NoteMetrics,
		RegisterMetrics:       *RegisterMetrics,
		DeauthorizeMetrics:    *DeauthorizeMetrics,
		DefaultPageMetrics:    *DefaultPageMetrics,
		GetDefaultPageMetrics: *GetDefaultPageMetrics,
		HelpMetrics:           *HelpMetrics,
	}
}

func (m *MetricsClient) IncreaseNoteCount(lvs []string) { m.NoteMetrics.WithLabelValues(lvs...).Inc() }
func (m *MetricsClient) IncreaseRegisterCount(lvs []string) {
	m.RegisterMetrics.WithLabelValues(lvs...).Inc()
}
func (m *MetricsClient) IncreaseDeauthorizeCount(lvs []string) {
	m.DeauthorizeMetrics.WithLabelValues(lvs...).Inc()
}
func (m *MetricsClient) IncreaseDefaultPageCount(lvs []string) {
	m.DefaultPageMetrics.WithLabelValues(lvs...).Inc()
}
func (m *MetricsClient) IncreaseGetDefaultPageCount(lvs []string) {
	m.GetDefaultPageMetrics.WithLabelValues(lvs...).Inc()
}
func (m *MetricsClient) IncreaseHelpCount(lvs []string) {
	m.HelpMetrics.WithLabelValues(lvs...).Inc()
}

var (
	NoteMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "note_command_count",
			Help: "number of note commands executed",
		},
		[]string{"user_id"},
	)
	RegisterMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "register_command_count",
			Help: "number of register commands executed",
		},
		[]string{"user_id"},
	)
	DeauthorizeMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "deauthorize_command_count",
			Help: "number of deauthorize commands executed",
		},
		[]string{"user_id"},
	)
	DefaultPageMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "defaultpage_command_count",
			Help: "number of defaultpage commands executed",
		},
		[]string{"user_id"},
	)
	GetDefaultPageMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "getdefaultpage_command_count",
			Help: "number of getdefaultpage commands executed",
		},
		[]string{"user_id"},
	)
	HelpMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "help_command_count",
			Help: "number of help commands executed",
		},
		[]string{"user_id"},
	)
)
