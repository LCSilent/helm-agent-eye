package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	helmclient "github.com/LCSilent/helm-agent-eye/pkg/helm"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s"
	"github.com/spf13/cobra"
)

var helmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Helm release operations with K8s diagnostics",
}

var helmInstallCmd = &cobra.Command{
	Use:   "install-and-diagnose",
	Short: "Install a Helm release and diagnose K8s resources",
	Example: `  # Install from a local chart
  helm-agent-eye helm install-and-diagnose myapp ./charts/myapp -n myns

  # Install from a repo chart with custom values
  helm-agent-eye helm install-and-diagnose myapp stable/nginx -n myns --set replicas=2

  # Install without waiting for resources
  helm-agent-eye helm install-and-diagnose myapp ./charts/myapp -n myns --no-wait`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		releaseName := args[0]
		chart := args[1]

		namespace, _ := cmd.Flags().GetString("namespace")
		valuesStr, _ := cmd.Flags().GetString("values")
		setVals, _ := cmd.Flags().GetString("set")
		version, _ := cmd.Flags().GetString("version")
		noWait, _ := cmd.Flags().GetBool("no-wait")
		timeoutStr, _ := cmd.Flags().GetString("timeout")
		createNS, _ := cmd.Flags().GetBool("create-namespace")
		atomic, _ := cmd.Flags().GetBool("atomic")

		timeout := 5 * time.Minute
		if timeoutStr != "" {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				timeout = d
			}
		}

		opts := helmclient.InstallOptions{
			ReleaseName:     releaseName,
			Chart:           chart,
			Namespace:       namespace,
			Values:          valuesStr,
			SetValues:       setVals,
			Version:         version,
			Wait:            !noWait,
			Timeout:         timeout,
			CreateNamespace: createNS,
			Atomic:          atomic,
		}

		k8sClient, err := k8s.NewKubernetes()
		if err != nil {
			log.Fatalf("Failed to create k8s client: %v", err)
		}
		hc := helmclient.NewHelmClient(k8sClient, k8sClient.RestConfig())

		result, err := hc.InstallAndDiagnose(cmd.Context(), opts)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
		if result != nil {
			printHelmResult(result)
		}
	},
}

var helmUpgradeCmd = &cobra.Command{
	Use:   "upgrade-and-diagnose",
	Short: "Upgrade an existing Helm release and diagnose K8s resources",
	Example: `  # Upgrade with new chart version
  helm-agent-eye helm upgrade-and-diagnose myapp stable/nginx -n myns --version 1.2.0

  # Upgrade reusing existing values
  helm-agent-eye helm upgrade-and-diagnose myapp ./charts/myapp -n myns --reuse-values

  # Upgrade with cleanup on failure
  helm-agent-eye helm upgrade-and-diagnose myapp ./charts/myapp -n myns --cleanup-on-fail`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		releaseName := args[0]
		chart := args[1]

		namespace, _ := cmd.Flags().GetString("namespace")
		valuesStr, _ := cmd.Flags().GetString("values")
		setVals, _ := cmd.Flags().GetString("set")
		version, _ := cmd.Flags().GetString("version")
		noWait, _ := cmd.Flags().GetBool("no-wait")
		timeoutStr, _ := cmd.Flags().GetString("timeout")
		atomic, _ := cmd.Flags().GetBool("atomic")
		reuseValues, _ := cmd.Flags().GetBool("reuse-values")
		force, _ := cmd.Flags().GetBool("force")
		cleanupOnFail, _ := cmd.Flags().GetBool("cleanup-on-fail")

		timeout := 5 * time.Minute
		if timeoutStr != "" {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				timeout = d
			}
		}

		opts := helmclient.UpgradeOptions{
			ReleaseName:   releaseName,
			Chart:         chart,
			Namespace:     namespace,
			Values:        valuesStr,
			SetValues:     setVals,
			Version:       version,
			Wait:          !noWait,
			Timeout:       timeout,
			Atomic:        atomic,
			ReuseValues:   reuseValues,
			Force:         force,
			CleanupOnFail: cleanupOnFail,
		}

		k8sClient, err := k8s.NewKubernetes()
		if err != nil {
			log.Fatalf("Failed to create k8s client: %v", err)
		}
		hc := helmclient.NewHelmClient(k8sClient, k8sClient.RestConfig())

		result, err := hc.UpgradeAndDiagnose(cmd.Context(), opts)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
		if result != nil {
			printHelmResult(result)
		}
	},
}

func printHelmResult(result interface{}) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal result: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func init() {
	// install flags
	helmInstallCmd.Flags().StringP("namespace", "n", "default", "Target namespace")
	helmInstallCmd.Flags().StringP("values", "f", "", "Values as a YAML string")
	helmInstallCmd.Flags().String("set", "", "Set values on the command line (comma-separated key=value)")
	helmInstallCmd.Flags().String("version", "", "Chart version")
	helmInstallCmd.Flags().Bool("no-wait", false, "Do not wait for resources to be ready before diagnosing")
	helmInstallCmd.Flags().String("timeout", "5m", "Time to wait for resources to be ready")
	helmInstallCmd.Flags().Bool("create-namespace", false, "Create the namespace if it does not exist")
	helmInstallCmd.Flags().Bool("atomic", false, "Roll back on failure")

	// upgrade flags
	helmUpgradeCmd.Flags().StringP("namespace", "n", "default", "Target namespace")
	helmUpgradeCmd.Flags().StringP("values", "f", "", "Values as a YAML string")
	helmUpgradeCmd.Flags().String("set", "", "Set values on the command line (comma-separated key=value)")
	helmUpgradeCmd.Flags().String("version", "", "Chart version to upgrade to")
	helmUpgradeCmd.Flags().Bool("no-wait", false, "Do not wait for resources to be ready before diagnosing")
	helmUpgradeCmd.Flags().String("timeout", "5m", "Time to wait for resources to be ready")
	helmUpgradeCmd.Flags().Bool("atomic", false, "Roll back on failure")
	helmUpgradeCmd.Flags().Bool("reuse-values", false, "Reuse the last release's values")
	helmUpgradeCmd.Flags().Bool("force", false, "Force resource updates through a replacement strategy")
	helmUpgradeCmd.Flags().Bool("cleanup-on-fail", false, "Delete new resources created in this upgrade when upgrade fails")

	helmCmd.AddCommand(helmInstallCmd)
	helmCmd.AddCommand(helmUpgradeCmd)
	rootCmd.AddCommand(helmCmd)
}
