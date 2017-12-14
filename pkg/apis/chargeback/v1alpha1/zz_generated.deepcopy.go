// +build !ignore_autogenerated

// This ***REMOVED***le was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

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
func (in *PrestoTable) DeepCopyInto(out *PrestoTable) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.State.DeepCopyInto(&out.State)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrestoTable.
func (in *PrestoTable) DeepCopy() *PrestoTable {
	if in == nil {
		return nil
	}
	out := new(PrestoTable)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PrestoTable) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrestoTableColumn) DeepCopyInto(out *PrestoTableColumn) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrestoTableColumn.
func (in *PrestoTableColumn) DeepCopy() *PrestoTableColumn {
	if in == nil {
		return nil
	}
	out := new(PrestoTableColumn)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrestoTableCreationParameters) DeepCopyInto(out *PrestoTableCreationParameters) {
	*out = *in
	if in.SerdeProps != nil {
		in, out := &in.SerdeProps, &out.SerdeProps
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Columns != nil {
		in, out := &in.Columns, &out.Columns
		*out = make([]PrestoTableColumn, len(*in))
		copy(*out, *in)
	}
	if in.Partitions != nil {
		in, out := &in.Partitions, &out.Partitions
		*out = make([]PrestoTableColumn, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrestoTableCreationParameters.
func (in *PrestoTableCreationParameters) DeepCopy() *PrestoTableCreationParameters {
	if in == nil {
		return nil
	}
	out := new(PrestoTableCreationParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrestoTableList) DeepCopyInto(out *PrestoTableList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*PrestoTable, len(*in))
		for i := range *in {
			if (*in)[i] == nil {
				(*out)[i] = nil
			} ***REMOVED*** {
				(*out)[i] = new(PrestoTable)
				(*in)[i].DeepCopyInto((*out)[i])
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrestoTableList.
func (in *PrestoTableList) DeepCopy() *PrestoTableList {
	if in == nil {
		return nil
	}
	out := new(PrestoTableList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PrestoTableList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrestoTablePartition) DeepCopyInto(out *PrestoTablePartition) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrestoTablePartition.
func (in *PrestoTablePartition) DeepCopy() *PrestoTablePartition {
	if in == nil {
		return nil
	}
	out := new(PrestoTablePartition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrestoTableState) DeepCopyInto(out *PrestoTableState) {
	*out = *in
	in.CreationParameters.DeepCopyInto(&out.CreationParameters)
	if in.Partitions != nil {
		in, out := &in.Partitions, &out.Partitions
		*out = make([]PrestoTablePartition, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrestoTableState.
func (in *PrestoTableState) DeepCopy() *PrestoTableState {
	if in == nil {
		return nil
	}
	out := new(PrestoTableState)
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
func (in *ReportDataSource) DeepCopyInto(out *ReportDataSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportDataSource.
func (in *ReportDataSource) DeepCopy() *ReportDataSource {
	if in == nil {
		return nil
	}
	out := new(ReportDataSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportDataSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportDataSourceList) DeepCopyInto(out *ReportDataSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*ReportDataSource, len(*in))
		for i := range *in {
			if (*in)[i] == nil {
				(*out)[i] = nil
			} ***REMOVED*** {
				(*out)[i] = new(ReportDataSource)
				(*in)[i].DeepCopyInto((*out)[i])
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportDataSourceList.
func (in *ReportDataSourceList) DeepCopy() *ReportDataSourceList {
	if in == nil {
		return nil
	}
	out := new(ReportDataSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportDataSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} ***REMOVED*** {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportDataSourceSpec) DeepCopyInto(out *ReportDataSourceSpec) {
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

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportDataSourceSpec.
func (in *ReportDataSourceSpec) DeepCopy() *ReportDataSourceSpec {
	if in == nil {
		return nil
	}
	out := new(ReportDataSourceSpec)
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
	if in.DataSources != nil {
		in, out := &in.DataSources, &out.DataSources
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
