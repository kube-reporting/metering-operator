package e2e

import (
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/operator-framework/operator-metering/test/framework"
)

var testFramework *framework.Framework

func TestMain(m *testing.M) {
	kubecon***REMOVED***g := flag.String("kubecon***REMOVED***g", "", "kube con***REMOVED***g path, e.g. $HOME/.kube/con***REMOVED***g")
	ns := flag.String("namespace", "chargeback-ci", "test namespace")
	httpsAPI := flag.Bool("https-api", false, "If true, use https to talk to Metering API")
	flag.Parse()

	var err error
	if testFramework, err = framework.New(*ns, *kubecon***REMOVED***g, *httpsAPI); err != nil {
		logrus.Fatalf("failed to setup framework: %v\n", err)
	}

	os.Exit(m.Run())
}
