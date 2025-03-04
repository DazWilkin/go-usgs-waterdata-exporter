# Expected in the environment
local image = std.extVar("image");

local name = "exporter";
local namespace = "waterdata";

local labels = {
    "service": "usgs",
    "lang": "golang",
    "app": "exporter"
};

local port = 8080;

local sites = [
    "--sitecode=12149000", # Carnation
    "--sitecode=12150400", # Duvall
    "--sitecode=12150800"  # Monroe
    ];

{
    "apiVersion": "v1",
    "kind": "List",
    "metadata": {},
    "items": [
        {
            "apiVersion": "v1",
            "kind": "Namespace",
            "metadata": {
                "name": namespace,
                "labels": {
                    "kubernetes.io/metadata.name": namespace
                }
            }
        },
        {
            "apiVersion": "v1",
            "kind": "ServiceAccount",
            "metadata": {
                "name": name,
                "namespace": namespace
            }
        },
        {
            "apiVersion": "apps/v1",
            "kind": "Deployment",
            "metadata": {
                "name": name,
                "namespace": namespace,
                "labels": labels
            },
            "spec": {
                "replicas": 1,
                "selector": {
                    "matchLabels": labels
                },
                "template": {
                    "metadata": {
                        "labels": labels
                    },
                    "spec": {
                        "serviceAccountName": "exporter",
                        "containers": [
                            {
                                "name": name,
                                "image": image,
                                "imagePullPolicy": "IfNotPresent",
                                "command": [
                                    "/server"
                                ],
                                "args": [
                                    "--endpoint=0.0.0.0:" + port,
                                    "--path=/metrics"
                                ] + sites,
                                "ports": [
                                    {
                                        "name": "metrics",
                                        "protocol": "TCP",
                                        "containerPort": port
                                    }
                                ],
                                "livenessProbe": {
                                    "httpGet": {
                                        "path": "/healthz",
                                        "port": "metrics"
                                    },
                                    "initialDelaySeconds": 10
                                },
                                "resources": {
                                    "limits": {
                                        "memory": "500Mi"
                                    },
                                    "requests": {
                                        "cpu": "250m",
                                        "memory": "250Mi"
                                    }
                                },
                                "securityContext": {
                                    "allowPrivilegeEscalation": false,
                                    "privileged": false,
                                    "readOnlyRootFilesystem": true,
                                    "runAsNonRoot": true,
                                    "runAsUser": 1000,
                                    "runAsGroup": 1000
                                }
                            }
                        ],
                        "securityContext": {}
                    }
                }
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {
                "name": name,
                "namespace": namespace,
                "labels": labels
            },
            "spec": {
                "selector": labels,
                "ports": [
                    {
                        "name": "metrics",
                        "port": port,
                        "protocol": "TCP",
                        "targetPort": port
                    }
                ],
                "type": "ClusterIP"
            }
        },
        {
            "apiVersion": "networking.k8s.io/v1",
            "kind": "Ingress",
            "metadata": {
                "name": name,
                "namespace": namespace,
                "labels": labels
            },
            "spec": {
                "defaultBackend": {
                    "service": {
                        "name": name,
                        "port": {
                            "number": port
                        }
                    }
                },
                "ingressClassName": "tailscale",
                "tls": [
                    {
                        "hosts": [
                            "waterdata-exporter"
                        ]
                    }
                ]
            }
        },
        {
            "apiVersion": "monitoring.coreos.com/v1",
            "kind": "ServiceMonitor",
            "metadata": {
                "name": name,
                "namespace": namespace,
                "labels": labels
            },
            "spec": {
                "selector": {
                    "matchLabels": labels
                },
                "endpoints": [
                    {
                        "port": "metrics"
                    }
                ]
            }
        },
        {
            "apiVersion": "autoscaling.k8s.io/v1",
            "kind": "VerticalPodAutoscaler",
            "metadata": {
                "name": name,
                "namespace": namespace,
                "labels": labels
            },
            "spec": {
                "targetRef": {
                    "apiVersion": "apps/v1",
                    "kind": "Deployment",
                    "name": name
                },
                "updatePolicy": {
                    "updateMode": "Off"
                }
            }
        }
    ]
}
