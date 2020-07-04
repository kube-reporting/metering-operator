package testhelpers

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/operator/reportingutil"
	"github.com/kube-reporting/metering-operator/pkg/presto"
)

// NewReport creates a mock report used for testing purposes.
func NewReport(name, namespace, testQueryName string, reportStart, reportEnd *time.Time, status metering.ReportStatus, schedule *metering.ReportSchedule, runImmediately bool) *metering.Report {
	var start, end *meta.Time
	if reportStart != nil {
		start = &meta.Time{Time: *reportStart}
	}
	if reportEnd != nil {
		end = &meta.Time{Time: *reportEnd}
	}
	return &metering.Report{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: metering.ReportSpec{
			QueryName:      testQueryName,
			ReportingStart: start,
			ReportingEnd:   end,
			Schedule:       schedule,
			RunImmediately: runImmediately,
		},
		Status: status,
	}
}

func NewReportQuery(name, namespace string, columns []metering.ReportQueryColumn) *metering.ReportQuery {
	return &metering.ReportQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: metering.ReportQuerySpec{
			Columns: columns,
		},
	}
}

func NewReportDataSource(name, namespace string) *metering.ReportDataSource {
	return &metering.ReportDataSource{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func NewPrestoTable(name, namespace, catalog, schema string, columns []presto.Column) *metering.PrestoTable {
	return &metering.PrestoTable{
		ObjectMeta: meta.ObjectMeta{
			Name:      reportingutil.TableResourceNameFromKind("Report", namespace, name),
			Namespace: namespace,
		},
		Status: metering.PrestoTableStatus{
			Catalog:   catalog,
			Schema:    schema,
			TableName: name,
			Columns:   columns,
		},
	}
}

type ReportDataSourceStore struct {
	datasources map[string]*metering.ReportDataSource
}

func NewReportDataSourceStore(datasources []*metering.ReportDataSource) (store *ReportDataSourceStore) {
	m := make(map[string]*metering.ReportDataSource)
	for _, dataSource := range datasources {
		m[dataSource.Namespace+"/"+dataSource.Name] = dataSource
	}
	return &ReportDataSourceStore{m}
}

func (store *ReportDataSourceStore) GetReportDataSource(namespace, name string) (*metering.ReportDataSource, error) {
	dataSource, ok := store.datasources[namespace+"/"+name]
	if ok {
		return dataSource, nil
	}
	return nil, errors.NewNotFound(metering.Resource("ReportDataSource"), name)
}

type ReportQueryStore struct {
	queries map[string]*metering.ReportQuery
}

func NewReportQueryStore(queries []*metering.ReportQuery) (store *ReportQueryStore) {
	m := make(map[string]*metering.ReportQuery)
	for _, query := range queries {
		m[query.Namespace+"/"+query.Name] = query
	}
	return &ReportQueryStore{m}
}

func (store *ReportQueryStore) GetReportQuery(namespace, name string) (*metering.ReportQuery, error) {
	query, ok := store.queries[namespace+"/"+name]
	if ok {
		return query, nil
	}
	return nil, errors.NewNotFound(metering.Resource("ReportQuery"), name)
}

type ReportStore struct {
	reports map[string]*metering.Report
}

func NewReportStore(reports []*metering.Report) (store *ReportStore) {
	m := make(map[string]*metering.Report)
	for _, report := range reports {
		m[report.Namespace+"/"+report.Name] = report
	}
	return &ReportStore{m}
}

func (store *ReportStore) GetReport(namespace, name string) (*metering.Report, error) {
	report, ok := store.reports[namespace+"/"+name]
	if ok {
		return report, nil
	}
	return nil, errors.NewNotFound(metering.Resource("Report"), name)
}

func PtrToBool(val bool) *bool {
	return &val
}

func SetupLogger(logLevelStr string) logrus.FieldLogger {
	var err error

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01-02-2006 15:04:05",
	})

	logger := logrus.WithFields(logrus.Fields{
		"app": "deploy",
	})
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logger.WithError(err).Fatalf("invalid log level: %s", logLevel)
	}
	logger.Infof("Setting the log level to %s", logLevel.String())
	logger.Logger.Level = logLevel

	return logger
}

// SetupLoggerToFile is a helper function that initializes and returns a logrus
// FieldLogger instance that directs its output to the @path file instead of
// os.Stdout.
func SetupLoggerToFile(path, logLevel string, fields logrus.Fields) (logrus.FieldLogger, *os.File, error) {
	logger := logrus.New()

	if logLevel == "" {
		logLevel = "debug"
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse the %s log level: %v", logLevel, err)
	}

	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01-02-2006 15:04:05",
	})

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open the %s file path: %v", path, err)
	}
	logger.SetOutput(file)

	return logger.WithFields(fields), file, nil
}

// ExecActionOptions holds all the metadata required to fire off a
// Pod exec REST API call. This is mainly a wrapper around the
// corev1.ExecAction type: https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#ExecAction
type ExecActionOptions struct {
	Name      string
	Namespace string
	Container string
	Command   []string
	UseTTY    bool
}

func NewExecOptions(name, namespace, container string, useTTY bool, cmd []string) *ExecActionOptions {
	return &ExecActionOptions{
		Name:      name,
		Namespace: namespace,
		Container: container,
		UseTTY:    useTTY,
		Command:   cmd,
	}
}

func ExecPodCommandWithOptions(config *rest.Config, client kubernetes.Interface, options *ExecActionOptions) (bytes.Buffer, bytes.Buffer, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	req := client.CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Name(options.Name).
		Namespace(options.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: options.Container,
			Command:   options.Command,
			TTY:       options.UseTTY,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return stdoutBuf, stderrBuf, err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
		Tty:    options.UseTTY,
	})
	if err != nil {
		return stdoutBuf, stderrBuf, err
	}

	return stdoutBuf, stderrBuf, nil
}
