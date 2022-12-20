package metric

import (
	"context"
	"fmt"
	"strings"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/pkg/errors"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	timestamp "google.golang.org/protobuf/types/known/timestamppb"
)

type Counter interface {
	Count(ctx context.Context, metric string, count int64, labels map[string]string) error
}

func NewAPICounter(project string) (Counter, error) {
	return &APICounter{
		projectID: project,
		labels: map[string]string{
			"project_id": project,
		},
	}, nil
}

func MakeMetricType(v string) string {
	return fmt.Sprintf("custom.googleapis.com/%s", strings.ToLower(v))
}

type APICounter struct {
	projectID string
	labels    map[string]string
}

func (r *APICounter) Count(ctx context.Context, metricType string, count int64, labels map[string]string) error {
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating client")
	}
	defer c.Close()
	now := &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}

	if labels == nil {
		labels = map[string]string{}
	}

	// HACK: prevents time series from being overwritten \
	// for timespan which leads to errors on write.
	labels["nanos"] = fmt.Sprintf("e-%d", now.AsTime().UnixMilli())

	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + r.projectID,
		TimeSeries: []*monitoringpb.TimeSeries{{
			Resource: &monitoredres.MonitoredResource{
				Type:   "global",
				Labels: r.labels,
			},
			Metric: &metricpb.Metric{
				Type:   metricType,
				Labels: labels,
			},
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					StartTime: now,
					EndTime:   now,
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_Int64Value{
						Int64Value: count,
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
