package chargeback

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	QueryKind = "Query"
	QueryName = "queries"
)

type QueryGetter interface {
	Queries(namespace string) QueryInterface
}

type QueryInterface interface {
	Create(*Query) (*Query, error)
	Get(name string) (*Query, error)
	Update(*Query) (*Query, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type queries struct {
	restClient rest.Interface
	client     *dynamic.ResourceClient
	ns         string
}

func newQueries(r rest.Interface, c *dynamic.Client, namespace string) *queries {
	return &queries{
		r,
		c.Resource(
			&metav1.APIResource{
				Kind:       QueryKind,
				Name:       QueryName,
				Namespaced: true,
			},
			namespace,
		),
		namespace,
	}
}

func (p *queries) Create(o *Query) (*Query, error) {
	up, err := UnstructuredFromQuery(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Create(up)
	if err != nil {
		return nil, err
	}

	return QueryFromUnstructured(up)
}

func (p *queries) Get(name string) (*Query, error) {
	obj, err := p.client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return QueryFromUnstructured(obj)
}

func (p *queries) Update(o *Query) (*Query, error) {
	up, err := UnstructuredFromQuery(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Update(up)
	if err != nil {
		return nil, err
	}

	return QueryFromUnstructured(up)
}

func (p *queries) Delete(name string, options *metav1.DeleteOptions) error {
	return p.client.Delete(name, options)
}

func (p *queries) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := p.restClient.Get().
		Namespace(p.ns).
		Resource(QueryName).
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var queries QueryList
	return &queries, json.Unmarshal(b, &queries)
}

func (p *queries) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := p.restClient.Get().
		Pre***REMOVED***x("watch").
		Namespace(p.ns).
		Resource(QueryName).
		// VersionedParams(&options, v1.ParameterCodec).
		FieldsSelectorParam(nil).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&queryDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

// QueryFromUnstructured unmarshals a Query object.
func QueryFromUnstructured(r *unstructured.Unstructured) (*Query, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var p Query
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	p.TypeMeta.Kind = QueryKind
	p.TypeMeta.APIVersion = Group + "/" + Version
	return &p, nil
}

// UnstructuredFromQuery marshals a Query object.
func UnstructuredFromQuery(p *Query) (*unstructured.Unstructured, error) {
	p.TypeMeta.Kind = QueryKind
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

type queryDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *queryDecoder) Close() {
	d.close()
}

func (d *queryDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object Query
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
