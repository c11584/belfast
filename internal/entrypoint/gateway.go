package entrypoint

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/akamensky/argparse"
	"github.com/fsnotify/fsnotify"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/packets"
)

var gatewayOnce sync.Once

const gatewayConfigReloadDelay = 200 * time.Millisecond

// RunGateway 启动 gateway 服务端，返回错误信息供调用方处理
func RunGateway() error {
	initGatewayRuntime()
	parser := argparse.NewParser("gateway", "Azur Lane gateway")
	configPath := parser.String("", "config", &argparse.Options{
		Required: false,
		Help:     "Path to TOML config file",
		Default:  "gateway.toml",
	})
	if err := parser.Parse(os.Args); err != nil {
		return fmt.Errorf("argument parse error: %w", err)
	}
	loadedConfig, err := config.LoadGateway(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load gateway config '%s': %w", *configPath, err)
	}
	if loadedConfig.Mode == "proxy" {
		runtime := newGatewayProxyRuntime(loadedConfig)
		go watchGatewayConfig(*configPath, loadedConfig, func(updated config.GatewayConfig) {
			if updated.Mode != "proxy" {
				return
			}
			runtime.Update(updated)
		})
		if err := runtime.Run(); err != nil {
			return fmt.Errorf("gateway proxy error: %w", err)
		}
		return nil
	}
	server := connection.NewServer(loadedConfig.BindAddress, loadedConfig.Port, packets.Dispatch)
	if loadedConfig.RequirePrivateClients != nil {
		server.SetRequirePrivateClients(*loadedConfig.RequirePrivateClients)
	}
	go watchGatewayConfig(*configPath, loadedConfig, func(updated config.GatewayConfig) {
		if updated.Mode != "serve" {
			return
		}
		if updated.RequirePrivateClients == nil {
			return
		}
		server.SetRequirePrivateClients(*updated.RequirePrivateClients)
	})
	if err := server.Run(); err != nil {
		return fmt.Errorf("gateway run error: %w", err)
	}
	return nil
}

func initGatewayRuntime() {
	gatewayOnce.Do(func() {
		registerGatewayPackets()
	})
}

func watchGatewayConfig(path string, currentConfig config.GatewayConfig, onReload func(config.GatewayConfig)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.LogEvent("Gateway", "Config", fmt.Sprintf("failed to init watcher: %s", err.Error()), logger.LOG_LEVEL_WARN)
		return
	}
	defer watcher.Close()

	configDir := filepath.Dir(path)
	configBase := filepath.Base(path)
	if err := watcher.Add(configDir); err != nil {
		logger.LogEvent("Gateway", "Config", fmt.Sprintf("failed to watch config dir: %s", err.Error()), logger.LOG_LEVEL_WARN)
		return
	}

	var reloadTimer *time.Timer
	reloadConfig := func() {
		updatedConfig, err := config.LoadGateway(path)
		if err != nil {
			logger.LogEvent("Gateway", "Config", fmt.Sprintf("failed to reload config: %s", err.Error()), logger.LOG_LEVEL_WARN)
			return
		}
		logger.LogEvent("Gateway", "Config", "config reloaded", logger.LOG_LEVEL_INFO)
		if updatedConfig.BindAddress != currentConfig.BindAddress || updatedConfig.Port != currentConfig.Port {
			logger.LogEvent("Gateway", "Config", "bind_address/port changes require restart", logger.LOG_LEVEL_WARN)
		}
		if updatedConfig.Mode != currentConfig.Mode {
			logger.LogEvent("Gateway", "Config", "mode changes require restart", logger.LOG_LEVEL_WARN)
		}
		if onReload != nil {
			onReload(updatedConfig)
		}
		currentConfig = updatedConfig
	}
	scheduleReload := func() {
		if reloadTimer == nil {
			reloadTimer = time.AfterFunc(gatewayConfigReloadDelay, reloadConfig)
			return
		}
		reloadTimer.Reset(gatewayConfigReloadDelay)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(event.Name) != configBase {
				continue
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
				scheduleReload()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.LogEvent("Gateway", "Config", fmt.Sprintf("watcher error: %s", err.Error()), logger.LOG_LEVEL_WARN)
		}
	}
}
