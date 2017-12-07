// +build !ignore_autogenerated

// This ***REMOVED***le was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	reflect "reflect"
)

func init() {
	SchemeBuilder.Register(RegisterDeepCopies)
}

// RegisterDeepCopies adds deep-copy functions to the given scheme. Public
// to allow building arbitrary schemes.
//
// Deprecated: deepcopy registration will go away when static deepcopy is fully implemented.
func RegisterDeepCopies(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedDeepCopyFuncs(
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*AWSBillingDataSource).DeepCopyInto(out.(*AWSBillingDataSource))
			return nil
		}, InType: reflect.TypeOf(&AWSBillingDataSource{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*DataStoreSource).DeepCopyInto(out.(*DataStoreSource))
			return nil
		}, InType: reflect.TypeOf(&DataStoreSource{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*GenQueryColumn).DeepCopyInto(out.(*GenQueryColumn))
			return nil
		}, InType: reflect.TypeOf(&GenQueryColumn{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*GenQueryView).DeepCopyInto(out.(*GenQueryView))
			return nil
		}, InType: reflect.TypeOf(&GenQueryView{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*LocalStorage).DeepCopyInto(out.(*LocalStorage))
			return nil
		}, InType: reflect.TypeOf(&LocalStorage{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*PromsumDataSource).DeepCopyInto(out.(*PromsumDataSource))
			return nil
		}, InType: reflect.TypeOf(&PromsumDataSource{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*PromsumDataSourceStorageLocation).DeepCopyInto(out.(*PromsumDataSourceStorageLocation))
			return nil
		}, InType: reflect.TypeOf(&PromsumDataSourceStorageLocation{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*Report).DeepCopyInto(out.(*Report))
			return nil
		}, InType: reflect.TypeOf(&Report{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportDataStore).DeepCopyInto(out.(*ReportDataStore))
			return nil
		}, InType: reflect.TypeOf(&ReportDataStore{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportDataStoreList).DeepCopyInto(out.(*ReportDataStoreList))
			return nil
		}, InType: reflect.TypeOf(&ReportDataStoreList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportDataStoreSpec).DeepCopyInto(out.(*ReportDataStoreSpec))
			return nil
		}, InType: reflect.TypeOf(&ReportDataStoreSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportGenerationQuery).DeepCopyInto(out.(*ReportGenerationQuery))
			return nil
		}, InType: reflect.TypeOf(&ReportGenerationQuery{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportGenerationQueryList).DeepCopyInto(out.(*ReportGenerationQueryList))
			return nil
		}, InType: reflect.TypeOf(&ReportGenerationQueryList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportGenerationQuerySpec).DeepCopyInto(out.(*ReportGenerationQuerySpec))
			return nil
		}, InType: reflect.TypeOf(&ReportGenerationQuerySpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportList).DeepCopyInto(out.(*ReportList))
			return nil
		}, InType: reflect.TypeOf(&ReportList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportPrometheusQuery).DeepCopyInto(out.(*ReportPrometheusQuery))
			return nil
		}, InType: reflect.TypeOf(&ReportPrometheusQuery{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportPrometheusQueryList).DeepCopyInto(out.(*ReportPrometheusQueryList))
			return nil
		}, InType: reflect.TypeOf(&ReportPrometheusQueryList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportPrometheusQuerySpec).DeepCopyInto(out.(*ReportPrometheusQuerySpec))
			return nil
		}, InType: reflect.TypeOf(&ReportPrometheusQuerySpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportSpec).DeepCopyInto(out.(*ReportSpec))
			return nil
		}, InType: reflect.TypeOf(&ReportSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportStatus).DeepCopyInto(out.(*ReportStatus))
			return nil
		}, InType: reflect.TypeOf(&ReportStatus{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportStorageLocation).DeepCopyInto(out.(*ReportStorageLocation))
			return nil
		}, InType: reflect.TypeOf(&ReportStorageLocation{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*ReportTemplateSpec).DeepCopyInto(out.(*ReportTemplateSpec))
			return nil
		}, InType: reflect.TypeOf(&ReportTemplateSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*S3Bucket).DeepCopyInto(out.(*S3Bucket))
			return nil
		}, InType: reflect.TypeOf(&S3Bucket{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*StorageLocation).DeepCopyInto(out.(*StorageLocation))
			return nil
		}, InType: reflect.TypeOf(&StorageLocation{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*StorageLocationList).DeepCopyInto(out.(*StorageLocationList))
			return nil
		}, InType: reflect.TypeOf(&StorageLocationList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*StorageLocationSpec).DeepCopyInto(out.(*StorageLocationSpec))
			return nil
		}, InType: reflect.TypeOf(&StorageLocationSpec{})},
	)
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSBillingDataSource) DeepCopyInto(out *AWSBillingDataSource) {
	*out = *in
	if in.Source != nil {
		in, out := &in.Source, &out.Source
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(S3Bucket)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSBillingDataSource.
func (in *AWSBillingDataSource) DeepCopy() *AWSBillingDataSource {
	if in == nil {
		return nil
	}
	out := new(AWSBillingDataSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataStoreSource) DeepCopyInto(out *DataStoreSource) {
	*out = *in
	if in.Promsum != nil {
		in, out := &in.Promsum, &out.Promsum
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(PromsumDataSource)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.AWSBilling != nil {
		in, out := &in.AWSBilling, &out.AWSBilling
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(AWSBillingDataSource)
			(*in).DeepCopyInto(*out)
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataStoreSource.
func (in *DataStoreSource) DeepCopy() *DataStoreSource {
	if in == nil {
		return nil
	}
	out := new(DataStoreSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenQueryColumn) DeepCopyInto(out *GenQueryColumn) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenQueryColumn.
func (in *GenQueryColumn) DeepCopy() *GenQueryColumn {
	if in == nil {
		return nil
	}
	out := new(GenQueryColumn)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenQueryView) DeepCopyInto(out *GenQueryView) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenQueryView.
func (in *GenQueryView) DeepCopy() *GenQueryView {
	if in == nil {
		return nil
	}
	out := new(GenQueryView)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalStorage) DeepCopyInto(out *LocalStorage) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalStorage.
func (in *LocalStorage) DeepCopy() *LocalStorage {
	if in == nil {
		return nil
	}
	out := new(LocalStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PromsumDataSource) DeepCopyInto(out *PromsumDataSource) {
	*out = *in
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(PromsumDataSourceStorageLocation)
			(*in).DeepCopyInto(*out)
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PromsumDataSource.
func (in *PromsumDataSource) DeepCopy() *PromsumDataSource {
	if in == nil {
		return nil
	}
	out := new(PromsumDataSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PromsumDataSourceStorageLocation) DeepCopyInto(out *PromsumDataSourceStorageLocation) {
	*out = *in
	if in.StorageSpec != nil {
		in, out := &in.StorageSpec, &out.StorageSpec
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(StorageLocationSpec)
			(*in).DeepCopyInto(*out)
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PromsumDataSourceStorageLocation.
func (in *PromsumDataSourceStorageLocation) DeepCopy() *PromsumDataSourceStorageLocation {
	if in == nil {
		return nil
	}
	out := new(PromsumDataSourceStorageLocation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Report) DeepCopyInto(out *Report) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Report.
func (in *Report) DeepCopy() *Report {
	if in == nil {
		return nil
	}
	out := new(Report)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Report) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportDataStore) DeepCopyInto(out *ReportDataStore) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportDataStore.
func (in *ReportDataStore) DeepCopy() *ReportDataStore {
	if in == nil {
		return nil
	}
	out := new(ReportDataStore)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportDataStore) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportDataStoreList) DeepCopyInto(out *ReportDataStoreList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*ReportDataStore, len(*in))
		for i := range *in {
			if (*in)[i] == nil {
				(*out)[i] = nil
			} ***REMOVED*** {
				(*out)[i] = new(ReportDataStore)
				(*in)[i].DeepCopyInto((*out)[i])
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportDataStoreList.
func (in *ReportDataStoreList) DeepCopy() *ReportDataStoreList {
	if in == nil {
		return nil
	}
	out := new(ReportDataStoreList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportDataStoreList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportDataStoreSpec) DeepCopyInto(out *ReportDataStoreSpec) {
	*out = *in
	in.DataStoreSource.DeepCopyInto(&out.DataStoreSource)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportDataStoreSpec.
func (in *ReportDataStoreSpec) DeepCopy() *ReportDataStoreSpec {
	if in == nil {
		return nil
	}
	out := new(ReportDataStoreSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportGenerationQuery) DeepCopyInto(out *ReportGenerationQuery) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportGenerationQuery.
func (in *ReportGenerationQuery) DeepCopy() *ReportGenerationQuery {
	if in == nil {
		return nil
	}
	out := new(ReportGenerationQuery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportGenerationQuery) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportGenerationQueryList) DeepCopyInto(out *ReportGenerationQueryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*ReportGenerationQuery, len(*in))
		for i := range *in {
			if (*in)[i] == nil {
				(*out)[i] = nil
			} ***REMOVED*** {
				(*out)[i] = new(ReportGenerationQuery)
				(*in)[i].DeepCopyInto((*out)[i])
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportGenerationQueryList.
func (in *ReportGenerationQueryList) DeepCopy() *ReportGenerationQueryList {
	if in == nil {
		return nil
	}
	out := new(ReportGenerationQueryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportGenerationQueryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportGenerationQuerySpec) DeepCopyInto(out *ReportGenerationQuerySpec) {
	*out = *in
	if in.ReportQueries != nil {
		in, out := &in.ReportQueries, &out.ReportQueries
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DataStores != nil {
		in, out := &in.DataStores, &out.DataStores
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Columns != nil {
		in, out := &in.Columns, &out.Columns
		*out = make([]GenQueryColumn, len(*in))
		copy(*out, *in)
	}
	out.View = in.View
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportGenerationQuerySpec.
func (in *ReportGenerationQuerySpec) DeepCopy() *ReportGenerationQuerySpec {
	if in == nil {
		return nil
	}
	out := new(ReportGenerationQuerySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportList) DeepCopyInto(out *ReportList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*Report, len(*in))
		for i := range *in {
			if (*in)[i] == nil {
				(*out)[i] = nil
			} ***REMOVED*** {
				(*out)[i] = new(Report)
				(*in)[i].DeepCopyInto((*out)[i])
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportList.
func (in *ReportList) DeepCopy() *ReportList {
	if in == nil {
		return nil
	}
	out := new(ReportList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportPrometheusQuery) DeepCopyInto(out *ReportPrometheusQuery) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportPrometheusQuery.
func (in *ReportPrometheusQuery) DeepCopy() *ReportPrometheusQuery {
	if in == nil {
		return nil
	}
	out := new(ReportPrometheusQuery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportPrometheusQuery) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportPrometheusQueryList) DeepCopyInto(out *ReportPrometheusQueryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*ReportPrometheusQuery, len(*in))
		for i := range *in {
			if (*in)[i] == nil {
				(*out)[i] = nil
			} ***REMOVED*** {
				(*out)[i] = new(ReportPrometheusQuery)
				(*in)[i].DeepCopyInto((*out)[i])
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportPrometheusQueryList.
func (in *ReportPrometheusQueryList) DeepCopy() *ReportPrometheusQueryList {
	if in == nil {
		return nil
	}
	out := new(ReportPrometheusQueryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportPrometheusQueryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportPrometheusQuerySpec) DeepCopyInto(out *ReportPrometheusQuerySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportPrometheusQuerySpec.
func (in *ReportPrometheusQuerySpec) DeepCopy() *ReportPrometheusQuerySpec {
	if in == nil {
		return nil
	}
	out := new(ReportPrometheusQuerySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportSpec) DeepCopyInto(out *ReportSpec) {
	*out = *in
	in.ReportingStart.DeepCopyInto(&out.ReportingStart)
	in.ReportingEnd.DeepCopyInto(&out.ReportingEnd)
	if in.GracePeriod != nil {
		in, out := &in.GracePeriod, &out.GracePeriod
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(v1.Duration)
			**out = **in
		}
	}
	if in.Output != nil {
		in, out := &in.Output, &out.Output
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(ReportStorageLocation)
			(*in).DeepCopyInto(*out)
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportSpec.
func (in *ReportSpec) DeepCopy() *ReportSpec {
	if in == nil {
		return nil
	}
	out := new(ReportSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportStatus) DeepCopyInto(out *ReportStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportStatus.
func (in *ReportStatus) DeepCopy() *ReportStatus {
	if in == nil {
		return nil
	}
	out := new(ReportStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportStorageLocation) DeepCopyInto(out *ReportStorageLocation) {
	*out = *in
	if in.StorageSpec != nil {
		in, out := &in.StorageSpec, &out.StorageSpec
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(StorageLocationSpec)
			(*in).DeepCopyInto(*out)
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportStorageLocation.
func (in *ReportStorageLocation) DeepCopy() *ReportStorageLocation {
	if in == nil {
		return nil
	}
	out := new(ReportStorageLocation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportTemplateSpec) DeepCopyInto(out *ReportTemplateSpec) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportTemplateSpec.
func (in *ReportTemplateSpec) DeepCopy() *ReportTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(ReportTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *S3Bucket) DeepCopyInto(out *S3Bucket) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new S3Bucket.
func (in *S3Bucket) DeepCopy() *S3Bucket {
	if in == nil {
		return nil
	}
	out := new(S3Bucket)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageLocation) DeepCopyInto(out *StorageLocation) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageLocation.
func (in *StorageLocation) DeepCopy() *StorageLocation {
	if in == nil {
		return nil
	}
	out := new(StorageLocation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageLocation) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageLocationList) DeepCopyInto(out *StorageLocationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*StorageLocation, len(*in))
		for i := range *in {
			if (*in)[i] == nil {
				(*out)[i] = nil
			} ***REMOVED*** {
				(*out)[i] = new(StorageLocation)
				(*in)[i].DeepCopyInto((*out)[i])
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageLocationList.
func (in *StorageLocationList) DeepCopy() *StorageLocationList {
	if in == nil {
		return nil
	}
	out := new(StorageLocationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageLocationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageLocationSpec) DeepCopyInto(out *StorageLocationSpec) {
	*out = *in
	if in.S3 != nil {
		in, out := &in.S3, &out.S3
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(S3Bucket)
			**out = **in
		}
	}
	if in.Local != nil {
		in, out := &in.Local, &out.Local
		if *in == nil {
			*out = nil
		} ***REMOVED*** {
			*out = new(LocalStorage)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageLocationSpec.
func (in *StorageLocationSpec) DeepCopy() *StorageLocationSpec {
	if in == nil {
		return nil
	}
	out := new(StorageLocationSpec)
	in.DeepCopyInto(out)
	return out
}
