package k8s

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

// --- Pod status override branches (not covered by existing tests) ---

func TestPopulate_PodInitContainerFailureOverride(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"name": "app"},
			},
		},
		"status": map[string]interface{}{
			"phase": "Pending",
			"containerStatuses": []interface{}{
				map[string]interface{}{
					"name":         "app",
					"ready":        false,
					"restartCount": float64(0),
					"state": map[string]interface{}{
						"waiting": map[string]interface{}{
							"reason": "PodInitializing",
						},
					},
				},
			},
			"initContainerStatuses": []interface{}{
				map[string]interface{}{
					"name":  "init",
					"ready": false,
					"state": map[string]interface{}{
						"terminated": map[string]interface{}{
							"reason": "Error",
						},
					},
				},
			},
		},
	}
	ti := model.Item{Status: "Pending"}
	populateResourceDetails(&ti, obj, "Pod")
	assert.Equal(t, "Error", ti.Status)
}

func TestPopulate_PodRunningNotReadyBecomesNotReady(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"name": "app"},
				map[string]interface{}{"name": "sidecar"},
			},
		},
		"status": map[string]interface{}{
			"phase": "Running",
			"containerStatuses": []interface{}{
				map[string]interface{}{
					"name":         "app",
					"ready":        true,
					"restartCount": float64(0),
				},
				map[string]interface{}{
					"name":         "sidecar",
					"ready":        false,
					"restartCount": float64(0),
				},
			},
		},
	}
	ti := model.Item{Status: "Running"}
	populateResourceDetails(&ti, obj, "Pod")
	assert.Equal(t, "NotReady", ti.Status)
}

func TestPopulate_PodSucceededKeepsStatus(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"name": "app"},
			},
		},
		"status": map[string]interface{}{
			"phase": "Succeeded",
			"containerStatuses": []interface{}{
				map[string]interface{}{
					"name":         "app",
					"ready":        false,
					"restartCount": float64(0),
				},
			},
		},
	}
	ti := model.Item{Status: "Succeeded"}
	populateResourceDetails(&ti, obj, "Pod")
	assert.Equal(t, "Succeeded", ti.Status)
}

func TestPopulate_PodFailedPreferredOverPodInitializing(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"name": "app"},
			},
		},
		"status": map[string]interface{}{
			"phase": "Failed",
			"containerStatuses": []interface{}{
				map[string]interface{}{
					"name":         "app",
					"ready":        false,
					"restartCount": float64(0),
					"state": map[string]interface{}{
						"waiting": map[string]interface{}{
							"reason": "PodInitializing",
						},
					},
				},
			},
			"initContainerStatuses": []interface{}{
				map[string]interface{}{
					"name":  "init",
					"ready": false,
					"state": map[string]interface{}{
						"waiting": map[string]interface{}{
							"reason": "PodInitializing",
						},
					},
				},
			},
		},
	}
	ti := model.Item{Status: "Failed"}
	populateResourceDetails(&ti, obj, "Pod")
	// PodInitializing + Failed => reason cleared, status stays "Failed"
	// because the else-if only sets NotReady when status is "Running".
	assert.Equal(t, "Failed", ti.Status)
}

func TestPopulate_PodNilStatusReturnsEarly(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"name": "app"},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Pod")
	assert.Empty(t, ti.Ready)
}

func TestPopulate_PodRestartCountInt64(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"name": "app"},
			},
		},
		"status": map[string]interface{}{
			"containerStatuses": []interface{}{
				map[string]interface{}{
					"name":         "app",
					"ready":        true,
					"restartCount": int64(5),
				},
			},
		},
	}
	ti := model.Item{Status: "Running"}
	populateResourceDetails(&ti, obj, "Pod")
	assert.Equal(t, "5", ti.Restarts)
}

// --- Service: additional branches ---

func TestPopulate_ServiceLoadBalancerHostname(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"type": "LoadBalancer",
		},
		"status": map[string]interface{}{
			"loadBalancer": map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{"hostname": "my-lb.example.com"},
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Service")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "my-lb.example.com", colMap["External Address"])
}

