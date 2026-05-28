package quota

import (
	"testing"

	"repomind-temp/internal/service"
)

func TestFeatureKey(t *testing.T) {
	cases := map[service.QuotaMetric]string{
		service.MetricRepoCount:   "repo.max_count",
		service.MetricSymbolCount: "symbol.max_count",
		service.MetricStorageMB:   "storage.max_mb",
		service.MetricAPIKeyCount: "api_key.max_count",
		service.MetricMemberCount: "member.max_count",
	}
	for metric, want := range cases {
		if got := featureKey(metric); got != want {
			t.Fatalf("featureKey(%s)=%s want %s", metric, got, want)
		}
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
