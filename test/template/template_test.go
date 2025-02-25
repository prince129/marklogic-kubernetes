package template_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

func TestChartTemplateNoLogCollection(t *testing.T) {

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../charts")
	releaseName := "marklogic-test"
	t.Log(helmChartPath, releaseName)
	require.NoError(t, err)

	// Set up the namespace; confirm that the template renders the expected value for the namespace.
	namespaceName := "marklogic-" + strings.ToLower(random.UniqueId())
	t.Logf("Namespace: %s\n", namespaceName)

	// Setup the args for helm install
	options := &helm.Options{
		SetValues: map[string]string{
			"image.repository":      "marklogicdb/marklogic-db",
			"image.tag":             "latest",
			"persistence.enabled":   "false",
			"logCollection.enabled": "false",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	// render the tempate
	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/statefulset.yaml"})

	var statefulset appsv1.StatefulSet
	helm.UnmarshalK8SYaml(t, output, &statefulset)

	// Verify the name and namespace matches
	require.Equal(t, namespaceName, statefulset.Namespace)

	// Verify the image matches
	expectedImage := "marklogicdb/marklogic-db:latest"
	statefulSetContainers := statefulset.Spec.Template.Spec.Containers
	require.Equal(t, len(statefulSetContainers), 1)
	require.Equal(t, statefulSetContainers[0].Image, expectedImage)
}

func TestChartTemplateLogCollection(t *testing.T) {

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../charts")
	releaseName := "marklogic-test"
	t.Log(helmChartPath, releaseName)
	require.NoError(t, err)

	// Set up the namespace; confirm that the template renders the expected value for the namespace.
	namespaceName := "marklogic-" + strings.ToLower(random.UniqueId())

	// Setup the args for helm install
	options := &helm.Options{
		SetValues: map[string]string{
			"image.repository":      "marklogicdb/marklogic-db",
			"image.tag":             "latest",
			"persistence.enabled":   "true",
			"logCollection.enabled": "true",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	// render the tempate
	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/statefulset.yaml"})

	var statefulset appsv1.StatefulSet
	helm.UnmarshalK8SYaml(t, output, &statefulset)

	// Verify the name and namespace matches
	require.Equal(t, namespaceName, statefulset.Namespace)

	// Verify the image matches
	expectedImage1 := "marklogicdb/marklogic-db:latest"
	expectedImage2 := "fluent/fluent-bit:2.0.6"

	statefulSetContainers := statefulset.Spec.Template.Spec.Containers
	require.Equal(t, len(statefulSetContainers), 2)
	require.Equal(t, statefulSetContainers[0].Image, expectedImage1)
	require.Equal(t, statefulSetContainers[1].Image, expectedImage2)
}

func TestTemplatePersistence(t *testing.T) {

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../charts")
	releaseName := "marklogic-test"
	require.NoError(t, err)

	// Set up the namespace; confirm that the template renders the expected value for the namespace.
	namespaceName := "marklogic-" + strings.ToLower(random.UniqueId())

	additionalVolumeClaimTemplatesName := "log-volume"
	additionalVolumeClaimTemplatesAccessModes := "ReadWriteOnce"
	additionalVolumeClaimTemplatesResourcesRequestsStorage := "20Gi"

	// Setup the args for helm install
	options := &helm.Options{
		SetValues: map[string]string{
			"image.repository": "marklogicdb/marklogic-db",
			"image.tag":        "latest",
			"additionalVolumeClaimTemplates[0].metadata.name":                   additionalVolumeClaimTemplatesName,
			"additionalVolumeClaimTemplates[0].spec.accessModes[0]":             additionalVolumeClaimTemplatesAccessModes,
			"additionalVolumeClaimTemplates[0].spec.resources.requests.storage": additionalVolumeClaimTemplatesResourcesRequestsStorage,
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	// render the tempate
	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/statefulset.yaml"})

	var statefulset appsv1.StatefulSet
	helm.UnmarshalK8SYaml(t, output, &statefulset)

	// Verify default persistent volume claim
	require.Equal(t, len(statefulset.Spec.VolumeClaimTemplates), 2)
	require.Equal(t, string(statefulset.Spec.VolumeClaimTemplates[0].Name), "datadir")
	require.Equal(t, string(statefulset.Spec.VolumeClaimTemplates[0].Spec.AccessModes[0]), "ReadWriteOnce")
	require.Equal(t, statefulset.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage().String(), "10Gi")

	// Verify the additional volume claim
	require.Equal(t, string(statefulset.Spec.VolumeClaimTemplates[1].Name), additionalVolumeClaimTemplatesName)
	require.Equal(t, string(statefulset.Spec.VolumeClaimTemplates[1].Spec.AccessModes[0]), additionalVolumeClaimTemplatesAccessModes)
	require.Equal(t, statefulset.Spec.VolumeClaimTemplates[1].Spec.Resources.Requests.Storage().String(), additionalVolumeClaimTemplatesResourcesRequestsStorage)
}
