package networking

import (
	"encoding/json"
	"fmt"

	"github.com/LCSilent/helm-agent-eye/pkg/common"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// AnalyzeNetworkPolicy analyzes the network policies and returns a list of failures.
func (a *Analyzer) AnalyzeNetworkPolicy(r common.Request) (string, error) {
	kind := "NetworkPolicy"
	apiDoc := base.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "networking.k8s.io",
			Version: "v1",
		},
		OpenapiSchema: a.k8s.OpenapiSchema,
	}

	policyList, err := a.k8s.Clientset.NetworkingV1().NetworkPolicies(r.Namespace).List(r.Context, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, policy := range policyList.Items {
		var failures []common.Failure

		// Check if the policy has no pod selector
		if len(policy.Spec.PodSelector.MatchLabels) == 0 {
			doc := apiDoc.GetApiDocV2("spec.podSelector")
			failures = append(failures, common.Failure{
				Text:          fmt.Sprint("NetworkPolicy has empty pod selector, will select all pods"),
				KubernetesDoc: doc,
			})

			for _, policyType := range policy.Spec.PolicyTypes {
				switch policyType {
				case "Ingress":
					if len(policy.Spec.Ingress) == 0 {
						doc := apiDoc.GetApiDocV2("spec.ingress")
						failures = append(failures, common.Failure{
							Text:          fmt.Sprint("NetworkPolicy will deny all ingress traffic"),
							KubernetesDoc: doc,
						})
					}
				case "Egress":
					if len(policy.Spec.Egress) == 0 {
						doc := apiDoc.GetApiDocV2("spec.egress")
						failures = append(failures, common.Failure{
							Text:          fmt.Sprint("NetworkPolicy will deny all egress traffic"),
							KubernetesDoc: doc,
						})
					}
				}
			}
		} else {
			podList, err := a.k8s.Clientset.CoreV1().Pods(policy.Namespace).List(r.Context, metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
					MatchLabels: policy.Spec.PodSelector.MatchLabels,
				}),
			})
			if err != nil {
				return "", err
			}
			if len(podList.Items) == 0 {
				doc := apiDoc.GetApiDocV2("spec.podSelector")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprint("NetworkPolicy has no matching pods"),
					KubernetesDoc: doc,
				})
			}
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)] = common.PreAnalysis{
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
		results = append(results, result)
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
