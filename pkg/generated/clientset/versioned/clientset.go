package versioned

import (
	chargebackv1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned/typed/chargeback/v1alpha1"
	glog "github.com/golang/glog"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	ChargebackV1alpha1() chargebackv1alpha1.ChargebackV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Chargeback() chargebackv1alpha1.ChargebackV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	chargebackV1alpha1 *chargebackv1alpha1.ChargebackV1alpha1Client
}

// ChargebackV1alpha1 retrieves the ChargebackV1alpha1Client
func (c *Clientset) ChargebackV1alpha1() chargebackv1alpha1.ChargebackV1alpha1Interface {
	return c.chargebackV1alpha1
}

// Deprecated: Chargeback retrieves the default version of ChargebackClient.
// Please explicitly pick a version.
func (c *Clientset) Chargeback() chargebackv1alpha1.ChargebackV1alpha1Interface {
	return c.chargebackV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForCon***REMOVED***g creates a new Clientset for the given con***REMOVED***g.
func NewForCon***REMOVED***g(c *rest.Con***REMOVED***g) (*Clientset, error) {
	con***REMOVED***gShallowCopy := *c
	if con***REMOVED***gShallowCopy.RateLimiter == nil && con***REMOVED***gShallowCopy.QPS > 0 {
		con***REMOVED***gShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(con***REMOVED***gShallowCopy.QPS, con***REMOVED***gShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.chargebackV1alpha1, err = chargebackv1alpha1.NewForCon***REMOVED***g(&con***REMOVED***gShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForCon***REMOVED***g(&con***REMOVED***gShallowCopy)
	if err != nil {
		glog.Errorf("failed to create the DiscoveryClient: %v", err)
		return nil, err
	}
	return &cs, nil
}

// NewForCon***REMOVED***gOrDie creates a new Clientset for the given con***REMOVED***g and
// panics if there is an error in the con***REMOVED***g.
func NewForCon***REMOVED***gOrDie(c *rest.Con***REMOVED***g) *Clientset {
	var cs Clientset
	cs.chargebackV1alpha1 = chargebackv1alpha1.NewForCon***REMOVED***gOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForCon***REMOVED***gOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.chargebackV1alpha1 = chargebackv1alpha1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
