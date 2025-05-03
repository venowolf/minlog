/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"venomouswolf/minlog/app/global"
	"venomouswolf/minlog/app/minlog/k8s"

	"github.com/spf13/cobra"
)

// runCmdHelper is a *cobra.Command, and holding the command line arguments
var (
	//labelNodeName used to specify the host label, if not specified, it will be default to the host name
	labelNodeName string

	//namesapce
	nameSpace string

	//rPhaseOnly, will only push the running-pod logs, default to true
	rPhaseOnly bool
	//loki endpoint
	lokiep string
)

var runCmd *cobra.Command = &cobra.Command{
	Use:   "run",
	Short: "daemon process, which will run in the background",
	Long:  `Deployed as daemonset in kubenetes cluster as well`,
	Run: func(cmd *cobra.Command, args []string) {
		//l := log.GetLogger()
		//l.Sugar().Infof("%v", os.Args)
		//l.Sugar().Info("minlog starting...")
		// checking grafana-agent-flow health
		//create k8s client
		kc := k8s.NewKClient(labelNodeName, nameSpace, lokiep, rPhaseOnly)
		//at first, get all pods in the node
		kc.Profilling()
		kc.Run(context.Background())
	},
}

func init() {
	runCmd.Flags().BoolVar(&rPhaseOnly, "running-only", true, "only push the running-pod logs")

	runCmd.Flags().StringVar(&labelNodeName, "label-nodename", "", "specify the node node name, pod.spec.nodeName")

	runCmd.Flags().StringVar(&nameSpace, "namespace", "", "specify the namespace, pod.metedata.namespace")

	runCmd.Flags().StringVar(&lokiep, "loki", "http://loki:3100/loki/api/v1/push", "loki api url, http://loki:3100/loki/api/v1/push")

	runCmd.Flags().StringVar(&global.RiverFile, "river-file", "/etc/grafana-agent-flow.river", "grafana-agent-flow.river, same config file with grafana-agent-flow process")

	runCmd.Flags().StringVar(&global.Gafurl, "agent-url", "http://127.0.0.1:12345", "grafana-agent-flow restful api")
}

func GetRunCommand() *cobra.Command {
	return runCmd
}
