package metric

import (
	"context"
	"math/rand"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

func NewCounter(project, metric string, labels map[string]string) (*Counter, error) {
	return &Counter{
		projectID:    project,
		metric:       metric,
		metricLabels: labels,
	}, nil
}

type Counter struct {
	projectID    string
	metric       string
	metricLabels map[string]string
}

func (r *Counter) Count(ctx context.Context, resource string, resourceLabels map[string]string, count int64) error {
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating client")
	}
	defer c.Close()
	now := &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + r.projectID,
		TimeSeries: []*monitoringpb.TimeSeries{{
			Metric: &metricpb.Metric{
				Type:   r.metric,
				Labels: r.metricLabels,
			},
			Resource: &monitoredres.MonitoredResource{
				Type:   resource,
				Labels: resourceLabels,
			},
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					StartTime: now,
					EndTime:   now,
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_Int64Value{
						Int64Value: rand.Int63n(count),
					},
				},
			}},
		}},
	}

	err = c.CreateTimeSeries(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "writing time series request: %+v", req)
	}
	return nil
}
