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

	// As apart of the metering-ansible-operator's reconciliation workflow,
	// it marks all resources that were created by the operator with a
	// `metering.openshift.io/ns-prune=<namespace>` to help prune stale resources.
	// If we query for a list of PodDisruptionBudget resources that exist in the
	// rf.Namespace by that label selector, we should get back all the PDBs that
	// the operator created itself.
	pdbs, err := rf.KubeClient.PolicyV1beta1().PodDisruptionBudgets(rf.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("metering.openshift.io/ns-prune=%s", rf.Namespace),
	})
	require.NoError(t, err, "expected querying for the list of Metering operand PodDisruptionBudget resources would produce no error")
	require.NotEmpty(t, pdbs.Items, "expected the list of metering operand PodDisruptionBudget resources would exceed a length of zero")

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

				matched = true
				assert.Equal(t, pdb.Spec.MinAvailable.IntVal, tc.MinAvailable, "expected the actual spec.minAvailable would match the expected spec.minAvailable")
			}

			assert.EqualValues(t, matched, tc.Exists, "expected the actual existence of the operand PodDisruptionBudget resource would match the expected existence")
		})
	}
}
