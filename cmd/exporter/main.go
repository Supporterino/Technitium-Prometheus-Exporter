package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

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

	leveler := &atomicLevel{}
	switch os.Getenv("EXPORTER_LOG_LEVEL") {
	case "debug":
		leveler.Set(slog.LevelDebug)
	case "warn":
		leveler.Set(slog.LevelWarn)
	case "error":
		leveler.Set(slog.LevelError)
	default:
		leveler.Set(slog.LevelInfo)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: leveler}))

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	setLogLevel(leveler, cfg.Exporter.LogLevel)

	mu := &sync.Mutex{}
	var targets []*collector.TechnitiumCollector
	var handler http.Handler = promhttp.HandlerFor(prometheus.NewRegistry(), promhttp.HandlerOpts{})

	registerCollectors := func() {
		mu.Lock()
		defer mu.Unlock()

		var registries []prometheus.Gatherer
		var newTargets []*collector.TechnitiumCollector

		for _, target := range cfg.Targets {
			reg := prometheus.NewRegistry()
			c := collector.New(target, target.RequestTimeout, cfg.Exporter.ScrapeTimeout, logger)
			reg.MustRegister(c)
			registries = append(registries, reg)
			newTargets = append(newTargets, c)
		}

		targets = newTargets
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
			setLogLevel(leveler, newCfg.Exporter.LogLevel)
			cfg = newCfg
			registerCollectors()
			logger.Info("config reloaded successfully")
		}
	}()

	http.HandleFunc(cfg.Exporter.MetricsPath, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		h := handler
		ts := make([]*collector.TechnitiumCollector, len(targets))
		copy(ts, targets)
		mu.Unlock()

		if scrapeTimeout := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); scrapeTimeout != "" {
			if secs, err := strconv.ParseFloat(scrapeTimeout, 64); err == nil && secs > 0 {
				timeout := time.Duration(secs * float64(time.Second))
				for _, t := range ts {
					t.SetScrapeTimeout(timeout)
				}
			}
		}

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

func setLogLevel(leveler *atomicLevel, level string) {
	switch level {
	case "debug":
		leveler.Set(slog.LevelDebug)
	case "warn":
		leveler.Set(slog.LevelWarn)
	case "error":
		leveler.Set(slog.LevelError)
	default:
		leveler.Set(slog.LevelInfo)
	}
}

type atomicLevel struct {
	level atomic.Int32
}

func (l *atomicLevel) Level() slog.Level {
	return slog.Level(l.level.Load())
}

func (l *atomicLevel) Set(level slog.Level) {
	l.level.Store(int32(level))
}
