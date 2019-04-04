package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/urfave/cli"

	yaml "gopkg.in/yaml.v2"
)

// Config :YAMLの全体
type Config struct {
	Target TargetConfig `yaml:"target"`
	Trash  TrashConfig  `yaml:"trash"`
}

// TargetConfig :削除対象についての情報
type TargetConfig struct {
	Folders []string `yaml:"folders"`
}

// TrashConfig :ゴミ箱情報
type TrashConfig struct {
	Directory string `yaml:"directory"`
}

func main() {
	app := cli.NewApp()
	app.Name = "TMP File Cleaner"
	app.Usage = ""
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Value: "config.yaml",
			Usage: "Config file(YAML)",
		},
		cli.BoolFlag{
			Name:  "Yes, Y",
			Usage: "Unshow confirm message",
		},
		cli.BoolFlag{
			Name:  "trash, t",
			Usage: "move files into trash can",
		},
	}

	app.Action = func(c *cli.Context) error {

		configfile := c.String("file")
		skipconfirm := c.Bool("Yes")
		trashmode := c.Bool("trash")

		t := time.Now()

		data, err := ioutil.ReadFile(configfile)
		if err != nil {
			log.Fatalf("cannot read config file: %v", err)
		}
		var config Config
		yaml.Unmarshal(data, &config)
		if err != nil {
			log.Fatalf("cannot unmarshal data: %v", err)
		}

		// ゴミ箱モードの時は退避用ディレクトリ(名前は日付)を作成する
		trashdir := filepath.Join(config.Trash.Directory, strconv.Itoa(t.Day()))
		if trashmode {
			if err := os.MkdirAll(trashdir, 0777); err != nil {
				fmt.Println(err)
			}
		}

		targetFolders := config.Target

		// 対象ディレクトリ内のファイルを削除する
		for _, folder := range targetFolders.Folders {
			fmt.Printf("ディレクトリ[" + folder + "]内のファイルをすべて削除します。")
			if skipconfirm {
				fmt.Println("")
			} else {
				fmt.Println("よろしいですか？[Y/n]")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				if scanner.Text() != "Y" {
					continue
				}
			}

			// ゴミ箱モードの時はファイルを退避する
			if trashmode {
				CopyAll(folder, trashdir)
			}

			os.Chdir(folder)

			files, err := ioutil.ReadDir(folder)
			if err != nil {
				log.Fatal(err)
				fmt.Println("warning: cannot read directory: " + folder)
			} else {
				for _, file := range files {
					fmt.Printf("削除中：" + file.Name() + "\r")
					os.RemoveAll(file.Name())
					fmt.Printf("                                           \r")
				}
				fmt.Println("")
			}
		}
		return nil
	}

	app.Run(os.Args)
}

// CopyAll ファイル、ディレクトリ(サブディレクトリ含む)をコピーする
func CopyAll(src string, dst string) {
	fInfo, _ := os.Stat(src)
	if fInfo.IsDir() {
		fmt.Println("directory copy:" + fInfo.Name())
		// ディレクトリの場合はコピー先にディレクトリを作成してから中身をコピーする
		newdir := filepath.Join(dst,fInfo.Name())
		os.Mkdir(newdir, 0777)
		files, err := ioutil.ReadDir(src)
		if err != nil {
			log.Fatal(err)
			fmt.Println("warning: cannot read directory: " + src)
		} else {
			for _, file := range files {
				CopyAll(filepath.Join(src,file.Name()), newdir)
			}
		}
	} else {
		// ファイルの場合
		fmt.Println("file copy:" + fInfo.Name())
		Copy(src , filepath.Join(dst ,fInfo.Name()))
	}
	return
}

// Copy ファイルをコピーする
func Copy(srcfile string, dstfile string){

    src, err := os.Open(srcfile)
    if err != nil {
        panic(err)
    }
    defer src.Close()

    dst, err := os.Create(dstfile)
    if err != nil {
        panic(err)
    }
    defer dst.Close()

    _, err = io.Copy(dst, src)
    if  err != nil {
        panic(err)
    }
}
