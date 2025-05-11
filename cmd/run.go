/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"venowolf/minlog/app/minlog/k8s"

	"github.com/spf13/cobra"
)

var runCmd *cobra.Command = &cobra.Command{
	Use:   "run",
	Short: "daemon process, which will run in the background",
	Long:  `Deployed as daemonset in kubenetes cluster as well`,
	Run: func(cmd *cobra.Command, args []string) {
		//l := log.GetLogger()
		//l.Sugar().Infof("%v", os.Args)
		//l.Sugar().Info("minlog starting...")
		// checking grafana-alloy health
		//create k8s client
		kc := k8s.NewKClient(labelNodeName, nameSpace, rPhaseOnly)
		//at first, get all pods in the node
		kc.Profilling(true)
		kc.Run(context.Background())
	},
}

func init() {
	parseflags(runCmd)
}

func GetRunCommand() *cobra.Command {
	return runCmd
}
