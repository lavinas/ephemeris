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
			name: "valid request with customer and service and todos datatype",
			req: func() ComptimeResyncRequest {
				c := int64(10)
				s := int64(20)
				return ComptimeResyncRequest{
					VendorID:   1,
					StartDate:  "2026-01-01",
					EndDate:    "2026-03-31",
					DataType:   "todos",
					CustomerID: &c,
					ServiceID:  &s,
				}
			}(),
			wantErr: false,
		},
		{
			name: "normalize negative customer and service to nil",
			req: func() ComptimeResyncRequest {
				c := int64(-1)
				s := int64(0)
				return ComptimeResyncRequest{
					VendorID:   1,
					StartDate:  "2026-01-01",
					EndDate:    "2026-03-31",
					DataType:   "",
					CustomerID: &c,
					ServiceID:  &s,
				}
			}(),
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

func TestComptimeDeleteRequest_Validation(t *testing.T) {
	reqValid := ComptimeDeleteRequest{ID: 10}
	if err := reqValid.Validate(); err != nil {
		t.Errorf("expected valid for ID 10, got err: %v", err)
	}

	reqInvalid := ComptimeDeleteRequest{ID: 0}
	if err := reqInvalid.Validate(); err == nil {
		t.Errorf("expected error for ID 0")
	}

	reqNegative := ComptimeDeleteRequest{ID: -5}
	if err := reqNegative.Validate(); err == nil {
		t.Errorf("expected error for ID -5")
	}
}

func TestComptimeAdjustRequest_Validation(t *testing.T) {
	// 1. Valid by ComptimeID
	reqByID := ComptimeAdjustRequest{
		ComptimeID:      10,
		AdjustedMinutes: 45,
		Description:     "Bônus contratual",
	}
	if err := reqByID.Validate(nil); err != nil {
		t.Errorf("expected valid request by ComptimeID, got: %v", err)
	}

	// 2. Valid by customer/service/month with negative minutes
	reqByMonth := ComptimeAdjustRequest{
		CustomerID:      100,
		ServiceID:       200,
		Month:           "2026-04",
		AdjustedMinutes: -30,
		Description:     "Estorno de sessão duplicada",
	}
	if err := reqByMonth.Validate(nil); err != nil {
		t.Errorf("expected valid request by customer/service/month, got: %v", err)
	}

	// 3. Missing description when AdjustedMinutes != 0
	reqNoDesc := ComptimeAdjustRequest{
		ComptimeID:      10,
		AdjustedMinutes: 15,
		Description:     "   ",
	}
	if err := reqNoDesc.Validate(nil); err == nil {
		t.Errorf("expected error when description is empty")
	}

	// 4. Zero minutes is valid even without description (acts as zero/reset)
	reqZero := ComptimeAdjustRequest{
		ComptimeID:      10,
		AdjustedMinutes: 0,
	}
	if err := reqZero.Validate(nil); err != nil {
		t.Errorf("expected zero minutes to be valid, got: %v", err)
	}

	// 5. Missing identifier and customer/service/month
	reqMissingIDs := ComptimeAdjustRequest{
		AdjustedMinutes: 30,
		Description:     "Justificativa",
	}
	if err := reqMissingIDs.Validate(nil); err == nil {
		t.Errorf("expected error when both comptime_id and customer/service/month are missing")
	}
}
