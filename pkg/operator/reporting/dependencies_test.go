package reporting

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/test/testhelpers"
)

func newDefault(s string) *json.RawMessage {
	v := json.RawMessage(s)
	return &v
}

func TestDependencyResolver(t *testing.T) {
	testNs := "test-ns"

	ds1 := testhelpers.NewReportDataSource("datasource1", testNs)
	ds2 := testhelpers.NewReportDataSource("datasource2", testNs)
	ds3 := testhelpers.NewReportDataSource("datasource4", testNs)

	testInputs := []metering.ReportQueryInputDefinition{
		{
			Name:     "ds1",
			Type:     "ReportDataSource",
			Required: true,
			Default:  newDefault(`"datasource1"`),
		},
		{
			Name:     "ds2",
			Type:     "ReportDataSource",
			Required: true,
			Default:  newDefault(`"datasource2"`),
		},
		{
			Name:     "q1",
			Type:     "ReportQuery",
			Required: true,
			Default:  newDefault(`"query1"`),
		},
		{
			Name:     "q2",
			Type:     "ReportQuery",
			Required: true,
			Default:  newDefault(`"query2"`),
		},
	}

	query1 := &metering.ReportQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "query1",
			Namespace: testNs,
		},
		Spec: metering.ReportQuerySpec{
			Inputs: []metering.ReportQueryInputDefinition{
				{
					Name:     "ds3",
					Type:     "ReportDataSource",
					Required: true,
					Default:  newDefault(`"datasource4"`),
				},
			},
		},
	}

	query2 := &metering.ReportQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "query2",
			Namespace: testNs,
		},
		Spec: metering.ReportQuerySpec{
			Inputs: []metering.ReportQueryInputDefinition{
				{
					Name:     "q3",
					Type:     "ReportQuery",
					Required: true,
					Default:  newDefault(`"query3"`),
				},
			},
		},
	}

	query3 := &metering.ReportQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "query3",
			Namespace: testNs,
		},
		Spec: metering.ReportQuerySpec{},
	}

	expectedDeps := &ReportQueryDependencies{
		ReportDataSources: []*metering.ReportDataSource{
			ds1, ds2, ds3,
		},
		ReportQueries: []*metering.ReportQuery{
			query1, query2, query3,
		},
		Reports: []*metering.Report{},
	}

	dataSourceGetter := testhelpers.NewReportDataSourceStore([]*metering.ReportDataSource{
		ds1, ds2, ds3,
	})
	queryGetter := testhelpers.NewReportQueryStore([]*metering.ReportQuery{
		query1, query2, query3,
	})
	reportGetter := testhelpers.NewReportStore(nil)

	resolver := NewDependencyResolver(
		queryGetter,
		dataSourceGetter,
		reportGetter,
	)

	results, err := resolver.ResolveDependencies(testNs, testInputs, nil)
	require.NoError(t, err)
	require.Equal(t, expectedDeps, results.Dependencies)
}
