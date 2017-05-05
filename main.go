package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/hnakamur/randutil"
	ltsv "github.com/hnakamur/zap-ltsv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var statusChoices = []randutil.Choice{
	{Weight: 70, Item: 200},
	{Weight: 15, Item: 301},
	{Weight: 5, Item: 400},
	{Weight: 10, Item: 404},
	{Weight: 5, Item: 503},
}

var schemeChoices = []randutil.Choice{
	{Weight: 60, Item: "https"},
	{Weight: 40, Item: "http"},
}

var cacheChoices = []randutil.Choice{
	{Weight: 60, Item: "HIT"},
	{Weight: 20, Item: "MISS"},
	{Weight: 20, Item: "-"},
}

func ISO8601NoNanoTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05Z0700"))
}

func normRand(min, max int, mean, stdDev float64) int {
	r := int(rand.NormFloat64()*stdDev + mean)
	if r < min {
		return min
	} else if max < r {
		return max
	}
	return r
}

func randHost(intner randutil.Intner, siteCount int) (string, error) {
	siteIndex, err := intner.Intn(siteCount)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.example.jp", siteIndex), nil
}

func main() {
	var bytesSentMean float64
	flag.Float64Var(&bytesSentMean, "bytes-sent-mean", 1e5, "bytes_sent_mean")
	var bytesSentStdDev float64
	flag.Float64Var(&bytesSentStdDev, "bytes-sent-std-dev", 1e4, "bytes_sent_std_dev")
	var bytesSentMax int
	flag.IntVar(&bytesSentMax, "bytes-sent-max", 1e7, "bytes_sent_max")
	var siteCount int
	flag.IntVar(&siteCount, "site-count", 1e4, "site count")
	flag.Parse()

	err := ltsv.RegisterLTSVEncoder()
	if err != nil {
		panic(err)
	}

	cfg := ltsv.NewProductionConfig()
	cfg.EncoderConfig.MessageKey = ""
	cfg.EncoderConfig.LevelKey = ""
	cfg.EncoderConfig.CallerKey = ""
	cfg.EncoderConfig.EncodeTime = ISO8601NoNanoTimeEncoder
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	intner := randutil.NewCryptoIntner()
	statusChooser, err := randutil.NewChooser(intner, statusChoices)
	if err != nil {
		logger.Fatal("", zap.Error(err))
	}
	schemeChooser, err := randutil.NewChooser(intner, schemeChoices)
	if err != nil {
		logger.Fatal("", zap.Error(err))
	}
	cacheChooser, err := randutil.NewChooser(intner, cacheChoices)
	if err != nil {
		logger.Fatal("", zap.Error(err))
	}

	for i := 0; i < 100; i++ {
		scheme, err := schemeChooser.Choose()
		if err != nil {
			logger.Error("", zap.Error(err))
		}
		status, err := statusChooser.Choose()
		if err != nil {
			logger.Error("", zap.Error(err))
		}
		cache, err := cacheChooser.Choose()
		if err != nil {
			logger.Error("", zap.Error(err))
		}
		host, err := randHost(intner, siteCount)
		if err != nil {
			logger.Error("", zap.Error(err))
		}
		bytesSent := normRand(0, bytesSentMax, bytesSentMean, bytesSentStdDev)
		logger.Info("",
			zap.String("host", host),
			zap.String("http_host", host),
			zap.String("scheme", scheme.(string)),
			zap.Int("status", status.(int)),
			zap.Int("bytes_sent", bytesSent),
			zap.String("sent_http_x_cache", cache.(string)),
		)
	}
}
