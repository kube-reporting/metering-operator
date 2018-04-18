package integration

import (
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/coreos-inc/kube-chargeback/test/framework"
)

var (
	testFramework *framework.Framework
)

func TestMain(m *testing.M) {
	kubeconfig := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	ns := flag.String("namespace", "chargeback-ci", "test namespace")
	flag.Parse()

	var err error

	if testFramework, err = framework.New(*ns, *kubeconfig); err != nil {
		logrus.Fatalf("failed to setup framework: %v\n", err)
	}

	os.Exit(m.Run())
}
