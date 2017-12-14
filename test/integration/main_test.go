package integration

import (
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/coreos-inc/kube-chargeback/test/framework"
)

var testFramework *framework.Framework

func TestMain(m *testing.M) {
	kubecon***REMOVED***g := flag.String("kubecon***REMOVED***g", "", "kube con***REMOVED***g path, e.g. $HOME/.kube/con***REMOVED***g")
	ns := flag.String("namespace", "chargeback-ci", "test namespace")
	flag.Parse()

	var err error

	if testFramework, err = framework.New(*ns, *kubecon***REMOVED***g); err != nil {
		logrus.Fatalf("failed to setup framework: %v\n", err)
	}

	os.Exit(m.Run())
}
