package v1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Cron struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
}

type CronList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*Cron `json:"items"`
}
