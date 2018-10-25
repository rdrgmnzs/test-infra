package objectstorage

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/go-cloud/blob"
	"github.com/google/go-cloud/blob/gcsblob"
	"github.com/google/go-cloud/blob/s3blob"
	"github.com/google/go-cloud/gcp"
)

func setupAWS(ctx context.Context, options map[string]string) (*blob.Bucket, error) {
	sessionOptions := session.Options{
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},
	}

	sessionSession, err := session.NewSessionWithOptions(sessionOptions)
	if err != nil {
		return nil, err
	}
	bucket, err := awsBucket(ctx, sessionSession, options)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func setupGCP(ctx context.Context, options map[string]string) (*blob.Bucket, error) {
	roundTripper := gcp.DefaultTransport()
	credentials, err := gcp.DefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	tokenSource := gcp.CredentialsTokenSource(credentials)
	httpClient, err := gcp.NewHTTPClient(roundTripper, tokenSource)
	if err != nil {
		return nil, err
	}
	bucket, err := gcpBucket(ctx, httpClient, options)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func awsBucket(ctx context.Context, cp client.ConfigProvider, options map[string]string) (*blob.Bucket, error) {
	u, err := url.Parse(options["bucket"])
	if err != nil {
		return nil, err
	}

	return s3blob.OpenBucket(ctx, cp, u.Host)
}

func gcpBucket(ctx context.Context, client2 *gcp.HTTPClient, options map[string]string) (*blob.Bucket, error) {
	u, err := url.Parse(options["bucket"])
	if err != nil {
		return nil, err
	}

	o := &gcsblob.Options{}
	return gcsblob.OpenBucket(ctx, u.Host, client2, o)
}
