/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package command

import (
	"errors"

	"github.com/gingray/quitedb/internal/http"
	"github.com/gingray/quitedb/pkg/app"
	"github.com/gingray/quitedb/pkg/config"
	"github.com/gingray/quitedb/pkg/httpserver"
	"github.com/gingray/quitedb/pkg/lifecycle"
	"github.com/gingray/quitedb/pkg/store"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		devMode, _ := cmd.Flags().GetBool("dev")
		if !devMode {
			return
		}
		_ = godotenv.Load()
		return
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg, cfgErr := config.NewConfig()
		app, err := app.NewApp(cfg)
		db := store.NewDb(&cfg.StorageConfig)
		if err != nil {
			err = errors.Join(cfgErr, err)
		}
		if err != nil {
			app.Logger.Error("init app", "error", err)
			return
		}

		server := httpserver.NewServer(&cfg.HTTPServiceConfig, app)
		router := http.NewRouter(db)
		router.SetupRoutes(app.HttpRouter)

		supervisor := lifecycle.NewSupervisor(app.Logger)
		rootNode := supervisor.CreateRootNode()
		dbNode := supervisor.CreateNode(db)
		appNode := supervisor.CreateNode(app)
		serverNode := supervisor.CreateNode(server)

		rootNode.AddNode(dbNode)
		dbNode.AddNode(appNode)
		appNode.AddNode(serverNode)
		err = rootNode.Run(cmd.Context())
		if err != nil {
			app.Logger.Error("run root node", "error", err)
		}
	},
}

func init() {
	serverCmd.Flags().BoolP("dev", "t", false, "Run server in dev mode")
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
