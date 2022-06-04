package aws

import (
	"context"

	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

func asTags(label map[string]string) []*ec2.Tag {
	tags := []*ec2.Tag{}
	for k, v := range label {
		key := k
		val := v
		tags = append(tags, &ec2.Tag{
			Key:   &key,
			Value: &val,
		})
	}

	//add spawner default tags
	for k, v := range labels.DefaultTags() {
		key := k
		val := v
		tags = append(tags, &ec2.Tag{
			Key:   &key,
			Value: val,
		})
	}

	return tags
}

func (a *awsController) addTagToNodeInstances(ctx context.Context, region, clusterName, accountName, nodegroup string, label map[string]string) error {
	session, err := NewSession(ctx, region, accountName)

	a.logger.Debug(ctx, "fetching cluster status for '%s', region '%s'", clusterName, region)
	if err != nil {
		return errors.Wrap(err, "addTag")
	}
	eksClient := session.getEksClient()

	//check if cluster exist
	_, err = getClusterSpec(ctx, eksClient, clusterName)
	if err != nil {
		return errors.Wrap(err, "addTag: ")
	}

	ec := session.getEC2Client()

	clusterKey := "tag:eks:cluster-name"
	nodeKey := "tag:eks:nodegroup-name"
	res, err := ec.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   &clusterKey,
				Values: []*string{&clusterName},
			},
			{
				Name:   &nodeKey,
				Values: []*string{&nodegroup},
			},
		}})

	if err != nil {
		a.logger.Error(ctx, "get instance", "error", err)
		return err
	}

	//NOTE: keep an eye out on this field, dono why this is a list.

	if len(res.Reservations) == 0 {
		a.logger.Info(ctx, "no instances in cluster to tag")
		return errors.New("no instances in cluster to tag")
	}

	rids := []*string{}
	for _, r := range res.Reservations[0].Instances {

		id := *r.InstanceId
		rids = append(rids, &id)
	}

	if len(rids) == 0 {
		//NOTE: cluster may be still creating and hasnt setup ec2 instance yet, wait for node to be created and then try adding again
		return errors.New("no instances in cluster to tag")
	}

	a.logger.Info(ctx, "adding tags to the following resources", "id", rids)

	tags := asTags(label)
	_, err = ec.CreateTags(&ec2.CreateTagsInput{
		Resources: rids,
		Tags:      tags,
	})

	if err != nil {
		return errors.Wrap(err, "addTag")
	}
	return nil

}

//TagNodeInstance tag underlying ec2 instances of a cluster node
func (a *awsController) TagNodeInstance(ctx context.Context, req *proto.TagNodeInstanceRequest) (*proto.TagNodeInstanceResponse, error) {
	err := a.addTagToNodeInstances(ctx, req.Region, req.ClusterName, req.AccountName, req.NodeGroup, req.Labels)
	if err != nil {
		a.logger.Error(ctx, "failed to add tag to node instances in a cluster ", "error", err, "clustere", req.ClusterName)
		return nil, err
	}

	return &proto.TagNodeInstanceResponse{}, nil
}
