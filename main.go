package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mcuadros/go-defaults"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	lastPush = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "last_push_timestamp_seconds",
			Help: "Unix timestamp of the last received metrics push in seconds.",
		},
	)
	configuration = &Configuration{}
	config        = ExporterConfiguration{}
	collector     = &mqttCollector{}
)

type ExporterConfig struct {
	ListeningAddress  string `mapstructure:"listeningAddress" default:":9393"`
	MetricsPath       string `mapstructure:"metricsPath" default:"/metrics"`
	ConfigurationFile string `mapstructure:"configurationFile"`
	Verbose           bool   `mapstructure:"verbose" default:"false"`
}

type ExporterMqttConfig struct {
	Topic  string `mapstructure:"topic" default:"/mqttexporter/#"`
	Broker string `mapstructure:"broker" default:"tcp://127.0.0.1:1883"`
}

type ExporterConfiguration struct {
	Config ExporterConfig     `mapstructure:"config"`
	Mqtt   ExporterMqttConfig `mapstructure:"mqtt"`
}

type MappingEntry struct {
	Name      string   `json:"name"`
	Instances []string `json:"instances"`
}

type Entity struct {
	Name        string `json:"group"`
	LastUpdated string `json:"last_updated"`
}

type Configuration struct {
	Mappings map[string]MappingEntry `json:"mappings"`
	Prefix   string                  `json:"prefix"`
}

func metricName(group string, name string) string {
	result := configuration.Prefix
	result += fmt.Sprintf("%s_%s", group, name)
	return result
}

func metricHelp(group string, name string) string {
	return fmt.Sprintf("new mqttexporter: Name: '%s_%s'", group, name)
}

func metricType(m MappingEntry) (prometheus.ValueType, error) {
	return prometheus.GaugeValue, nil
}

func metricKey(host string, group string, name string, instance string, index string, vindex int) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s-%d", host, group, name, instance, index, vindex)
}

type newmqttSample struct {
	Id      string
	Name    string
	Labels  map[string]string
	Help    string
	Value   float64
	DType   string
	Dstype  string
	Time    float64
	Type    prometheus.ValueType
	Unit    string
	Expires time.Time
}

type mqttCollector struct {
	samples map[string]*newmqttSample
	mu      *sync.Mutex
	ch      chan *newmqttSample
}

func newmqttCollector() *mqttCollector {
	c := &mqttCollector{
		ch:      make(chan *newmqttSample, 0),
		mu:      &sync.Mutex{},
		samples: map[string]*newmqttSample{},
	}
	go c.processSamples()
	return c
}

