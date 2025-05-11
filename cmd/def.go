package cmd

import (
	"venowolf/minlog/app/global"

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

func parseflags(c *cobra.Command) {

	c.Flags().BoolVar(&rPhaseOnly, "running-only", true, "only push the running-pod logs")

	c.Flags().StringVar(&labelNodeName, "label-nodename", "", "specify the node node name, pod.spec.nodeName")

	c.Flags().StringVar(&nameSpace, "namespace", "", "specify the namespace, pod.metedata.namespace")

	c.Flags().StringVar(&lokiep, "loki", global.GetFromEnvOrDefaultValue("LOKIEP"), "loki api url, http://loki:3100/loki/api/v1/push")

	c.Flags().StringVar(&global.AlloyFile, "alloy-file", global.GetFromEnvOrDefaultValue(""), "config.alloy, same config file with grafana-alloy process")

	c.Flags().StringVar(&global.GAlloyurl, "alloy-url", global.GetFromEnvOrDefaultValue("ALLOYURL"), "grafana-alloy restful api")
}
