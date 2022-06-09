package es

import (
	"context"
	"time"

	elastic "github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"

	"go.k6.io/k6/output"
)

type XK6ElasticSample struct {
	Name  string            `json:"name"`
	Type  string            `json:"type"`
	Time  int64             `json:"time"`
	Tags  map[string]string `tags:"tags"`
	Value float64           `tags:"value"`
}

// Output implements the lib.Output interface
type Output struct {
	output.SampleBuffer

	config          Config
	periodicFlusher *output.PeriodicFlusher
	logger          logrus.FieldLogger
	esClient        *elastic.Client
}

var _ output.Output = new(Output)

// New creates an instance of the collector
func New(p output.Params) (*Output, error) {
	conf, err := NewConfig(p)
	if err != nil {
		return nil, err
	}
	// Some setupping code

	client, err := elastic.NewClient(
		elastic.SetURL(conf.Address),
		elastic.SetBasicAuth(conf.Username, conf.Password),
		elastic.SetSniff(conf.SnifferEnabled),
	)
	if err != nil {
		p.Logger.WithError(err).Error("Error when creating elasticsearch client")
		client = &elastic.Client{}
	}

	return &Output{
		config:   conf,
		logger:   p.Logger,
		esClient: client,
	}, nil
}

func (o *Output) Description() string {
	return "elasticsearch output: " + o.config.Address
}

func (o *Output) Stop() error {
	o.logger.Debug("Stopping...")
	defer o.logger.Debug("Stopped!")
	o.periodicFlusher.Stop()
	return nil
}

func (o *Output) Start() error {
	o.logger.Debug("Starting...")

	pf, err := output.NewPeriodicFlusher(o.config.PushInterval, o.flushMetrics)
	if err != nil {
		return err
	}
	o.logger.Debug("Started!")
	o.periodicFlusher = pf

	return nil
}

func (o *Output) flushMetrics() {
	samples := o.GetBufferedSamples()
	start := time.Now()
	var count int

	bulkRequest := o.esClient.Bulk()

	for _, sc := range samples {
		samples := sc.GetSamples()
		count += len(samples)

		for _, sample := range samples {
			esSample := XK6ElasticSample{
				Name:  sample.Metric.Name,
				Type:  sample.Metric.Type.String(),
				Tags:  sample.Tags.CloneTags(),
				Time:  sample.Time.UnixMilli(),
				Value: sample.Value,
			}

			// This is too slow ...
			// o.esClient.Index().
			// 	Index(o.config.Index).
			// 	BodyJson(esSample).
			// 	Do(ctx)

			// bulk is better !
			bulkRequest = bulkRequest.Add(
				elastic.NewBulkIndexRequest().OpType("create").Index(o.config.Index).Doc(esSample),
			)

			// too spammy
			// o.logger.WithField("name", sample.Metric.Name).WithField("type", sample.Metric.Type.String()).WithField("value", sample.Value).Debug("Metric added in bulk")
		}

	}
	if count > 0 {
		o.logger.WithField("t", time.Since(start)).WithField("count", count).Debug("Flushing metrics to elasticsearch")
		_, err := bulkRequest.Do(context.TODO())
		if err != nil {
			o.logger.WithError(err).Debug("Bulk request failed")
		}
	}
}
