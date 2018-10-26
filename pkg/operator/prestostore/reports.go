package prestostore

import (
	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

type ReportResultsGetter interface {
	GetReportResults(tableName string, columns []presto.Column) ([]presto.Row, error)
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
