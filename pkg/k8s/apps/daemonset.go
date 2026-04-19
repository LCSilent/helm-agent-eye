package apps

import (
	"encoding/json"
	"fmt"

	"github.com/LCSilent/helm-agent-eye/pkg/common"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/base"
	"github.com/LCSilent/helm-agent-eye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (a *Analyzer) AnalyzeDaemonSet(r common.Request) (string, error) {
	kind := "DaemonSet"
	apiDoc := base.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "apps",
			Version: "v1",
		},
		OpenapiSchema: a.k8s.OpenapiSchema,
	}

	dsList, err := a.k8s.Clientset.AppsV1().DaemonSets(r.Namespace).List(r.Context, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, ds := range dsList.Items {
		var failures []common.Failure

		// Check if desired number of pods are scheduled and ready
		if ds.Status.DesiredNumberScheduled != ds.Status.NumberReady {
			doc := apiDoc.GetApiDocV2("status.numberReady")
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf(
					"DaemonSet %s/%s has %d desired pods but only %d are ready",
					ds.Namespace, ds.Name,
					ds.Status.DesiredNumberScheduled,
					ds.Status.NumberReady,
				),
				KubernetesDoc: doc,
			})
		}

		// Check for pods that are not scheduled due to node selector / tolerations
		if ds.Status.NumberMisscheduled > 0 {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf(
					"DaemonSet %s/%s has %d misscheduled pods",
					ds.Namespace, ds.Name,
					ds.Status.NumberMisscheduled,
				),
			})
		}

		// Check update rollout stuck
		if ds.Status.UpdatedNumberScheduled < ds.Status.DesiredNumberScheduled {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf(
					"DaemonSet %s/%s rollout in progress: %d/%d pods updated",
					ds.Namespace, ds.Name,
					ds.Status.UpdatedNumberScheduled,
					ds.Status.DesiredNumberScheduled,
				),
			})
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", ds.Namespace, ds.Name)] = common.PreAnalysis{
				DaemonSet:      ds,
				FailureDetails: failures,
			}
		}
	}

	results := make([]common.Result, 0)
	for key, value := range preAnalysis {
		result := common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}
		parent, found := utils.GetParent(a.k8s.Clientset, value.DaemonSet.ObjectMeta)
		if found {
			result.ParentObject = parent
		}
		results = append(results, result)
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
