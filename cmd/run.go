package cmd

import (
	"github.com/cjp2600/assr/config"
	"github.com/cjp2600/assr/log"
	"github.com/cjp2600/assr/server"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run server",
	Long:  ``,
	Run:   RunExecute,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func RunExecute(cmd *cobra.Command, args []string) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	src, err := config.GetStaticSrc()
	if err != nil {
		log.Error(err)
		os.Exit(0)
	}

	server.PreloadOnStart()

	go func() {
		stServ := server.NewStatic()
		err := stServ.Run(src)
		if err != nil {
			log.Fatal(err.Error())
		}
	}()
	log.Info(config.GetProjectName() + " is running on " + config.GetAppPort() + " port 🍑 - " + config.GetAppDomain())

	prServ := server.NewParser()
	if err := prServ.Run(src); err != nil {
		log.Fatal(err.Error())
	}
}
