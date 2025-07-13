/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/remiehneppo/be-task-management/config"
	"github.com/spf13/cobra"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
)

// initDbCmd represents the initDb command
var initDbCmd = &cobra.Command{
	Use:   "init-db",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init-db called")
		cfgYml, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadConfig(cfgYml)
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}
		jsonCfg, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Println("Error marshalling config to JSON:", err)
			return
		}
		fmt.Println("Loaded config:", string(jsonCfg))

		vectorDbCfg := weaviate.Config{
			Host:   cfg.Weaviate.Host,
			Scheme: cfg.Weaviate.Scheme,
		}
		if cfg.Weaviate.APIKey != "" {
			vectorDbCfg.AuthConfig = auth.ApiKey{
				Value: cfg.Weaviate.APIKey,
			}
			vectorDbCfg.Headers = map[string]string{
				"X-Weaviate-Api-Key":     cfg.Weaviate.APIKey,
				"X-Weaviate-Cluster-Url": fmt.Sprintf("%s://%s", cfg.Weaviate.Scheme, cfg.Weaviate.Host),
			}
		}
		for _, header := range cfg.Weaviate.Header {
			vectorDbCfg.Headers[header.Key] = header.Value
		}
		weaviateClient, err := weaviate.NewClient(vectorDbCfg)
		if err != nil {
			fmt.Println("Error creating Weaviate client:", err)
			return
		}
		fmt.Println("Weaviate client created successfully:", weaviateClient)
		// delete all classes
		schema, err := weaviateClient.Schema().Getter().Do(cmd.Context())
		if err != nil {
			fmt.Println("Error getting Weaviate classes:", err)
			return
		}
		for _, class := range schema.Classes {
			fmt.Printf("Try to delete class \"%s\" from vector db\n", class.Class)
			// request confirm
			fmt.Printf("Are you sure you want to delete class '%s'? (y/N): ", class.Class)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				continue
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Printf("Skipping deletion of class: %s\n", class.Class)
				continue
			}

			err = weaviateClient.Schema().ClassDeleter().WithClassName(class.Class).Do(cmd.Context())
			if err != nil {
				fmt.Printf("Error deleting class %s: %v\n", class.Class, err)
			} else {
				fmt.Printf("Deleted class: %s\n", class.Class)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(initDbCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initDbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initDbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	initDbCmd.Flags().StringP("config", "c", "config.yaml", "Path to the configuration file")
}
