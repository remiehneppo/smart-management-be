/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/remiehneppo/be-task-management/internal/service"
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

		// write to a text file
		outputFile, err := os.Create("out.json")
		if err != nil {
			fmt.Println("Error creating output file:", err)
			return
		}
		defer outputFile.Close()
		filePath := "./test_data/cong_nghe_dong_va_sua_chua_tau_thuy.pdf"
		totalPages, err := pdfService.GetTotalPages(filePath)
		if err != nil {
			panic(err)
		}
		fmt.Println("total pages", totalPages)
		chunks, err := pdfService.ProcessPDF(filePath)
		if err != nil {
			panic(err)
		}
		outputFile.WriteString("[\n")
		for _, chunk := range chunks {
			jsonData, err := json.MarshalIndent(chunk, "", "  ")
			if err != nil {
				fmt.Println("Error marshalling chunk to JSON:", err)
				return
			}
			outputFile.Write(jsonData)
			outputFile.WriteString(",\n")
		}
		outputFile.WriteString("]\n")

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
