package defaults

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	// DaemonResources map contains the default resource requirements for the
	// various OCS daemons
	DaemonResources = map[string]corev1.ResourceRequirements{
		"osd": corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
		},
		"mon": corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
		},
		"mds": corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
		},
		"rgw": corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
		},
		"mgr": corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
		},
		"noobaa-core": corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
		},
		"noobaa-db": corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("0Gi"),
			},
		},
		"noobaa-db-vol": corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("50Gi"),
			},
		},
	}
)
