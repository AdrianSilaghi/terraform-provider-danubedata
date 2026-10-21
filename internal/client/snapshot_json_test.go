package client

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSnapshotSizesDecodeCurrentAPI(t *testing.T) {
	for _, value := range []string{`"12.75"`, `12.75`, `null`} {
		t.Run(value, func(t *testing.T) {
			want := 12.75
			if value == "null" {
				want = 0
			}
			var vps VpsSnapshot
			var cache CacheSnapshot
			var database DatabaseSnapshot
			for _, tc := range []struct {
				field string
				into  interface{}
				size  *float64
			}{
				{"size_gb", &vps, &vps.SizeGB},
				{"size_mb", &cache, &cache.SizeMB},
				{"size_gb", &database, &database.SizeGB},
			} {
				if err := json.Unmarshal([]byte(fmt.Sprintf(`{"id":19,"%s":%s}`, tc.field, value)), tc.into); err != nil {
					t.Fatalf("%T: %v", tc.into, err)
				}
				if *tc.size != want {
					t.Errorf("%T size = %v, want %v", tc.into, *tc.size, want)
				}
			}
			if vps.SizeKnown != (value != "null") || cache.SizeKnown != (value != "null") || database.SizeKnown != (value != "null") {
				t.Fatal("nullable snapshot size presence lost")
			}
		})
	}
}
