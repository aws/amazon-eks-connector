package main

import (
	"flag"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

var rootCmd = &cobra.Command{
	Use:   "eks-connector",
	Short: "connects any k8s cluster to AWS cloud",
}

func main() {
	addKlogFlags(rootCmd.PersistentFlags())

	if err := rootCmd.Execute(); err != nil {
		klog.Exitln(err)
	}
}

// addKlogFlags adds flags from k8s.io/klog/v2
// marks the flags as hidden to avoid polluting the help text
func addKlogFlags(fs *pflag.FlagSet) {
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		newflag := pflag.PFlagFromGoFlag(fl)
		newflag.Hidden = true
		fs.AddFlag(newflag)
	})
}
