package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func (a *awsController) PresignS3Url(ctx context.Context, req *proto.PresignS3UrlRequest) (*proto.PresignS3UrlResponse, error) {

	//default timeout for signed url
	timeout := time.Duration(10)

	// read the timeout for url from the request, if its not set, defualts to 10min
	if req.TimeoutInMinute != 0 {
		timeout = time.Duration(req.TimeoutInMinute)
	}
	requestExpireTime := timeout * time.Minute

	//creating session
	session, err := NewSession(ctx, req.Region, req.AccountName)

	if err != nil {
		a.logger.Error(ctx, "failed to create a new aws session", "error", err)
		return nil, errors.Wrap(err, "PresignS3Url ")
	}

	s3Client := session.getS3Client()

	file := req.File
	//must start with `/`
	if !strings.HasPrefix(file, "/") {
		// we will prefix with / if not exist.
		file = fmt.Sprintf("/%s", file)
	}

	request, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(req.Bucket),
		Key:    &file,
	})

	// NOTE: there is no context version of this API in this implementation.
	// We need v2 api's, which is out of the scope for the migration
	request.SetContext(ctx)
	url, err := request.Presign(requestExpireTime)
	if err != nil {
		a.logger.Error(ctx, "failed to presign bucket resource", "bucket", req.Bucket, "file", req.File)
		return nil, errors.Wrap(err, "PresignS3Url: failed to presign")
	}

	a.logger.Info(ctx, "bucket file signed successfully", "url", url, "bucket", req.Bucket, "file", req.File)
	return &proto.PresignS3UrlResponse{
		SignedUrl: url,
	}, nil
}
