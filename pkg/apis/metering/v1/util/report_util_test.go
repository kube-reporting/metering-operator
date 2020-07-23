package util

import (
	"testing"
	"time"

	v1 "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"

	kapiV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewReportCondition(t *testing.T) {
	expect := v1.ReportCondition{
		Type:    v1.ReportRunning,
		Status:  kapiV1.ConditionTrue,
		Reason:  "reason",
		Message: "message",
	}
	condition := NewReportCondition(expect.Type, expect.Status, expect.Reason, expect.Message)
	if condition == nil {
		t.Error("expected condition to be returned but got nil")
	}

	if condition.Type != expect.Type ||
		condition.Status != expect.Status ||
		condition.Reason != expect.Reason ||
		condition.Message != expect.Message {
		t.Errorf("unexepcted change in returned condition inputs.  Expected: %#v.  Got: %#v", expect, condition)
	}
}

func TestGetReportCondition(t *testing.T) {
	tests := map[string]struct {
		status   v1.ReportStatus
		condType v1.ReportConditionType
		expect   *v1.ReportCondition
	}{
		"not found": {
			status: v1.ReportStatus{
				Conditions: []v1.ReportCondition{v1.ReportCondition{
					Type: v1.ReportConditionType("foo"),
				}},
			},
			condType: v1.ReportRunning,
			expect:   nil,
		},
		"found": {
			status: v1.ReportStatus{
				Conditions: []v1.ReportCondition{
					{Type: v1.ReportRunning},
				},
			},
			condType: v1.ReportRunning,
			expect:   &v1.ReportCondition{Type: v1.ReportRunning},
		},
		"nil conditions": {
			status:   v1.ReportStatus{Conditions: nil},
			condType: v1.ReportRunning,
			expect:   nil,
		},
		"empty conditions": {
			status:   v1.ReportStatus{Conditions: []v1.ReportCondition{}},
			condType: v1.ReportRunning,
			expect:   nil,
		},
	}

	for name, test := range tests {
		actual := GetReportCondition(test.status, test.condType)
		if actual == nil && test.expect != nil {
			t.Errorf("%s expected %#v condition but received nil", name, actual)
			continue
		}

		if actual != nil && test.expect == nil {
			t.Errorf("%s expected nil condition but received %#v", name, actual)
			continue
		}

		if actual == nil && test.expect == nil {
			continue
		}

		// got values for both, check type
		if test.expect.Type != actual.Type {
			t.Errorf("%s expected condition of type %s but got %s", name, test.expect.Type, actual.Type)
		}
	}
}

func TestSetReportCondition(t *testing.T) {
	// shouldn't fail with nil
	SetReportCondition(nil, v1.ReportCondition{})

	// this is a helper method to allow setting a time.  Also dereferences the pointer so the value can be used
	// in structs to make the tests easier to define.
	conditionWithTime := func(condType v1.ReportConditionType, status kapiV1.ConditionStatus, reason, message string, yesterday bool) v1.ReportCondition {
		cond := NewReportCondition(condType, status, reason, message)
		if yesterday {
			cond.LastTransitionTime = metaV1.NewTime(time.Now().AddDate(0, 0, -1))
		}
		return *cond
	}

	tests := map[string]struct {
		status                     *v1.ReportStatus
		cond                       v1.ReportCondition
		expectTransitionTimeChange bool
	}{
		"new condition": {
			status: &v1.ReportStatus{},
			cond:   conditionWithTime(v1.ReportRunning, kapiV1.ConditionTrue, "reason", "message", false),
		},
		"existing condition with no change in status, reason, or message": {
			status: &v1.ReportStatus{Conditions: []v1.ReportCondition{
				conditionWithTime(v1.ReportRunning, kapiV1.ConditionTrue, "reason", "message", false),
			}},
			cond: conditionWithTime(v1.ReportRunning, kapiV1.ConditionTrue, "reason", "message", false),
		},
		"existing condition with change in status": {
			status: &v1.ReportStatus{Conditions: []v1.ReportCondition{
				conditionWithTime(v1.ReportRunning, kapiV1.ConditionTrue, "reason", "message", true),
			}},
			cond:                       conditionWithTime(v1.ReportRunning, kapiV1.ConditionFalse, "reason", "message", false),
			expectTransitionTimeChange: true,
		},
		"existing condition with change in reason": {
			status: &v1.ReportStatus{Conditions: []v1.ReportCondition{
				conditionWithTime(v1.ReportRunning, kapiV1.ConditionTrue, "reason", "message", false),
			}},
			cond: conditionWithTime(v1.ReportRunning, kapiV1.ConditionTrue, "reason1", "message", false),
		},

		//TODO we don't account for changes in message this test will fail, should we?
		//"existing condition with change in message":    {
		//	status: &v1.ReportStatus{Conditions: []v1.ReportCondition{
		//		conditionWithTime(v1.ReportRunning, kapiV1.ConditionTrue, "reason", "message", false),
		//	}},
		//	cond: conditionWithTime(v1.ReportRunning, kapiV1.ConditionTrue, "reason", "message1", false),
		//},
	}

	for name, test := range tests {
		originalCondition := GetReportCondition(*test.status, test.cond.Type)
		SetReportCondition(test.status, test.cond)
		newCondition := GetReportCondition(*test.status, test.cond.Type)

		if newCondition == nil {
			t.Errorf("%s expected condition to be present in report status but was nil", name)
			continue
		}

		if newCondition.Reason != test.cond.Reason ||
			newCondition.Type != test.cond.Type ||
			newCondition.Message != test.cond.Message {
			t.Errorf("%s found unexpected type condition type, reason, message.  Wanted %s, %s, %s but got %s, %s, %s",
				name,
				test.cond.Type, test.cond.Reason, test.cond.Message,
				newCondition.Type, newCondition.Reason, newCondition.Message)
		}

		if test.expectTransitionTimeChange && originalCondition.LastTransitionTime.Equal(&newCondition.LastTransitionTime) {
			t.Errorf("%s expected transition time change. Original: %v, new: %v",
				name, originalCondition.LastTransitionTime, newCondition.LastTransitionTime)
		}
	}
}

func TestRemoveReportCondition(t *testing.T) {
	// shouldn't fail with nil
	RemoveReportCondition(nil, v1.ReportConditionType("foo"))

	status := &v1.ReportStatus{
		Conditions: []v1.ReportCondition{
			{Type: v1.ReportRunning},
			{Type: v1.ReportConditionType("foo")},
		},
	}

	originalLength := len(status.Conditions)
	RemoveReportCondition(status, v1.ReportRunning)

	if len(status.Conditions) != originalLength-1 {
		t.Errorf("expected a condition to be removed but found invalid remaining condition length: %d, %d", originalLength, len(status.Conditions))
	}

	if cond := GetReportCondition(*status, v1.ReportRunning); cond != nil {
		t.Errorf("expected condition of type %s to be removed", v1.ReportRunning)
	}
}
