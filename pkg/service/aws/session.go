package aws

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

type Session struct {
	AwsSession *session.Session
	Region     string
	TeamId     string
}

func getSessionUsingAwsAccessKey(region string) (*session.Session, error) {

	conf := config.Get()

	cred := credentials.NewStaticCredentials(conf.AWSAccessID, conf.AWSSecretKey, conf.AWSToken)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: cred,
	})

	return sess, err
}

func getSessionUsingAwsProfile(profile string) (*session.Session, error) {

	sess, err := session.NewSessionWithOptions(
		session.Options{
			Profile: profile,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "unable to find the credential config in system")
	}

	//NewSessionWithOptions doesn't return error if we pass the invalid profile, checking here
	// if we got the credentials correctly
	_, err = sess.Config.Credentials.Get()

	if err != nil {
		return nil, errors.Wrap(err, "unable to find the credential config in system")
	}

	return sess, nil

}

func getLocalEnvSession(region string) (*session.Session, error) {

	if config.Get().AWSAccessID != "" {
		// creating creds from config variables
		return getSessionUsingAwsAccessKey(region)

	}

	log.Println("AWSAccessID is empty, using 'default' profile from user's home directory")
	return getSessionUsingAwsProfile("default")

}

//NewSession create a session for given user account by fetching user credentials from the secret stores.
// if running locally and env set to local, it will use credentials from the config
func NewSession(ctx context.Context, region string, accountName string) (*Session, error) {

	var (
		awsCreds *credentials.Credentials
		err      error
	)

	conf := config.Get()
	if conf.Env == "local" {
		sess, err := getLocalEnvSession(region)

		if err != nil {
			return nil, err
		}

		return &Session{
			AwsSession: sess,
			Region:     region,
			TeamId:     accountName,
		}, nil

	} else {
		//secret manager is hosted in particular region, all writes happen to the same region
		secretHostRegion := config.Get().SecretHostRegion
		awsCreds, err = system.GetAwsCredentials(ctx, secretHostRegion, accountName)
		if err != nil {
			return nil, err
		}
	}

	//get credentials for the user of given team id
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: awsCreds,
	})

	if err != nil {
		return nil, err
	}

	return &Session{
		TeamId:     accountName,
		Region:     region,
		AwsSession: sess,
	}, nil
}

func newKubeConfig(session *session.Session, cluster *eks.Cluster) (*rest.Config, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: aws.StringValue(cluster.Name),
		Session:   session,
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}
	return &rest.Config{
		Host:        aws.StringValue(cluster.Endpoint),
		BearerToken: tok.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	}, nil
}

func newClientset(session *session.Session, cluster *eks.Cluster) (*kubernetes.Clientset, error) {
	config, err := newKubeConfig(session, cluster)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

//---

func (ses *Session) getEksClient() *eks.EKS {
	return eks.New(ses.AwsSession)
}

func (ses *Session) getEC2Client() *ec2.EC2 {
	return ec2.New(ses.AwsSession)
}

func (ses *Session) getCostExplorerClient() *costexplorer.CostExplorer {
	return costexplorer.New(ses.AwsSession)
}

//deprecating, as the client usage is only for fetching the account id, we can have a function to get the account id
func (ses *Session) _getSTSClient() *sts.STS {
	return sts.New(ses.AwsSession)
}

func (ses *Session) getK8sClient(cluster *eks.Cluster) (*kubernetes.Clientset, error) {
	return newClientset(ses.AwsSession, cluster)
}

func (ses *Session) getIAMClient() *iam.IAM {
	return iam.New(ses.AwsSession)
}

func (ses *Session) getKubeConfig(cluster *eks.Cluster) (*rest.Config, error) {
	return newKubeConfig(ses.AwsSession, cluster)
}

func (ses *Session) getRoute53Client() *route53.Route53 {
	return route53.New(ses.AwsSession)
}

func (ses *Session) getEcrClient() *ecr.ECR {
	return ecr.New(ses.AwsSession)
}

func (ses *Session) getK8sDynamicClient(cluster *eks.Cluster) (dynamic.Interface, error) {
	config, err := newKubeConfig(ses.AwsSession, cluster)

	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return dynamicClient, nil
}

func (ses *Session) getAccountId() (string, error) {

	client := ses._getSTSClient()
	callerIdentity, err := client.GetCallerIdentity(nil)
	if err != nil {
		return "", err
	}
	return *callerIdentity.Account, nil
}
