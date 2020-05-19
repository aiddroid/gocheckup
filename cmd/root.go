/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"gocheckup"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var cfgVersion bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-checker",
	Short: "A fast multiple protocol checker written by Golang.",
	Long: `A fast multiple protocol checker written by Golang. For example, you can make following checks:

HTTP、DNS、FTP、SSH、TCP、UDP and etc.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// 加载checkup
		c := loadCheckup()
		log.Println("checkup:", c)

		// 执行检查
		results, errors := c.Check()
		log.Println("results:", results)
		log.Println("errors:", errors)

		// 存储检查结果
		if c.Storage != nil {
			c.Storage.Store(results)
		}
	},
}

// 加载checkup
func loadCheckup() checkup.Checkup {
	c := &checkup.Checkup{}

	// 读取配置文件
	b, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	// 解析出checkers和storage 2个字段备用
	raw := struct {
		Checkers []json.RawMessage `json:"checkers"`
		Storage  json.RawMessage   `json:"storage"`
	}{}
	json.Unmarshal(b, &raw)
	log.Println("raw.Checkers：", raw.Checkers)
	log.Println("raw.Storage：", string(raw.Storage[:]))

	if raw.Checkers == nil {
		return *c
	}

	// 解析出checkers对象和Storage对象
	tmp := struct {
		Checkers []struct {
			Type string `json:"type"`
		} `json:"checkers"`
		Storage struct {
			Type string `json:"type"`
		} `json:"storage"`
	}{}
	json.Unmarshal(b, &tmp)

	// 根据storage.type的值，进一步解析出具体的storage对象
	if raw.Storage != nil {
		switch tmp.Storage.Type {
		case "fs":
			var storage checkup.Fs
			json.Unmarshal(raw.Storage, &storage)
			c.Storage = storage
		}
	}

	// 根据checker.type的值，进一步解析出具体的checker对象，并赋值给checkup.checkers
	for i, cc := range tmp.Checkers {
		switch cc.Type {
		case "http":
			var checker checkup.HttpChecker
			err := json.Unmarshal(raw.Checkers[i], &checker)
			if err != nil {
				log.Fatalln(err)
			}
			checkers := append(c.Checkers, checker)
			c.Checkers = checkers

			log.Println("checker"+strconv.Itoa(i)+":", checker, "\n")
		}
	}

	return *c
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// 绑定config参数
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "checkup.json", "config file (default is $HOME/.go-checker.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// 绑定version参数
	rootCmd.PersistentFlags().BoolVarP(&cfgVersion, "version", "v", false, "Show version info")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgVersion {
		fmt.Println("GoCheckup: v1.0")
		os.Exit(0)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".go-checker" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".go-checker")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
