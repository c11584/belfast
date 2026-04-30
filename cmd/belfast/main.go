package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/ggmolly/belfast/internal/entrypoint"
)

func main() {
	err := entrypoint.Run(entrypoint.Options{
		CommandName:   "belfast",
		Description:   "Azur Lane server emulator",
		DefaultConfig: "server.toml",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n启动失败: %v\n", err)
		if runtime.GOOS == "windows" {
			fmt.Println("\n按回车键退出...")
			fmt.Scanln()
		}
		os.Exit(1)
	}
}
