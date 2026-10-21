package resources

import (
	"encoding/json"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStorageReviewExplicitEmptyJSONSurvivesReadback(t *testing.T) {
	platformCors := json.RawMessage(`{"CORSRules":[{"AllowedOrigins":["https://console.danubedata.ro"],"AllowedMethods":["GET","PUT","POST","DELETE","HEAD"],"AllowedHeaders":["*"],"ExposeHeaders":["ETag","x-amz-request-id","x-amz-id-2"],"MaxAgeSeconds":86400}]}`)
	for _, tc := range []struct {
		name    string
		bucket  client.StorageBucket
		initial StorageBucketResourceModel
		want    string
	}{
		{"lifecycle", client.StorageBucket{LifecycleRules: json.RawMessage(`[]`)}, StorageBucketResourceModel{LifecycleRules: types.StringValue(`[]`)}, `[]`},
		{"cors", client.StorageBucket{CorsConfiguration: platformCors}, StorageBucketResourceModel{CorsConfiguration: types.StringValue(`{"CORSRules":[]}`)}, `{"CORSRules":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			(&StorageBucketResource{}).mapBucketToState(&tc.bucket, &tc.initial)
			var got types.String
			if tc.name == "cors" {
				got = tc.initial.CorsConfiguration
			} else {
				got = tc.initial.LifecycleRules
			}
			if got.IsNull() || got.ValueString() != tc.want {
				t.Errorf("readback = %q, want %q", got.ValueString(), tc.want)
			}
		})
	}
}

func TestStorageReviewRejectsDuplicatePermissionResponse(t *testing.T) {
	want := map[string]string{"bucket-a": "read", "bucket-b": "full"}
	got := []client.StorageKeyBucketPermission{
		{BucketID: "bucket-a", Level: "read"},
		{BucketID: "bucket-a", Level: "read"},
	}
	if sameBucketPermissions(want, got) {
		t.Fatal("duplicate bucket-a must not hide missing bucket-b")
	}
}

func TestStorageReviewRejectsUnknownPermissionWithEmptyLevel(t *testing.T) {
	want := map[string]string{"bucket-a": "read"}
	got := []client.StorageKeyBucketPermission{{BucketID: "bucket-b", Level: ""}}
	if sameBucketPermissions(want, got) {
		t.Fatal("unknown bucket with an empty level must not match the requested scope")
	}
}
