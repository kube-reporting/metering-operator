package v1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1/types"
)

const (
	ReportPlural = "reports"
	ReportKind   = "Report"
)

type ReportGetter interface {
	Reports(namespace string) ReportInterface
}

type ReportInterface interface {
	Create(*types.Report) (*types.Report, error)
	Get(name string) (*types.Report, error)
	Update(*types.Report) (*types.Report, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type reports struct {
	restClient rest.Interface
	client     *dynamic.ResourceClient
	namespace  string
}

func newReports(r rest.Interface, c *dynamic.Client, namespace string) *reports {
	return &reports{
		r,
		c.Resource(
			&metav1.APIResource{
				Kind:       ReportKind,
				Name:       ReportPlural,
				Namespaced: true,
			},
			namespace,
		),
		namespace,
	}
}

func (p *reports) Create(o *types.Report) (*types.Report, error) {
	up, err := UnstructuredFromReport(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Create(up)
	if err != nil {
		return nil, err
	}

	return ReportFromUnstructured(up)
}

func (p *reports) Get(name string) (*types.Report, error) {
	obj, err := p.client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return ReportFromUnstructured(obj)
}

func (p *reports) Update(o *types.Report) (*types.Report, error) {
	up, err := UnstructuredFromReport(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Update(up)
	if err != nil {
		return nil, err
	}

	return ReportFromUnstructured(up)
}

func (p *reports) Delete(name string, options *metav1.DeleteOptions) error {
	return p.client.Delete(name, options)
}

func (p *reports) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := p.restClient.Get().
		Namespace(p.namespace).
		Resource(ReportPlural).
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var reports types.ReportList
	return &reports, json.Unmarshal(b, &reports)
}

func (p *reports) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := p.restClient.Get().
		Pre***REMOVED***x("watch").
		Namespace(p.namespace).
		Resource(ReportPlural).
		FieldsSelectorParam(nil).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&reportDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

// ReportFromUnstructured unmarshals a Report object.
func ReportFromUnstructured(r *unstructured.Unstructured) (*types.Report, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var p types.Report
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	p.TypeMeta.Kind = ReportKind
	p.TypeMeta.APIVersion = Group + "/" + Version
	return &p, nil
}

// UnstructuredFromReport marshals a Report object.
func UnstructuredFromReport(p *types.Report) (*unstructured.Unstructured, error) {
	p.TypeMeta.Kind = ReportKind
	p.TypeMeta.APIVersion = Group + "/" + Version
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var r unstructured.Unstructured
	if err := json.Unmarshal(b, &r.Object); err != nil {
		return nil, err
	}
	return &r, nil
}

type reportDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *reportDecoder) Close() {
	d.close()
}

func (d *reportDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object types.Report
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
