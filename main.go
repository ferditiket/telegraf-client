package main

import (
	"fmt"
	"log"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

type (
	MonitorStatsd interface {
		MonitorLatency(name, group string, latency time.Duration) error
		MonitorCounter(name, group string, status int) error
		MonitorSummary(name, group string, status int, latency time.Duration) error
		CustomMonitorCounter(name, groupName string, tags map[string]interface{}) error
		CustomMonitorLatency(name, groupName string, tags map[string]interface{}, latency time.Duration) error
		CustomMonitorSummary(name, groupName string, latency time.Duration, tags map[string]interface{}) error
	}

	monitorStatsd struct {
		Host         string
		Port         string
		StatsDClient *statsd.Client
	}
)

func NewMonitor(host, port string) (MonitorStatsd, error) {
	statsd, err := statsd.New(fmt.Sprintf("%s:%s", host, port))
	s := monitorStatsd{
		Host:         host,
		Port:         port,
		StatsDClient: statsd,
	}
	return &s, err
}

func (monitor *monitorStatsd) MonitorLatency(name, groupName string, latency time.Duration) error {
	tags := []string{groupName}

	monitor.StatsDClient.Timing(
		name,
		latency,
		tags,
		1.0,
	)
	return nil
}

func (monitor *monitorStatsd) MonitorCounter(name, groupName string, status int) error {
	tags := []string{groupName}
	monitor.StatsDClient.Count(
		name,
		1,
		tags,
		1,
	)

	return nil
}

func (monitor *monitorStatsd) MonitorSummary(name, groupName string, status int, latency time.Duration) error {
	err := monitor.MonitorLatency(name, groupName, latency)

	if err != nil {
		return err
	}

	err = monitor.MonitorCounter(name, groupName, status)

	return err
}

func (monitor *monitorStatsd) CustomMonitorCounter(name, groupName string, tags map[string]interface{}) error {
	statsdTags := []string{groupName}
	statsdTags = append(statsdTags, monitor.buildTagsString(tags)...)

	err := monitor.StatsDClient.Count(
		name,
		1,
		statsdTags,
		1.0,
	)

	return err
}

func (monitor *monitorStatsd) CustomMonitorLatency(name, groupName string, tags map[string]interface{}, latency time.Duration) error {
	statsdTags := []string{groupName}
	statsdTags = append(statsdTags, monitor.buildTagsString(tags)...)

	err := monitor.StatsDClient.Timing(
		name,
		latency,
		statsdTags,
		1.0,
	)
	return err
}

func (monitor *monitorStatsd) CustomMonitorSummary(name, groupName string, latency time.Duration, tags map[string]interface{}) error {
	err := monitor.CustomMonitorLatency(name, groupName, tags, latency)
	if err != nil {
		return err
	}

	err = monitor.CustomMonitorCounter(name, groupName, tags)
	return err
}

func (monitor *monitorStatsd) buildTagsString(tags map[string]interface{}) []string {
	var tagsString []string
	for key, val := range tags {
		tagsString = append(tagsString, fmt.Sprintf("%s:%s", key, val))
	}
	return tagsString
}

func main() {

	start := time.Now()

	statsdApp, err := NewMonitor("localhost", "8125")
	if err != nil {
		log.Fatalf("error %s", err.Error())
	}

	tags := map[string]interface{}{
		"status": "hit",
	}

	statsdApp.CustomMonitorCounter("app.hotel-price-engine.cache-hit-ratio.hotelid123", "REDIS", tags)
	// statsdApp.CustomMonitorCounter("app.hotel-price-engine.cache-miss-ratio.hotelid123", "REDIS", tags)
	statsdApp.MonitorLatency("app.hotel-price-engine.vendor-response.expedia-rapid", "API-OUT", time.Since(start))
	statsdApp.MonitorCounter("app.hotel-price-engine.cache-hit-ratio.hotel456", "REDIS", 200)
	statsdApp.MonitorSummary("app.hotel-price-engine.cache-hit-ratio.hotel789", "REDIS", 200, time.Since(start))
	statsdApp.CustomMonitorLatency("app.hotel-price-engine.vendor-get-avail.rakuten", "API-OUT", tags, time.Since(start))
	statsdApp.CustomMonitorLatency("app.hotel-price-engine.vendor-get-checkrate", "API-OUT", tags, time.Since(start))
	time.Sleep(5 * time.Second)

}
