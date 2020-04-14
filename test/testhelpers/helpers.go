package testhelpers

import (
	"github.com/sirupsen/logrus"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
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
