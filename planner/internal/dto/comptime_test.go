package dto

import (
	"testing"
)

func TestComptimeListRequest_Validation(t *testing.T) {
	req := &ComptimeListRequest{
		VendorID:   0,
		StartMonth: "2026-03",
		EndMonth:   "2026-06",
	}

	if err := req.Validate(nil); err != nil {
		t.Fatalf("expected valid request, got err: %v", err)
	}
	if req.VendorID != 1 {
		t.Errorf("expected default vendor_id 1, got %d", req.VendorID)
	}
	if req.Page != 1 || req.PageSize != 50 {
		t.Errorf("expected default pagination 1/50, got %d/%d", req.Page, req.PageSize)
	}

	filter := req.ToFilter()
	if filter.StartMonth == nil || filter.StartMonth.Month() != 3 {
		t.Errorf("expected StartMonth to be March, got %v", filter.StartMonth)
	}

	// Invalid date
	invalidReq := &ComptimeListRequest{
		StartMonth: "invalid-date",
	}
	if err := invalidReq.Validate(nil); err == nil {
		t.Errorf("expected error for invalid start_month")
	}
}

func TestComptimeResyncRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     ComptimeResyncRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: ComptimeResyncRequest{
				VendorID:  1,
				StartDate: "2026-01-01",
				EndDate:   "2026-03-31",
				DataType:  "all",
			},
			wantErr: false,
		},
		{
			name: "missing start date",
			req: ComptimeResyncRequest{
				VendorID: 1,
				EndDate:  "2026-03-31",
			},
			wantErr: true,
		},
		{
			name: "invalid end date",
			req: ComptimeResyncRequest{
				VendorID:  1,
				StartDate: "2026-01-01",
				EndDate:   "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid data type",
			req: ComptimeResyncRequest{
				VendorID:  1,
				StartDate: "2026-01-01",
				EndDate:   "2026-03-31",
				DataType:  "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
