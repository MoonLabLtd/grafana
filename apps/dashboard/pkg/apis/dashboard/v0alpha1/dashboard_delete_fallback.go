package v0alpha1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	fallbackCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dashboard_deletecollection_fallback_total",
		Help: "Counts fallbacks due to 405 Method Not Allowed.",
	})
	durationHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "org_deletion_duration_seconds",
		Help: "Time spent deleting dashboards for an organization.",
	})
)

// DashboardClientFallback represents the client interface needed for deletion
type DashboardClientFallback interface {
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	List(ctx context.Context, opts metav1.ListOptions) (*DashboardList, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

// DeleteAllDashboardsWithFallback intercepts 405 errors and performs individual deletes.
func DeleteAllDashboardsWithFallback(ctx context.Context, client DashboardClientFallback, orgID string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	start := time.Now()
	defer func() {
		durationHistogram.Observe(time.Since(start).Seconds())
	}()

	err := client.DeleteCollection(ctx, opts, listOpts)
	if err == nil {
		return nil
	}

	statusErr, ok := err.(*metav1.StatusError)
	if !ok || (statusErr.ErrStatus.Code != 405 && statusErr.ErrStatus.Reason != metav1.StatusReasonMethodNotAllowed) {
		return err
	}

	fallbackCounter.Inc()

	if listOpts.LabelSelector == "" {
		listOpts.LabelSelector = fmt.Sprintf("grafana.app/orgID=%s", orgID)
	}

	checkOpts := metav1.ListOptions{
		LabelSelector:   listOpts.LabelSelector,
		ResourceVersion: "0",
		Limit:           1,
	}

	list, err := client.List(ctx, checkOpts)
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		return nil
	}

	fullList, err := client.List(ctx, listOpts)
	if err != nil {
		return err
	}

	bg := metav1.DeletePropagationBackground
	var zero int64 = 0
	delOpts := opts
	delOpts.PropagationPolicy = &bg
	delOpts.GracePeriodSeconds = &zero

	var errs []error
	for _, item := range fullList.Items {
		var delErr error
		for i := 0; i < 3; i++ {
			delErr = client.Delete(ctx, item.Name, delOpts)
			if delErr == nil {
				break
			}
		}
		if delErr != nil {
			errs = append(errs, delErr)
		}
	}

	if len(errs) > 0 {
		return errors.New("ErrOrganizationHasResources: unable to delete all dashboards")
	}
	return nil
}
