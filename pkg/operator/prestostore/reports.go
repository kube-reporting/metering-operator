package prestostore

import (
	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

type ReportResultsGetter interface {
	GetReportResults(tableName string, columns []presto.Column) ([]presto.Row, error)
}

type ReportResultsStorer interface {
	StoreReportResults(tableName, query string) error
}

type ReportsResultsDeleter interface {
	DeleteReportResults(tableName string) error
}

type ReportResultsRepo interface {
	ReportResultsGetter
	ReportResultsStorer
	ReportsResultsDeleter
}

type reportResultsRepo struct {
	queryer db.Queryer
}

func NewReportResultsRepo(queryer db.Queryer) *reportResultsRepo {
	return &reportResultsRepo{queryer: queryer}
}

func (r *reportResultsRepo) GetReportResults(tableName string, columns []presto.Column) ([]presto.Row, error) {
	return presto.GetRows(r.queryer, tableName, columns)
}

func (r *reportResultsRepo) StoreReportResults(tableName, query string) error {
	return presto.InsertInto(r.queryer, tableName, query)
}

func (r *reportResultsRepo) DeleteReportResults(tableName string) error {
	return presto.DeleteFrom(r.queryer, tableName)
}
