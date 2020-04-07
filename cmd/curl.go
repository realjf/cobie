package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
)

func init() {
	rootCmd.AddCommand(curlCmd)
}

var curlCmd = &cobra.Command{
	Use:   "curl",
	Short: "curl [I] http://host:port/xxx",
	Long:  `like curl [I] http://host:port/xxx， ‘I’ means don't need to return html document'`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("usage: curl [I] http://host:port/xxx")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var returnHtml bool = true
		if len(args) == 2 {
			switch args[0] {
			case "I":
				get(args[1], false)
			}
		} else {
			get(args[0], returnHtml)
		}
	},
}

func get(url string, returnHtml bool) {
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
	if returnHtml {
		fmt.Println()
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			fmt.Println("读取页面数据失败")
			os.Exit(-1)
		}
		fmt.Println("[页面数据]： ", string(body))
	}
}
