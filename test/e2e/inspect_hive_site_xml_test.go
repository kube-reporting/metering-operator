package e2e

import (
	"context"
	"encoding/xml"
	"testing"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type config struct {
	Property []property `xml:"property"`
}

type property struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

const configmapName = "hive-config"

func testEnsurePostgresParametersAreMissing(t *testing.T, rf *reportingframework.ReportingFramework) {
	cm, err := rf.KubeClient.CoreV1().ConfigMaps(rf.Namespace).Get(context.Background(), configmapName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	config := config{}
	err = xml.Unmarshal([]byte(cm.Data["hive-site.xml"]), &config)
	require.NoError(t, err, "unmarshing the hive-site.xml should produce no error")

	blocklist := []string{
		"hive.metastore.transactional.event.listeners",
		"hive.metastore.event.listeners",
	}
	for _, property := range config.Property {
		for _, block := range blocklist {
			if property.Name == block {
				assert.NoError(t, err, "expected the %s parameter would not be present in hive-site.xml", property.Name)
			}
		}
	}
}
