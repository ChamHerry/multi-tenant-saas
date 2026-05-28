package quota

import (
	"strings"
	"testing"

	"repomind-temp/internal/service"
)

func TestFeatureKey(t *testing.T) {
	cases := map[service.QuotaMetric]string{
		service.MetricAPIKeyCount: "api_key.max_count",
		service.MetricMemberCount: "member.max_count",
	}
	for metric, want := range cases {
		if got := featureKey(metric); got != want {
			t.Fatalf("featureKey(%s)=%s want %s", metric, got, want)
		}
	}
	if got := featureKey(service.QuotaMetric("custom.count")); got != "" {
		t.Fatalf("featureKey(custom.count)=%s want empty", got)
	}
}

func TestUsageQueryOnlySupportsTemplateMetrics(t *testing.T) {
	cases := map[service.QuotaMetric]string{
		service.MetricAPIKeyCount: "public.api_keys",
		service.MetricMemberCount: "public.tenant_memberships",
	}
	for metric, wantFragment := range cases {
		got, err := usageQuery(metric)
		if err != nil {
			t.Fatalf("usageQuery(%s) error = %v", metric, err)
		}
		if !strings.Contains(got, wantFragment) {
			t.Fatalf("usageQuery(%s)=%s missing %s", metric, got, wantFragment)
		}
	}
	if _, err := usageQuery(service.QuotaMetric("custom.count")); err == nil {
		t.Fatal("usageQuery(custom.count) expected invalid metric error")
	}
}

func TestValidateReservationInput(t *testing.T) {
	valid := service.QuotaReservationInput{TenantID: "123e4567-e89b-12d3-a456-426614174000", Metric: service.MetricAPIKeyCount, Delta: 1, ResourceType: "api_key"}
	if err := validateReservationInput(valid); err != nil {
		t.Fatalf("validateReservationInput(valid) error = %v", err)
	}
	valid.Delta = 0
	if err := validateReservationInput(valid); err == nil {
		t.Fatal("validateReservationInput(delta=0) expected error")
	}
}
