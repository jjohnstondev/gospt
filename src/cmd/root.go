package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cmds "gitea.asdf.cafe/abs3nt/gospt/src/commands"

	"gitea.asdf.cafe/abs3nt/gospt/src/config"
	"gitea.asdf.cafe/abs3nt/gospt/src/gctx"
	"tuxpa.in/a/zlog"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	ctx         *gctx.Context
	commands    *cmds.Commands
	cfgFile     string
	userLicense string
	verbose     bool

	rootCmd = &cobra.Command{
		Use:   "gospt",
		Short: "A spotify TUI and CLI to manage playback, browse library, and generate radios",
		Long:  `A spotify TUI and CLI to manage playback, borwse library, and generate radios written in go`,
	}
)

// Execute executes the root command.
func Execute(defCmd string) {
	if len(os.Args) == 1 {
		args := append([]string{defCmd}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	zlog.SetGlobalLevel(zlog.DebugLevel)
	if len(os.Args) > 1 {
		if os.Args[1] == "completion" || os.Args[1] == "__complete" {
			return
		}
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	cobra.OnInitialize(func() {
		if verbose {
			zlog.SetGlobalLevel(zlog.TraceLevel)
		}
	})
	ctx = gctx.NewContext(context.Background())
	commands = &cmds.Commands{Context: ctx}
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	configDir, _ := os.UserConfigDir()
	cfgFile = filepath.Join(configDir, "gospt/client.yml")
	yamlDecoder := aconfigyaml.New()

	loader := aconfig.LoaderFor(&config.Values, aconfig.Config{
		AllowUnknownFields: true,
		AllowUnknownEnvs:   true,
		AllowUnknownFlags:  true,
		SkipFlags:          true,
		DontGenerateTags:   true,
		MergeFiles:         true,
		EnvPrefix:          "",
		FlagPrefix:         "",
		Files: []string{
			cfgFile,
		},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yml": yamlDecoder,
		},
	})
	if err := loader.Load(); err != nil {
		panic(err)
	}
	if config.Values.ClientSecretCmd != "" {
		args := strings.Fields(config.Values.ClientSecretCmd)
		secret, err := exec.Command(args[0], args[1:]...).Output()
		if err != nil {
			panic(err)
		}
		config.Values.ClientSecret = strings.TrimSpace(string(secret))
	}
}
