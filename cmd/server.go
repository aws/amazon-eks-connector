package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"

	"github.com/aws/amazon-eks-connector/pkg/config"
	"github.com/aws/amazon-eks-connector/pkg/fsnotify"
	"github.com/aws/amazon-eks-connector/pkg/proxy"
	"github.com/aws/amazon-eks-connector/pkg/server"
	"github.com/aws/amazon-eks-connector/pkg/serviceaccount"
	"github.com/aws/amazon-eks-connector/pkg/state"
)

var serverCmdViperFlag = viper.New()
var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Run EKS connector proxy server",
	Example: "",
	Run: func(cmd *cobra.Command, args []string) {
		configProvider := config.NewProvider(serverCmdViperFlag)
		configuration, err := configProvider.Get()
		if err != nil {
			klog.Fatalf("failed to load configuration: %v", err)
		}

		secretProvider := serviceaccount.NewProvider()

		server := server.Server{
			ProxyConfig:  configuration.ProxyConfig,
			ProxyHandler: proxy.NewProxyHandler(configuration.ProxyConfig, secretProvider),
		}

		if err = fsnotify.NewWatcher(configuration.StateConfig); err != nil {
			klog.Fatalf("failed to setup file watcher: %v", err)
		}

		server.Run()
	},
}

func init() {
	serverCmd.Flags().String("proxy.socketType",
		"unix",
		"The socket type of proxy. Can be 'unix' or 'tcp'")
	serverCmd.Flags().String("proxy.socketAddr",
		"/var/eks/shared/connector.sock",
		"The address of proxy, should be a FS path or network address depending on socket type")
	serverCmd.Flags().String("proxy.targetHost",
		"kubernetes.default.svc:443",
		"The target of the proxy, should be api server's address")
	serverCmd.Flags().String("proxy.targetProtocol",
		"https",
		"The target protocol of the proxy. Can be 'https' or 'http'")
	serverCmd.Flags().String("state.baseDir",
		state.DirSsmVault,
		"The vault folder of ssm agent container")
	serverCmd.Flags().String("state.secretNamePrefix",
		"eks-connector-state",
		"Prefix of Kubernetes Secret name used to persist eks-connector state")
	serverCmd.Flags().String("state.secretNamespace",
		"eks-connector",
		"Kubernetes namespace of the Secret used to persist eks-connector state")
	err := serverCmdViperFlag.BindPFlags(serverCmd.Flags())
	if err != nil {
		klog.Fatal("failed to bind cmd flags: %v", err)
	}

	rootCmd.AddCommand(serverCmd)
}
