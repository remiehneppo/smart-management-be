/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/remiehneppo/be-task-management/internal/service"
	"github.com/remiehneppo/be-task-management/types"
	"github.com/spf13/cobra"
)

// cheatCmd represents the cheat command
var cheatCmd = &cobra.Command{
	Use:   "cheat",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cheat called")
		pdfService := service.NewPDFService(service.DefaultDocumentServiceConfig)

		pages, err := pdfService.ExtractPageContent(&types.ExtractPageContentRequest{
			ToolUse:  "pdftotext",
			FilePath: "./test_data/ShipDesign.pdf",
			FromPage: 1,
			ToPage:   10,
		})

		if err != nil {
			fmt.Println("Error extracting pages:", err)
			return
		}
		for _, page := range pages {
			fmt.Println("Page content:", page)
		}
		// Example of using the PDF service to extract text from a PDF file

	},
}

func init() {
	rootCmd.AddCommand(cheatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cheatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cheatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
