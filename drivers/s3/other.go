package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/AlliotTech/openalist/internal/errs"
	"github.com/AlliotTech/openalist/internal/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	OtherMethodArchive       = "archive"
	OtherMethodArchiveStatus = "archive_status"
	OtherMethodThaw          = "thaw"
	OtherMethodThawStatus    = "thaw_status"
)

type ArchiveRequest struct {
	StorageClass string `json:"storage_class"`
}

type ThawRequest struct {
	Days int32  `json:"days"`
	Tier string `json:"tier"`
}

type ObjectDescriptor struct {
	Path   string `json:"path"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

type ArchiveResponse struct {
	Action       string           `json:"action"`
	Object       ObjectDescriptor `json:"object"`
	StorageClass string           `json:"storage_class"`
	RequestID    string           `json:"request_id,omitempty"`
	VersionID    string           `json:"version_id,omitempty"`
	ETag         string           `json:"etag,omitempty"`
	LastModified string           `json:"last_modified,omitempty"`
}

type ThawResponse struct {
	Action    string           `json:"action"`
	Object    ObjectDescriptor `json:"object"`
	RequestID string           `json:"request_id,omitempty"`
	Status    *RestoreStatus   `json:"status,omitempty"`
}

type RestoreStatus struct {
	Ongoing bool   `json:"ongoing"`
	Expiry  string `json:"expiry,omitempty"`
	Raw     string `json:"raw"`
}

func (d *S3) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	if args.Obj == nil {
		return nil, fmt.Errorf("missing object reference")
	}
	if args.Obj.IsDir() {
		return nil, errs.NotSupport
	}

	switch strings.ToLower(strings.TrimSpace(args.Method)) {
	case OtherMethodArchive:
		return d.archive(ctx, args)
	case OtherMethodArchiveStatus:
		return d.archiveStatus(ctx, args)
	case OtherMethodThaw:
		return d.thaw(ctx, args)
	case OtherMethodThawStatus:
		return d.thawStatus(ctx, args)
	default:
		return nil, errs.NotSupport
	}
}

func (d *S3) archive(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	key := getKey(args.Obj.GetPath(), false)
	payload := ArchiveRequest{}
	if err := decodeOtherArgs(args.Data, &payload); err != nil {
		return nil, fmt.Errorf("parse archive request: %w", err)
	}
	storageClass := normalizeStorageClass(payload.StorageClass)
	if storageClass == "" {
		return nil, fmt.Errorf("storage_class is required")
	}

	output, err := d.client.CopyObject(ctx, &awss3.CopyObjectInput{
		Bucket:            aws.String(d.Bucket),
		Key:               aws.String(key),
		CopySource:        aws.String(url.PathEscape(d.Bucket + "/" + key)),
		MetadataDirective: types.MetadataDirectiveCopy,
		StorageClass:      types.StorageClass(storageClass),
	})
	if err != nil {
		return nil, err
	}

	resp := ArchiveResponse{
		Action:       OtherMethodArchive,
		Object:       d.describeObject(args.Obj, key),
		StorageClass: storageClass,
		VersionID:    aws.ToString(output.VersionId),
	}
	resp.RequestID, _ = awsmiddleware.GetRequestIDMetadata(output.ResultMetadata)
	if output.CopyObjectResult != nil {
		resp.ETag = aws.ToString(output.CopyObjectResult.ETag)
		if output.CopyObjectResult.LastModified != nil {
			resp.LastModified = output.CopyObjectResult.LastModified.UTC().Format(time.RFC3339)
		}
	}
	if status, statusErr := d.describeObjectStatus(ctx, key); statusErr == nil && status.StorageClass != "" {
		resp.StorageClass = status.StorageClass
	}
	return resp, nil
}

func (d *S3) archiveStatus(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	key := getKey(args.Obj.GetPath(), false)
	status, err := d.describeObjectStatus(ctx, key)
	if err != nil {
		return nil, err
	}
	return ArchiveResponse{
		Action:       OtherMethodArchiveStatus,
		Object:       d.describeObject(args.Obj, key),
		StorageClass: status.StorageClass,
	}, nil
}

func (d *S3) thaw(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	key := getKey(args.Obj.GetPath(), false)
	payload := ThawRequest{Days: 1}
	if err := decodeOtherArgs(args.Data, &payload); err != nil {
		return nil, fmt.Errorf("parse thaw request: %w", err)
	}
	if payload.Days <= 0 {
		payload.Days = 1
	}
	restoreRequest := &types.RestoreRequest{Days: aws.Int32(payload.Days)}
	if tier := normalizeRestoreTier(payload.Tier); tier != "" {
		restoreRequest.GlacierJobParameters = &types.GlacierJobParameters{Tier: tier}
	}
	output, err := d.client.RestoreObject(ctx, &awss3.RestoreObjectInput{
		Bucket:         aws.String(d.Bucket),
		Key:            aws.String(key),
		RestoreRequest: restoreRequest,
	})
	if err != nil {
		return nil, err
	}

	resp := ThawResponse{
		Action: OtherMethodThaw,
		Object: d.describeObject(args.Obj, key),
	}
	resp.RequestID, _ = awsmiddleware.GetRequestIDMetadata(output.ResultMetadata)
	if status, statusErr := d.describeObjectStatus(ctx, key); statusErr == nil {
		resp.Status = status.Restore
	}
	return resp, nil
}

func (d *S3) thawStatus(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	key := getKey(args.Obj.GetPath(), false)
	status, err := d.describeObjectStatus(ctx, key)
	if err != nil {
		return nil, err
	}
	return ThawResponse{
		Action: OtherMethodThawStatus,
		Object: d.describeObject(args.Obj, key),
		Status: status.Restore,
	}, nil
}

func (d *S3) describeObject(obj model.Obj, key string) ObjectDescriptor {
	return ObjectDescriptor{Path: obj.GetPath(), Bucket: d.Bucket, Key: key}
}

type objectStatus struct {
	StorageClass string
	Restore      *RestoreStatus
}

func (d *S3) describeObjectStatus(ctx context.Context, key string) (*objectStatus, error) {
	head, err := d.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return &objectStatus{
		StorageClass: string(head.StorageClass),
		Restore:      parseRestoreHeader(head.Restore),
	}, nil
}

func parseRestoreHeader(header *string) *RestoreStatus {
	if header == nil {
		return nil
	}
	value := strings.TrimSpace(*header)
	if value == "" {
		return nil
	}
	status := &RestoreStatus{Raw: value}
	if ongoing, ok := restoreAttribute(value, "ongoing-request"); ok {
		status.Ongoing = strings.EqualFold(ongoing, "true")
	}
	if expiry, ok := restoreAttribute(value, "expiry-date"); ok {
		if parsed, err := time.Parse(time.RFC1123, expiry); err == nil {
			status.Expiry = parsed.UTC().Format(time.RFC3339)
		} else {
			status.Expiry = expiry
		}
	}
	return status
}

func restoreAttribute(header, name string) (string, bool) {
	prefix := name + "=\""
	start := strings.Index(header, prefix)
	if start < 0 {
		return "", false
	}
	value := header[start+len(prefix):]
	end := strings.IndexByte(value, '"')
	if end < 0 {
		return "", false
	}
	return value[:end], true
}

func decodeOtherArgs(data interface{}, target interface{}) error {
	if data == nil {
		return nil
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func normalizeStorageClass(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "-", "_"))
	if value == "" {
		return ""
	}
	if normalized, ok := storageClassLookup[strings.ToLower(value)]; ok {
		return normalized
	}
	return strings.ToUpper(value)
}

func normalizeRestoreTier(value string) types.Tier {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default":
		return ""
	case "bulk":
		return types.TierBulk
	case "standard":
		return types.TierStandard
	case "expedited":
		return types.TierExpedited
	default:
		return types.Tier(strings.TrimSpace(value))
	}
}
