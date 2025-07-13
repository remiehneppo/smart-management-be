/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/remiehneppo/be-task-management/config"
	"github.com/remiehneppo/be-task-management/internal/service"
	"github.com/remiehneppo/be-task-management/utils"
	"github.com/sashabaranov/go-openai"
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
		cfgYml, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadConfig(cfgYml)
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}
		vlmService := service.NewVLMService(cfg.VLM)
		_ = vlmService
		imagePath := "test_data/images/image.png"
		image, err := os.ReadFile(imagePath)
		if err != nil {
			fmt.Println("Error reading image file:", err)
			return
		}
		base64Encoding := utils.ConvertToBase64URL(image)

		messages := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "Trích xuất văn bản trong tài liệu sau dưới dạng markdown. Với các ảnh có trong văn bản, hãy đặt mô tả nội dung ảnh trong alt text markdown. Chỉ trích xuất văn bản, không nhận xét gì thêm",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: base64Encoding,
						},
					},
				},
			},
		}

		res, err := vlmService.ChatMultiContent(cmd.Context(), messages)
		if err != nil {
			fmt.Println("Error during chat with VLM:", err)
			return
		}
		fmt.Println("Response from VLM:", res)

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
	cheatCmd.Flags().StringP("config", "c", "config.yaml", "Path to the configuration file")
}
