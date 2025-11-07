package transport

import (
	"testing"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
)

func TestParseArtworkStatuses(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		want    []domain.ArtworkStatus
		wantErr bool
	}{
		{
			name:   "valid single status",
			values: []string{"available"},
			want:   []domain.ArtworkStatus{domain.ArtworkStatusAvailable},
			wantErr: false,
		},
		{
			name:   "valid multiple statuses",
			values: []string{"available", "sold", "not_for_sale"},
			want: []domain.ArtworkStatus{
				domain.ArtworkStatusAvailable,
				domain.ArtworkStatusSold,
				domain.ArtworkStatusNotForSale,
			},
			wantErr: false,
		},
		{
			name:    "invalid status",
			values:  []string{"invalid_status"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "mix of valid and invalid",
			values:  []string{"available", "invalid"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty slice",
			values:  []string{},
			want:    []domain.ArtworkStatus{},
			wantErr: false,
		},
		{
			name:   "all valid statuses",
			values: []string{"available", "sold", "not_for_sale", "unavailable", "coming_soon"},
			want: []domain.ArtworkStatus{
				domain.ArtworkStatusAvailable,
				domain.ArtworkStatusSold,
				domain.ArtworkStatusNotForSale,
				domain.ArtworkStatusUnavailable,
				domain.ArtworkStatusComingSoon,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArtworkStatuses(tt.values)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArtworkStatuses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("parseArtworkStatuses() returned %d statuses, want %d", len(got), len(tt.want))
					return
				}
				
				for i, status := range got {
					if status != tt.want[i] {
						t.Errorf("parseArtworkStatuses()[%d] = %v, want %v", i, status, tt.want[i])
					}
				}
			}
		})
	}
}
