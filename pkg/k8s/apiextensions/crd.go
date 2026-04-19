package apiextensions

import (
	"encoding/json"
	"fmt"

	"github.com/LCSilent/helm-agent-eye/pkg/common"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Analyzer) AnalyzeCRD(r common.Request) (string, error) {
	kind := "CustomResourceDefinition"

	crdList, err := a.k8s.ApiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().List(r.Context, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, crd := range crdList.Items {
		var failures []common.Failure

		established := false
		namesAccepted := false
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1.Established:
				if cond.Status == apiextensionsv1.ConditionTrue {
					established = true
				} else {
					failures = append(failures, common.Failure{
						Text: fmt.Sprintf("CRD %s is not Established: %s", crd.Name, cond.Message),
					})
				}
			case apiextensionsv1.NamesAccepted:
				if cond.Status == apiextensionsv1.ConditionTrue {
					namesAccepted = true
				} else {
					failures = append(failures, common.Failure{
						Text: fmt.Sprintf("CRD %s names not accepted: %s", crd.Name, cond.Message),
					})
				}
			}
		}

		if !established && len(crd.Status.Conditions) == 0 {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("CRD %s has no status conditions, may still be processing", crd.Name),
			})
		}

		_ = namesAccepted

		if len(failures) > 0 {
			preAnalysis[crd.Name] = common.PreAnalysis{
				CRD:            crd,
				FailureDetails: failures,
			}
		}
	}

	results := make([]common.Result, 0)
	for key, value := range preAnalysis {
		results = append(results, common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		})
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
