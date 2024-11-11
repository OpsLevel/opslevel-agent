/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"opslevel-agent/signal"
	"opslevel-agent/workers"
	"os"
	"sync"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "opslevel-agent",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := signal.Init(context.Background())
		var wg sync.WaitGroup
		go workers.NewK8SWorker().Run(ctx, &wg)
		time.Sleep(1 * time.Second)
		wg.Wait()
	},
}

func Execute(version, commit, date string) {
	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Msgf("error executing")
		os.Exit(1)
	}
}

func init() {

}
