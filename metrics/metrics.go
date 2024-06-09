package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsInterface interface {
	IncreaseNoteCount()
	IncreaseRegisterCount()
	IncreaseDeauthorizeCount()
	IncreaseDefaultPageCount()
	IncreaseGetDefaultPageCount()
	IncreaseHelpCount()
}

type MetricsClient struct {
	NoteMetrics           prometheus.Counter
	RegisterMetrics       prometheus.Counter
	DeauthorizeMetrics    prometheus.Counter
	DefaultPageMetrics    prometheus.Counter
	GetDefaultPageMetrics prometheus.Counter
	HelpMetrics           prometheus.Counter
}

func NewMetricsClient() MetricsInterface {
	return &MetricsClient{
		NoteMetrics:           NoteMetrics,
		RegisterMetrics:       RegisterMetrics,
		DeauthorizeMetrics:    DeauthorizeMetrics,
		DefaultPageMetrics:    DefaultPageMetrics,
		GetDefaultPageMetrics: GetDefaultPageMetrics,
		HelpMetrics:           HelpMetrics,
	}
}

func (m *MetricsClient) IncreaseNoteCount()           { m.NoteMetrics.Inc() }
func (m *MetricsClient) IncreaseRegisterCount()       { m.RegisterMetrics.Inc() }
func (m *MetricsClient) IncreaseDeauthorizeCount()    { m.DeauthorizeMetrics.Inc() }
func (m *MetricsClient) IncreaseDefaultPageCount()    { m.DefaultPageMetrics.Inc() }
func (m *MetricsClient) IncreaseGetDefaultPageCount() { m.GetDefaultPageMetrics.Inc() }
func (m *MetricsClient) IncreaseHelpCount()           { m.HelpMetrics.Inc() }

var (
	NoteMetrics = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "note_command_count",
			Help: "number of note commands executed",
		},
	)
	RegisterMetrics = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "register_command_count",
			Help: "number of register commands executed",
		},
	)
	DeauthorizeMetrics = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "deauthorize_command_count",
			Help: "number of deauthorize commands executed",
		},
	)
	DefaultPageMetrics = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "defaultpage_command_count",
			Help: "number of defaultpage commands executed",
		},
	)
	GetDefaultPageMetrics = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "getdefaultpage_command_count",
			Help: "number of getdefaultpage commands executed",
		},
	)
	HelpMetrics = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "help_command_count",
			Help: "number of help commands executed",
		},
	)
)
