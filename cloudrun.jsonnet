local image = std.extVar("image");
local region = std.extVar("region");

local name = "exporter";

local port = 8080;

local sitecodes = [
    "--sitecode=12149000", # Carnation
    "--sitecode=12150400", # Duvall
    "--sitecode=12150800"  # Monroe
];

{
  "apiVersion": "serving.knative.dev/v1",
  "kind": "Service",
  "metadata": {
    "name": name,
    "annotations": {
      "run.googleapis.com/build-enable-automatic-updates": "false",
      "run.googleapis.com/ingress": "all",
      "run.googleapis.com/ingress-status": "all"
    },
    "labels": {
      "cloud.googleapis.com/location": region
    }
  },
  "spec": {
    "template": {
      "metadata": {
        "annotations": {
          "autoscaling.knative.dev/maxScale": "1"
        },
        "labels": {
          "run.googleapis.com/startupProbeType": "Default"
        }
      },
      "spec": {
        "containerConcurrency": 80,
        "containers": [
          {
            "args": [
              "--endpoint=:" + port,
              "--path=/metrics"
            ] + sitecodes,
            "image": image,
            "ports": [
              {
                "containerPort": port,
                "name": "http1"
              }
            ],
            "resources": {
              "limits": {
                "cpu": "1000m",
                "memory": "512Mi"
              }
            },
            "startupProbe": {
              "failureThreshold": 1,
              "periodSeconds": 240,
              "tcpSocket": {
                "port": port
              },
              "timeoutSeconds": 240
            },
            "livenessProbe": {
              "failureThreshold": 1,
              "httpGet": {
                "path": "/healthz",
                "port": port
              },
              "periodSeconds": 240,
              "timeoutSeconds": 240
            }
          }
        ]
      }
    }
  }
}