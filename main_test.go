package main

import (
  "testing"

  "github.com/maxatome/go-testdeep/td"
  "github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
  result := app([]byte(`
image: nginx
`), AppOptions{Search: "image"})
  require.Empty(t, result.Error)
  td.Cmp(t, 0, result.ExitCode)
  td.Cmp(t, result.Yaml, ".image: nginx\n")
}

func TestSpecImage(t *testing.T) {
  result := app([]byte(`
spec:
  image: nginx
`), AppOptions{Search: "image"})
  require.Empty(t, result.Error)
  td.Cmp(t, 0, result.ExitCode)
  td.Cmp(t, result.Yaml, ".spec.image: nginx\n")
}

func TestService(t *testing.T) {
  result := app([]byte(`
apiVersion: v1
kind: Service
metadata:
  labels:
    app: selenium
    release: selenium
  name: selenium
spec:
  ports:
  - name: http
    port: 80
    targetPort: http
  selector:
    app: selenium
    release: selenium

`), AppOptions{Search: "labels"})
  require.Empty(t, result.Error)
  td.Cmp(t, 0, result.ExitCode)
  td.Cmp(t, result.Yaml, `.metadata.labels:
  app: selenium
  release: selenium
`)
}

func TestSlice(t *testing.T) {
  result := app([]byte(`
apiVersion: v1
items:
- apiVersion: v1
  kind: Service
  metadata:
    labels:
      app: selenium
      release: selenium
    name: selenium
  spec:
    ports:
    - name: http
      port: 80
      targetPort: http
    selector:
      app: selenium
      release: selenium
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
`), AppOptions{Search: "labels"})
  require.Empty(t, result.Error)
  td.Cmp(t, 0, result.ExitCode)
  td.Cmp(t, result.Yaml, `.items[0].metadata.labels:
  app: selenium
  release: selenium
`)
}

func TestPodImage(t *testing.T) {
  result := app([]byte(`
apiVersion: v1
kind: Pod
metadata:
  name: selenium-chrome
spec:
  containers:
  - env:
    - name: PORT
      value: "4444"
    - name: JAVA_OPTS
      value: -Xmx3072m
    - name: START_XVFB
      value: "false"
    image: selenium/standalone-chrome:3
    imagePullPolicy: IfNotPresent
`), AppOptions{Search: "image", ExactMatch: true})
  require.Empty(t, result.Error)
  td.Cmp(t, 0, result.ExitCode)
  td.Cmp(t, result.Yaml, ".spec.containers[0].image: selenium/standalone-chrome:3\n")
}

func TestNodeKernelVersion(t *testing.T) {
  result := app([]byte(`
apiVersion: v1
kind: Node
metadata:
  name: my-node
status:
  nodeInfo:
    kernelVersion: 4.19.0-11-amd64
`), AppOptions{Search: "ker", ShowStatus: true})
  require.Empty(t, result.Error)
  td.Cmp(t, 0, result.ExitCode)
  td.Cmp(t, result.Yaml, ".status.nodeInfo.kernelVersion: 4.19.0-11-amd64\n")
}

func TestPodNodeAffinity(t *testing.T) {
  result := app([]byte(`
apiVersion: v1
kind: Pod
metadata:
  name: selenium-chrome
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: availability/csssr.com
            operator: In
            values:
            - available
          - key: availability/csssr.express
            operator: In
            values:
            - available
  containers:
  - env:
    - name: PORT
      value: "4444"
    - name: JAVA_OPTS
      value: -Xmx3072m
    - name: START_XVFB
      value: "false"
    image: selenium/standalone-chrome:3
    imagePullPolicy: IfNotPresent
`), AppOptions{Search: "nodeAff"})
  require.Empty(t, result.Error)
  td.Cmp(t, 0, result.ExitCode)
  td.Cmp(t, result.Yaml, `.spec.affinity.nodeAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
    nodeSelectorTerms:
    - matchExpressions:
      - key: availability/csssr.com
        operator: In
        values:
        - available
      - key: availability/csssr.express
        operator: In
        values:
        - available
`)
}

func TestPodNodeAffinityIgnoreCase(t *testing.T) {
  result := app([]byte(`
apiVersion: v1
kind: Pod
metadata:
  name: selenium-chrome
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: availability/csssr.com
            operator: In
            values:
            - available
          - key: availability/csssr.express
            operator: In
            values:
            - available
  containers:
  - env:
    - name: PORT
      value: "4444"
    - name: JAVA_OPTS
      value: -Xmx3072m
    - name: START_XVFB
      value: "false"
    image: selenium/standalone-chrome:3
    imagePullPolicy: IfNotPresent
`), AppOptions{Search: "nodeaff", IgnoreCase: true})
  require.Empty(t, result.Error)
  td.Cmp(t, 0, result.ExitCode)
  td.Cmp(t, result.Yaml, `.spec.affinity.nodeAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
    nodeSelectorTerms:
    - matchExpressions:
      - key: availability/csssr.com
        operator: In
        values:
        - available
      - key: availability/csssr.express
        operator: In
        values:
        - available
`)
}
