package storagecluster

import (
	"bytes"
	"context"
	"github.com/go-logr/logr"
	consolev1 "github.com/openshift/api/console/v1"
	ocsv1 "github.com/openshift/ocs-operator/pkg/apis/ocs/v1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
	"path"
)

var (
	quickstartDir = "/quickstarts/"
)

func (r *ReconcileStorageCluster) ensureQuickStarts(sc *ocsv1.StorageCluster, reqLogger logr.Logger) error {
	allQuickStarts, err := FetchQuickStarts(quickstartDir)
	if err != nil {
		reqLogger.Error(err, "Failed to parse quickstarts")
		return nil
	}
	if len(allQuickStarts) == 0 {
		reqLogger.Info("No quickstarts found")
		return nil
	}
	for _, qs := range allQuickStarts {
		found := consolev1.ConsoleQuickStart{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: qs.Name, Namespace: qs.Namespace}, &found)
		if err != nil {
			if errors.IsNotFound(err) {
				err = r.client.Create(context.TODO(), &qs)
				if err != nil {
					reqLogger.Error(err, "Failed to create quickstart", "Name", qs.Name, "Namespace", qs.Namespace)
					return nil
				}
				reqLogger.Info("Creating quickstarts", "Name", qs.Name, "Namespace", qs.Namespace)
				continue
			}
			reqLogger.Error(err, "Error has occurred when fetching quickstarts")
			return nil
		}
		found.Spec = qs.Spec
		err = r.client.Update(context.TODO(), &found)
		if err != nil {
			reqLogger.Error(err, "Failed to update quickstart", "Name", qs.Name, "Namespace", qs.Namespace)
			return nil
		}
		reqLogger.Info("Updating quickstarts", "Name", qs.Name, "Namespace", qs.Namespace)
	}
	return nil
}

func FetchQuickStarts(dir string) ([]consolev1.ConsoleQuickStart, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	quickstarts := []consolev1.ConsoleQuickStart{}
	for _, f := range files {
		qsData, err := ioutil.ReadFile(path.Join(dir, f.Name()))
		if err != nil {
			return nil, err
		}
		qs := consolev1.ConsoleQuickStart{}
		if err := k8sYAML.NewYAMLOrJSONDecoder(bytes.NewBuffer(qsData), 1000).Decode(&qs); err != nil {
			return nil, err
		}
		quickstarts = append(quickstarts, qs)
	}
	return quickstarts, nil
}
