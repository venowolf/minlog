package cmd

import (
	"venowolf/minlog/app/minlog/k8s"

	"github.com/spf13/cobra"
)

var profileCmd *cobra.Command = &cobra.Command{

	Use:   "profile",
	Short: "create river file only, not send request to reloading grafana-alloy",
	Long:  `running in initcontainer "minlog profile"`,
	Run: func(cmd *cobra.Command, args []string) {
		//l := log.GetLogger()
		//l.Sugar().Infof("%v", os.Args)
		//l.Sugar().Info("minlog starting...")
		// checking grafana-alloy health
		//create k8s client
		kc := k8s.NewKClient(labelNodeName, nameSpace, rPhaseOnly)
		//at first, get all pods in the node
		kc.Profilling(false)
	},
}

func init() {
	parseflags(profileCmd)
}

func GetProfileCommand() *cobra.Command {
	return profileCmd
}