func TestPopulate_ServiceExternalIPs(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"type":        "ClusterIP",
			"externalIPs": []interface{}{"1.2.3.4", "5.6.7.8"},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Service")

	colMap := columnsToMap(ti.Columns)
	assert.Contains(t, colMap["External IPs"], "1.2.3.4")
	assert.Contains(t, colMap["External IPs"], "5.6.7.8")
}

func TestPopulate_ServiceSessionAffinityClientIP(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"type":            "ClusterIP",
			"sessionAffinity": "ClientIP",
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Service")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "ClientIP", colMap["Session Affinity"])
}

func TestPopulate_ServiceSessionAffinityNoneOmitted(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"type":            "ClusterIP",
			"sessionAffinity": "None",
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Service")

	colMap := columnsToMap(ti.Columns)
	_, found := colMap["Session Affinity"]
	assert.False(t, found)
}

func TestPopulate_ServiceNilSpec(t *testing.T) {
	obj := map[string]interface{}{}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Service")
	assert.Empty(t, ti.Columns)
}

// --- Ingress: additional branches ---

func TestPopulate_IngressDefaultBackendPortNumber(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"defaultBackend": map[string]interface{}{
				"service": map[string]interface{}{
					"name": "backend-svc",
					"port": map[string]interface{}{
						"number": float64(8080),
					},
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Ingress")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "backend-svc:8080", colMap["Default Backend"])
}

func TestPopulate_IngressDefaultBackendPortName(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"defaultBackend": map[string]interface{}{
				"service": map[string]interface{}{
					"name": "api-svc",
					"port": map[string]interface{}{
						"name": "http",
					},
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Ingress")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "api-svc:http", colMap["Default Backend"])
}

func TestPopulate_IngressDefaultBackendNoPort(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"defaultBackend": map[string]interface{}{
				"service": map[string]interface{}{
					"name": "simple-svc",
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Ingress")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "simple-svc", colMap["Default Backend"])
}

func TestPopulate_IngressTLSAndURL(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{
					"host": "app.example.com",
					"http": map[string]interface{}{
						"paths": []interface{}{
							map[string]interface{}{
								"path": "/api",
							},
						},
					},
				},
			},
			"tls": []interface{}{
				map[string]interface{}{
					"hosts": []interface{}{"app.example.com"},
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Ingress")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "https://app.example.com/api", colMap["__ingress_url"])
	assert.Contains(t, colMap["TLS Hosts"], "app.example.com")
}

func TestPopulate_IngressHTTPUrlNoTLS(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{
					"host": "plain.example.com",
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Ingress")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "http://plain.example.com", colMap["__ingress_url"])
}

func TestPopulate_IngressLBHostname(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{},
		"status": map[string]interface{}{
			"loadBalancer": map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{"hostname": "lb.aws.com"},
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Ingress")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "lb.aws.com", colMap["Address"])
}

// --- ConfigMap ---

func TestPopulate_ConfigMapData(t *testing.T) {
	obj := map[string]interface{}{
		"data": map[string]interface{}{
			"config.yaml": "key: value\n",
			"app.conf":    "setting=1",
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "ConfigMap")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "setting=1", colMap["data:app.conf"])
	assert.Equal(t, "key: value\n", colMap["data:config.yaml"])
}

// --- Secret ---

func TestPopulate_SecretBase64(t *testing.T) {
	obj := map[string]interface{}{
		"type": "Opaque",
		"data": map[string]interface{}{
			"password": base64.StdEncoding.EncodeToString([]byte("s3cr3t")),
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Secret")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "s3cr3t", colMap["secret:password"])
	assert.Equal(t, "Opaque", colMap["Type"])
}

func TestPopulate_SecretInvalidBase64Skipped(t *testing.T) {
	obj := map[string]interface{}{
		"type": "Opaque",
		"data": map[string]interface{}{
			"broken": "!!!not-valid-base64!!!",
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Secret")

	colMap := columnsToMap(ti.Columns)
	_, found := colMap["secret:broken"]
	assert.False(t, found)
}

// --- Node ---

func TestPopulate_NodeRolesAndTaints(t *testing.T) {
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				"node-role.kubernetes.io/control-plane": "",
				"node-role.kubernetes.io/worker":        "",
			},
		},
		"spec": map[string]interface{}{
			"taints": []interface{}{
				map[string]interface{}{
					"key":    "node-role.kubernetes.io/control-plane",
					"effect": "NoSchedule",
				},
				map[string]interface{}{
					"key":    "dedicated",
					"value":  "gpu",
					"effect": "NoExecute",
				},
			},
		},
		"status": map[string]interface{}{
			"addresses": []interface{}{
				map[string]interface{}{"type": "InternalIP", "address": "10.0.0.5"},
			},
			"allocatable": map[string]interface{}{
				"cpu":    "4",
				"memory": "8Gi",
			},
			"nodeInfo": map[string]interface{}{
				"kubeletVersion":          "v1.29.0",
				"osImage":                 "Ubuntu 22.04",
				"containerRuntimeVersion": "containerd://1.7.2",
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Node")

	colMap := columnsToMap(ti.Columns)
	assert.Contains(t, colMap["Role"], "control-plane")
	assert.Contains(t, colMap["Role"], "worker")
	assert.Equal(t, "10.0.0.5", colMap["InternalIP"])
	assert.Equal(t, "4", colMap["CPU Alloc"])
	assert.Equal(t, "8Gi", colMap["Mem Alloc"])
	assert.Equal(t, "v1.29.0", colMap["Version"])
	assert.Equal(t, "Ubuntu 22.04", colMap["OS"])
	assert.Equal(t, "containerd://1.7.2", colMap["Runtime"])
	assert.Contains(t, colMap["Taints"], "dedicated=gpu:NoExecute")
}

func TestPopulate_NodeEmptyRoleSuffix(t *testing.T) {
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				"node-role.kubernetes.io/": "",
			},
		},
		"status": map[string]interface{}{},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Node")

	colMap := columnsToMap(ti.Columns)
	_, found := colMap["Role"]
	assert.False(t, found)
}

// --- PVC ---

func TestPopulate_PVCBound(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"storage": "10Gi",
				},
			},
			"volumeName":       "pv-123",
			"accessModes":      []interface{}{"ReadWriteOnce"},
			"storageClassName": "standard",
			"volumeMode":       "Filesystem",
		},
		"status": map[string]interface{}{
			"phase": "Bound",
			"capacity": map[string]interface{}{
				"storage": "10Gi",
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "PersistentVolumeClaim")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "Bound", ti.Status)
	assert.Equal(t, "10Gi", colMap["Capacity"])
	assert.Equal(t, "10Gi", colMap["Request"])
	assert.Equal(t, "pv-123", colMap["Volume"])
	assert.Equal(t, "ReadWriteOnce", colMap["Access Modes"])
	assert.Equal(t, "standard", colMap["Storage Class"])
	assert.Equal(t, "Filesystem", colMap["Volume Mode"])
}

// --- CronJob ---

func TestPopulate_CronJobFields(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"schedule": "*/5 * * * *",
			"suspend":  false,
		},
		"status": map[string]interface{}{
			"lastScheduleTime": "2026-03-22T10:00:00Z",
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "CronJob")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "*/5 * * * *", colMap["Schedule"])
	assert.Equal(t, "false", colMap["Suspend"])
	assert.Equal(t, "2026-03-22T10:00:00Z", colMap["Last Schedule"])
}

// --- Job ---

func TestPopulate_JobZeroFailuresOmitted(t *testing.T) {
	obj := map[string]interface{}{
		"status": map[string]interface{}{
			"failed":    float64(0),
			"succeeded": float64(1),
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "Job")

	colMap := columnsToMap(ti.Columns)
	_, found := colMap["Failed"]
	assert.False(t, found)
}

// --- HPA: additional metric type branches ---

func TestPopulate_HPAPodsMetric(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"maxReplicas": float64(5),
			"metrics": []interface{}{
				map[string]interface{}{
					"type": "Pods",
					"pods": map[string]interface{}{
						"metric": map[string]interface{}{
							"name": "requests_per_second",
						},
						"target": map[string]interface{}{
							"averageValue": "100",
						},
					},
				},
			},
		},
		"status": map[string]interface{}{
			"currentReplicas": float64(3),
			"desiredReplicas": float64(3),
			"currentMetrics": []interface{}{
				map[string]interface{}{
					"type": "Pods",
					"pods": map[string]interface{}{
						"metric": map[string]interface{}{
							"name": "requests_per_second",
						},
						"current": map[string]interface{}{
							"averageValue": "85",
						},
					},
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "HorizontalPodAutoscaler")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "100", colMap["Target requests_per_second"])
	assert.Equal(t, "85", colMap["Current requests_per_second"])
}

func TestPopulate_HPAObjectMetric(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"maxReplicas": float64(10),
			"metrics": []interface{}{
				map[string]interface{}{
					"type": "Object",
					"object": map[string]interface{}{
						"metric": map[string]interface{}{
							"name": "queue_depth",
						},
						"target": map[string]interface{}{
							"value": "50",
						},
					},
				},
			},
		},
		"status": map[string]interface{}{
			"currentReplicas": float64(1),
			"desiredReplicas": float64(1),
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "HorizontalPodAutoscaler")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "50", colMap["Target queue_depth"])
}

func TestPopulate_HPANonMapMetricSkipped(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"maxReplicas": float64(5),
			"metrics": []interface{}{
				"not-a-map",
			},
		},
		"status": map[string]interface{}{
			"currentReplicas": float64(1),
			"desiredReplicas": float64(1),
			"currentMetrics": []interface{}{
				"not-a-map",
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "HorizontalPodAutoscaler")
	assert.Equal(t, fmt.Sprintf("%d/%d (%d-%d)", 1, 1, 0, 5), ti.Ready)
}

func TestPopulate_HPAScalingLimitedFalseIgnored(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"maxReplicas": float64(5),
		},
		"status": map[string]interface{}{
			"currentReplicas": float64(1),
			"desiredReplicas": float64(1),
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "ScalingLimited",
					"status": "False",
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "HorizontalPodAutoscaler")

	colMap := columnsToMap(ti.Columns)
	_, found := colMap["Scaling Limited"]
	assert.False(t, found)
}

func TestPopulate_HPAResourceAverageValue(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"maxReplicas": float64(5),
			"metrics": []interface{}{
				map[string]interface{}{
					"type": "Resource",
					"resource": map[string]interface{}{
						"name": "memory",
						"target": map[string]interface{}{
							"type":         "AverageValue",
							"averageValue": "500Mi",
						},
					},
				},
			},
		},
		"status": map[string]interface{}{
			"currentReplicas": float64(1),
			"desiredReplicas": float64(1),
			"currentMetrics": []interface{}{
				map[string]interface{}{
					"type": "Resource",
					"resource": map[string]interface{}{
						"name": "memory",
						"current": map[string]interface{}{
							"averageValue": "256Mi",
						},
					},
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "HorizontalPodAutoscaler")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "500Mi", colMap["Target Memory"])
	assert.Equal(t, "256Mi", colMap["Current Memory"])
}

// --- Unknown kind falls through to ext ---

func TestPopulate_UnknownKindFallsToExt(t *testing.T) {
	obj := map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "True",
				},
			},
		},
	}
	ti := model.Item{}
	populateResourceDetails(&ti, obj, "CustomWidget")

	colMap := columnsToMap(ti.Columns)
	assert.Equal(t, "True", colMap["Ready"])
}
