package v1

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
	CronPlural = "crons"
	CronKind   = "Cron"
)

type CronGetter interface {
	Crons() CronInterface
}

type CronInterface interface {
	Create(*Cron) (*Cron, error)
	Get(name string) (*Cron, error)
	Update(*Cron) (*Cron, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type crons struct {
	restClient rest.Interface
	client     *dynamic.ResourceClient
}

func newCrons(r rest.Interface, c *dynamic.Client) *crons {
	return &crons{
		r,
		c.Resource(
			&metav1.APIResource{
				Kind:       CronKind,
				Name:       CronPlural,
				Namespaced: true,
			},
			"",
		),
	}
}

func (p *crons) Create(o *Cron) (*Cron, error) {
	up, err := UnstructuredFromCron(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Create(up)
	if err != nil {
		return nil, err
	}

	return CronFromUnstructured(up)
}

func (p *crons) Get(name string) (*Cron, error) {
	obj, err := p.client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return CronFromUnstructured(obj)
}

func (p *crons) Update(o *Cron) (*Cron, error) {
	up, err := UnstructuredFromCron(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Update(up)
	if err != nil {
		return nil, err
	}

	return CronFromUnstructured(up)
}

func (p *crons) Delete(name string, options *metav1.DeleteOptions) error {
	return p.client.Delete(name, options)
}

func (p *crons) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := p.restClient.Get().
		Resource(CronPlural).
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var crons CronList
	return &crons, json.Unmarshal(b, &crons)
}

func (p *crons) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := p.restClient.Get().
		Pre***REMOVED***x("watch").
		Resource(CronPlural).
		FieldsSelectorParam(nil).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&cronDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

// CronFromUnstructured unmarshals a Cron object.
func CronFromUnstructured(r *unstructured.Unstructured) (*Cron, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var p Cron
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	p.TypeMeta.Kind = CronKind
	p.TypeMeta.APIVersion = Group + "/" + Version
	return &p, nil
}

// UnstructuredFromCron marshals a Cron object.
func UnstructuredFromCron(p *Cron) (*unstructured.Unstructured, error) {
	p.TypeMeta.Kind = CronKind
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

type cronDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *cronDecoder) Close() {
	d.close()
}

func (d *cronDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object Cron
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
