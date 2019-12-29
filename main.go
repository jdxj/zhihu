package main

import (
	"os"
	"os/signal"
	"syscall"
	"zhihu/modules"

	"github.com/astaxie/beego/logs"
)

func main() {
	err := logs.SetLogger(logs.AdapterFile,
		`{"filename":"zhihu.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10,"color":true}`)
	if err != nil {
		panic(err)
	}

	config, err := modules.ReadConfig("config.json")
	if err != nil {
		logs.Error("%s", err)
		return
	}

	zhiHu, err := modules.NewZhiHu(config.ZhiHu, config.MySQL)
	if err != nil {
		logs.Error("%s", err)
		return
	}
	defer zhiHu.Stop()

	go zhiHu.CollectURLToken()

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigs:
		logs.Info("receive int signal")
	case <-zhiHu.Wait():
		logs.Info("finish in normal")
	}
}
