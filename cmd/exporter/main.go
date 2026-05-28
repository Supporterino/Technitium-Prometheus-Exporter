package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"technitium-dns-exporter/internal/collector"
	"technitium-dns-exporter/internal/config"
)

var (
	configPath string
	version    = "dev"
	commit     = "unknown"
	date       = "unknown"
)

func main() {
	flag.StringVar(&configPath, "config", "", "Path to config file (default: ./config/config.yaml or $EXPORTER_CONFIG)")
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("EXPORTER_CONFIG")
	}
	if configPath == "" {
		configPath = "./config/config.yaml"
	}

	level := slog.LevelInfo
	switch os.Getenv("EXPORTER_LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	setLogLevel(logger, cfg.Exporter.LogLevel)

	mu := &sync.Mutex{}
	var handler http.Handler = promhttp.HandlerFor(prometheus.NewRegistry(), promhttp.HandlerOpts{})

	registerCollectors := func() {
		mu.Lock()
		defer mu.Unlock()

		var registries []prometheus.Gatherer

		for _, target := range cfg.Targets {
			reg := prometheus.NewRegistry()
			c := collector.New(target, cfg.Exporter.ScrapeTimeout, logger)
			reg.MustRegister(c)
			registries = append(registries, reg)
		}

		handler = promhttp.HandlerFor(prometheus.Gatherers(registries), promhttp.HandlerOpts{})
	}

	registerCollectors()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)

	go func() {
		for range sigCh {
			logger.Info("received SIGHUP, reloading config")
			newCfg, err := config.Load(configPath)
			if err != nil {
				logger.Warn("failed to reload config, keeping old config", "error", err)
				continue
			}
			setLogLevel(logger, newCfg.Exporter.LogLevel)
			cfg = newCfg
			registerCollectors()
			logger.Info("config reloaded successfully")
		}
	}()

	http.HandleFunc(cfg.Exporter.MetricsPath, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		h := handler
		mu.Unlock()
		h.ServeHTTP(w, r)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`<html>
<head><title>Technitium DNS Exporter</title></head>
<body>
<h1>Technitium DNS Exporter</h1>
<p><a href="` + cfg.Exporter.MetricsPath + `">Metrics</a></p>
</body>
</html>`))
	})

	logger.Info("starting exporter", "listen", cfg.Exporter.ListenAddress)
	if err := http.ListenAndServe(cfg.Exporter.ListenAddress, nil); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func setLogLevel(logger *slog.Logger, level string) {
	switch level {
	case "debug":
		logger.Debug("log level set", "level", level)
	case "warn":
		logger.Warn("log level set", "level", level)
	case "error":
		logger.Error("log level set", "level", level)
	default:
		logger.Info("log level set", "level", level)
	}
}
