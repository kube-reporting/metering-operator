package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
)

func testEnsurePodDisruptionBudgetsExist(t *testing.T, rf *reportingframework.ReportingFramework) {
	tt := []struct {
		Name         string
		Exists       bool
		MinAvailable int32
	}{
		{
			Name:         "presto-coordinator-pdb",
			Exists:       true,
			MinAvailable: 1,
		},
		{
			Name:         "presto-worker-pdb",
			MinAvailable: 0,
		},
		{
			Name:         "reporting-operator-pdb",
			Exists:       true,
			MinAvailable: 1,
		},
		{
			Name:         "hive-metastore-pdb",
			Exists:       true,
			MinAvailable: 1,
		},
		{
			Name:         "hive-server-pdb",
			Exists:       true,
			MinAvailable: 1,
		},
	}

	pdbs, err := rf.KubeClient.PolicyV1beta1().PodDisruptionBudgets(rf.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("metering.openshift.io/ns-prune=%s", rf.Namespace),
	})
	require.NoError(t, err, "expected querying for the list of Metering operand PodDisruptionBudget resources would produce no error")
	require.NotEmpty(t, pdbs.Items, "expected the list of metering operand PDBs would exceed a length of zero")

	for _, tc := range tt {
		// capture range variables
		tc := tc
		t := t

		t.Run(tc.Name, func(t *testing.T) {
			var matched bool
			for _, pdb := range pdbs.Items {
				if pdb.Name != tc.Name {
					continue
				}

				require.Equal(t, pdb.Spec.MinAvailable.IntVal, tc.MinAvailable, "expected the minAvailable would match the test case minAvailable")
				matched = true
			}

			// TODO: rework so there's no if/else structure
			if !tc.Exists {
				assert.Falsef(t, matched, "expected no PodDisruptionBudget would exist for the %s operand", tc.Name)
			} else {
				assert.Truef(t, matched, "expected there would be a PodDisruptionBudget resource for %s operand", tc.Name)
			}
		})
	}
}
