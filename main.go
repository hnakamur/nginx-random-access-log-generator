package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/hnakamur/ltsvlog"
	"github.com/hnakamur/randutil"
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

func randHost(intner randutil.Intner, siteCount int) (string, error) {
	siteIndex, err := intner.Intn(siteCount)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.example.jp", siteIndex), nil
}

func randBytesSent(intner randutil.Intner, bytesSentMax int) (int, error) {
	// https://www.wolframalpha.com/input/?i=y+%3D+exp(-x)
	intnMax := int(1e6)
	v, err := intner.Intn(intnMax)
	if err != nil {
		return 0, err
	}
	adjuster := float64(10)
	x := float64(v) / float64(intnMax) * adjuster
	y := math.Exp(-x) / math.E
	bytesSent := int(float64(bytesSentMax) * y)
	return bytesSent, nil
}

func main() {
	var bytesSentMax int
	flag.IntVar(&bytesSentMax, "bytes-sent-max", 1e7, "bytes_sent_max")
	var siteCount int
	flag.IntVar(&siteCount, "site-count", 1e4, "site count")
	var logFile string
	flag.StringVar(&logFile, "log-file", "access.log", "log file path")
	var tps int
	flag.IntVar(&tps, "tps", 1000, "access counts per second")
	var duration time.Duration
	flag.DurationVar(&duration, "duration", 10*time.Second, "run duration")
	flag.Parse()

	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("failed to open log file, err=%+v", err)
	}
	defer file.Close()
	ltsvlog.Logger = ltsvlog.NewLTSVLogger(file, false, ltsvlog.SetLevelLabel(""))

	intner := randutil.NewMathIntner(time.Now().UnixNano())
	statusChooser, err := randutil.NewChooser(intner, statusChoices)
	if err != nil {
		log.Fatalf("failed to create chooser for status, err=%+v", err)
	}
	schemeChooser, err := randutil.NewChooser(intner, schemeChoices)
	if err != nil {
		log.Fatalf("failed to create chooser for scheme, err=%+v", err)
	}
	cacheChooser, err := randutil.NewChooser(intner, cacheChoices)
	if err != nil {
		log.Fatalf("failed to create chooser for cache, err=%+v", err)
	}

	var lineCount int
	t := time.Now()
	due := t.Add(duration)
	for t.Before(due) {
		scheme, err := schemeChooser.Choose()
		if err != nil {
			log.Printf("failed to choose scheme, err=%+v", err)
		}
		status, err := statusChooser.Choose()
		if err != nil {
			log.Printf("failed to choose status, err=%+v", err)
		}
		cache, err := cacheChooser.Choose()
		if err != nil {
			log.Printf("failed to choose cache hit status, err=%+v", err)
		}
		host, err := randHost(intner, siteCount)
		if err != nil {
			log.Printf("failed to generate random host, err=%+v", err)
		}
		bytesSent, err := randBytesSent(intner, bytesSentMax)
		if err != nil {
			log.Printf("failed to generate random bytesSent, err=%+v", err)
		}

		ltsvlog.Logger.Info(
			ltsvlog.LV{"host", host},
			ltsvlog.LV{"http_host", host},
			ltsvlog.LV{"scheme", scheme},
			ltsvlog.LV{"status", status.(int)},
			ltsvlog.LV{"bytes_sent", bytesSent},
			ltsvlog.LV{"sent_http_x_cache", cache.(string)},
		)
		lineCount++
		t = time.Now()
	}
	fmt.Printf("lineCount=%d\n", lineCount)
}
