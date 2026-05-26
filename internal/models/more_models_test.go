package models

import "testing"

func TestPartnerStatusesAndHelpers(t *testing.T) {
	validSiteStatuses := []PartnerSiteStatus{
		PartnerSiteStatusDraft,
		PartnerSiteStatusPendingReview,
		PartnerSiteStatusActive,
		PartnerSiteStatusRejected,
		PartnerSiteStatusBlocked,
		PartnerSiteStatusArchived,
	}
	for _, status := range validSiteStatuses {
		if !status.IsValid() {
			t.Fatalf("expected valid site status: %s", status)
		}
	}
	if PartnerSiteStatus("weird").IsValid() {
		t.Fatal("expected invalid site status")
	}

	if !PartnerBlockStatusActive.IsValid() || !PartnerBlockStatusInactive.IsValid() {
		t.Fatal("expected valid partner block statuses")
	}
	if PartnerBlockStatus("draft").IsValid() {
		t.Fatal("expected invalid partner block status")
	}

	value := 42
	if got := NullInt64FromPtr(&value); !got.Valid || got.Int64 != 42 {
		t.Fatalf("unexpected null int64: %#v", got)
	}
	if got := NullInt64FromPtr(nil); got.Valid {
		t.Fatalf("expected invalid null int64, got %#v", got)
	}
}