func (c *mqttCollector) processSamples() {
	ticker := time.NewTicker(time.Minute).C
	for {
		select {
		case sample := <-c.ch:
			c.mu.Lock()
			c.samples[sample.Id] = sample
			c.mu.Unlock()
		case <-ticker:
			// Garbage collect expired samples.
			now := time.Now()
			c.mu.Lock()
			for k, sample := range c.samples {
				if now.After(sample.Expires) {
					delete(c.samples, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Collect implements prometheus.Collector.
func (c mqttCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- lastPush

	c.mu.Lock()
	samples := make([]*newmqttSample, 0, len(c.samples))
	for _, sample := range c.samples {
		samples = append(samples, sample)
	}
	c.mu.Unlock()

	now := time.Now()
	for _, sample := range samples {
		if now.After(sample.Expires) {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(sample.Name, sample.Help, []string{}, sample.Labels), sample.Type, sample.Value,
		)
	}
}

// Describe implements prometheus.Collector.
func (c mqttCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lastPush.Desc()
}

func parseValue(e string) (float64, error) {
	val, err := strconv.ParseFloat(e, 64)
	if err == nil {
		return val, err
	}
	if config.Config.Verbose {
		log.Debugf("Error parsing state=%s", e)
	}
	return -1.0, errors.New("Unvalid value")
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var data = msg.Payload()
	var stData = string(data[:len(data)-1])
	//	log.Infof("Received message: %s from topic: %s", stData, msg.Topic())

	var partsMessage = strings.Split(stData, ":")
	if len(partsMessage) > 1 {
		log.Debugf("Received value: %s from topic: %s", partsMessage[1], msg.Topic())

		var partsTopic = strings.Split(msg.Topic(), "/")
		if len(partsTopic) > 2 {
			if partsTopic[0] == "collectd" {
				var host = partsTopic[1]
				var groupPart = partsTopic[2]
				var group = ""
				var cinstance = ""
				var indexGroup = strings.Index(groupPart, "-")
				if indexGroup > 0 {
					group = groupPart[:indexGroup]
					cinstance = groupPart[indexGroup+1:]
				} else {
					group = groupPart
				}
				var namePart = partsTopic[3]
				var indexName = strings.Index(namePart, "-")
				var name = ""
				var index = ""
				if indexName > 0 {
					name = namePart[:indexName]
					index = namePart[indexName+1:]
				} else {

					name = namePart
				}

				for vindex, value := range partsMessage {
					if vindex == 0 {
						continue
					}
					var key = group + "." + name
					metric, found := configuration.Mappings[key]
					if found {
						keep := true
						var instances = metric.Instances
						if instances != nil {
							keep = false
							for _, instanceI := range instances {
								if instanceI == cinstance {
									keep = true
									break
								}
							}
						}
						if !keep {
							return
						}
						log.Debugf("Received value: %s for host %s for group %s[%s] and name %s[%s]/%d", value, host, group, cinstance, name, index, vindex)
						valueF, err := parseValue(value)
						if err == nil {
							now := time.Now()
							lastPush.Set(float64(now.UnixNano()) / 1e9)
							metricType, err := metricType(metric)
							if err == nil {
								labels := prometheus.Labels{}
								if host != "" {
									labels["host"] = host
								}
								if cinstance != "" {
									labels["cinstance"] = cinstance
								}
								if index != "" {
									labels["index"] = index
								}
								if len(partsMessage) > 2 {
									labels["vindex"] = fmt.Sprintf("%d", vindex-1)
								}
								collector.ch <- &newmqttSample{
									Id:      metricKey(host, group, name, cinstance, index, vindex),
									Name:    metricName(group, name),
									Labels:  labels,
									Help:    metricHelp(group, name),
									Value:   valueF,
									Type:    metricType,
									Expires: now.Add(time.Duration(300) * time.Second * 2),
								}
							}
						}
					}
				}
			}
		}
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Warnf("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Warnf("Connect lost: %v", err)
}

func startExporter() {

	if config.Config.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	configurationFile, err := os.Open(config.Config.ConfigurationFile)
	if err == nil {
		log.Info("Parsing Configuration file")
		byteValue, _ := ioutil.ReadAll(configurationFile)
		json.Unmarshal(byteValue, &configuration)
		if config.Config.Verbose {
			log.Debug(configuration)
		}
		log.Infof("Parsing Configuration file: %d entries", len(configuration.Mappings))
		defer configurationFile.Close()
	} else {
		log.Fatalf("Failed to open configuration file: %s", config.Config.ConfigurationFile)
	}

	collector = newmqttCollector()
	prometheus.MustRegister(collector)

	log.Info("Listening on " + config.Config.ListeningAddress)
	http.Handle(config.Config.MetricsPath, promhttp.Handler())

	opts := mqtt.NewClientOptions()
	opts.SetClientID("mqtt_exporter_client")
	opts.AddBroker(config.Mqtt.Broker)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Info("Connected to MQTT broker " + config.Mqtt.Broker)
	topic := config.Mqtt.Topic
	token := client.Subscribe(topic, 1, nil)
	log.Info("Waiting for messages")
	token.Wait()

	http.ListenAndServe(config.Config.ListeningAddress, nil)
}

func LoadConfig(path string) (err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("mqtt_exporter")
	viper.SetConfigType("json")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	defaults.SetDefaults(&config)
	err = viper.Unmarshal(&config)

	if viper.IsSet("verbose") {
		config.Config.Verbose = viper.GetBool("verbose")
	}

	return
}

func main() {
	viper.SetEnvPrefix("MQTT_EXPORTER")

	err := LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	startExporter()
}
