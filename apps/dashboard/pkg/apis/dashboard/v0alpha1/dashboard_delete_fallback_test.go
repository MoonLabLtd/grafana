package v0alpha1

import (
	"context"
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockDashboardClient struct {
	items       []Dashboard
	deleteCalls int
	deleteFail  bool
}

func (m *mockDashboardClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return &metav1.StatusError{
		ErrStatus: metav1.Status{
			Code:   405,
			Reason: metav1.StatusReasonMethodNotAllowed,
		},
	}
}

func (m *mockDashboardClient) List(ctx context.Context, opts metav1.ListOptions) (*DashboardList, error) {
	return &DashboardList{Items: m.items}, nil
}

func (m *mockDashboardClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	m.deleteCalls++
	if m.deleteFail {
		return errors.New("delete failed")
	}
	return nil
}

func TestDeleteAllDashboardsWithFallback_Empty(t *testing.T) {
	client := &mockDashboardClient{items: []Dashboard{}}
	err := DeleteAllDashboardsWithFallback(context.Background(), client, "org-1", metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if client.deleteCalls != 0 {
		t.Fatalf("expected 0 delete calls, got %d", client.deleteCalls)
	}
}

func TestDeleteAllDashboardsWithFallback_NonEmpty(t *testing.T) {
	client := &mockDashboardClient{
		items: []Dashboard{
			{ObjectMeta: metav1.ObjectMeta{Name: "dash1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "dash2"}},
		},
	}
	err := DeleteAllDashboardsWithFallback(context.Background(), client, "org-2", metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if client.deleteCalls != 2 {
		t.Fatalf("expected 2 delete calls, got %d", client.deleteCalls)
	}
}

func TestDeleteAllDashboardsWithFallback_RetryFailure(t *testing.T) {
	client := &mockDashboardClient{
		items: []Dashboard{
			{ObjectMeta: metav1.ObjectMeta{Name: "dash1"}},
		},
		deleteFail: true,
	}
	err := DeleteAllDashboardsWithFallback(context.Background(), client, "org-3", metav1.DeleteOptions{}, metav1.ListOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if client.deleteCalls != 3 {
		t.Fatalf("expected 3 delete calls (retries), got %d", client.deleteCalls)
	}
}
