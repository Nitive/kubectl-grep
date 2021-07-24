package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	result := app([]byte(`
image: nginx
`), "image")
	require.Empty(t, result.Error)
	require.Equal(t, result.ExitCode, 0)
	require.Equal(t, ".image: nginx\n", result.Yaml)
}

func TestSpecImage(t *testing.T) {
	result := app([]byte(`
spec:
  image: nginx
`), "image")
	require.Empty(t, result.Error)
	require.Equal(t, result.ExitCode, 0)
	require.Equal(t, ".spec.image: nginx\n", result.Yaml)
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

`), "labels")
	require.Empty(t, result.Error)
	require.Equal(t, result.ExitCode, 0)
	require.Equal(t, `.metadata.labels:
  app: selenium
  release: selenium
`, result.Yaml)
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
`), "labels")
	require.Empty(t, result.Error)
	require.Equal(t, result.ExitCode, 0)
	require.Equal(t, `.items[0].metadata.labels:
  app: selenium
  release: selenium
`, result.Yaml)
}
