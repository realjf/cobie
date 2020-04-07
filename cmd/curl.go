package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

func init() {
	rootCmd.AddCommand(curlCmd)
}

var curlCmd = &cobra.Command{
	Use:   "curl",
	Short: "curl http://host:port/xxx",
	Long:  `like curl -I http://host:port/xxx`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("usage: curl http://host:port/xxx")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		get(args[0])
	},
}

func get(url string) {
	fmt.Println("get " + url + "...")
	client := http.Client{}
	rsp, err := client.Get(url)
	if err != nil {
		fmt.Println("获取资源失败", err)
		os.Exit(-1)
	}
	defer rsp.Body.Close()

	// 确认服务器信息
	rspHeader := rsp.Header
	for k, v := range rspHeader {
		var val string
		for _, vv := range v {
			val += vv
		}
		fmt.Println(k + ": " + val)
	}
}
