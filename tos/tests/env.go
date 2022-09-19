package tests

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type testEnv struct {
	endpoint  string
	region    string
	accessKey string
	secretKey string
	t         *testing.T
}

func newTestEnv(t *testing.T) *testEnv {
	return &testEnv{
		endpoint:  os.Getenv("TOS_GO_SDK_ENDPOINT"),
		region:    os.Getenv("TOS_GO_SDK_REGION"),
		accessKey: os.Getenv("TOS_GO_SDK_AK"),
		secretKey: os.Getenv("TOS_GO_SDK_SK"),
		t:         t,
	}
}

func (e testEnv) prepareClient(bucketName string, extraOptions ...tos.ClientOption) *tos.ClientV2 {
	log := logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.TextFormatter{DisableQuote: true}
	options := []tos.ClientOption{
		tos.WithRegion(e.region),
		tos.WithCredentials(tos.NewStaticCredentials(e.accessKey, e.secretKey)),
		tos.WithEnableVerifySSL(false),
		tos.WithLogger(log),
	}
	options = append(options, extraOptions...)
	client, err := tos.NewClientV2(e.endpoint, options...)
	require.Nil(e.t, err)
	if bucketName != "" {
		create, err := client.CreateBucketV2(context.Background(), &tos.CreateBucketV2Input{
			Bucket: bucketName,
		})
		checkSuccess(e.t, create, err, 200)
	}
	return client
}
