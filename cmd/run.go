/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"venomouswolf/minlog/app/log"
	"venomouswolf/minlog/app/minlog/k8s"

	"github.com/spf13/cobra"
)

var runCmd *cobra.Command = nil
var rh *runHelper = nil

func init() {
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "daemon process, which will run in the background",
		Long:  `Deployed as daemonset in kubenetes cluster as well`,
		RunE: func(cmd *cobra.Command, args []string) error {
			l := log.GetLogger()
			l.Info(fmt.Sprintf("%v", os.Args))
			l.Info("minlog starting...")
			return rh.start()
		},
	}

	rh = &runHelper{
		runCmd: runCmd,
	}

	rh.parseFlags()

	rootCmd.AddCommand(runCmd)
}

// runCmdHelper is a *cobra.Command, and holding the command line arguments
type runHelper struct {
	runCmd *cobra.Command

	//labelNodeName used to specify the host label, if not specified, it will be default to the host name
	labelNodeName string

	//namesapce
	nameSpace string

	//rPhaseOnly, will only push the running-pod logs, default to true
	rPhaseOnly bool
}

func (r *runHelper) parseFlags() {

	r.runCmd.Flags().BoolVarP(&r.rPhaseOnly, "running-pod-only", "r", true, "only push the running-pod logs")

	r.runCmd.Flags().StringVarP(&r.labelNodeName, "label-nodename", "n", "", "specify the node node name, pod.spec.nodeName")

	r.runCmd.Flags().StringVarP(&r.nameSpace, "namespace", "n", "", "specify the namespace, pod.metedata.namespace")
}

func (r *runHelper) start() error {
	//create k8s client
	kc := k8s.NewKClient(r.labelNodeName, r.nameSpace, r.rPhaseOnly)
	return kc.Start()
}
