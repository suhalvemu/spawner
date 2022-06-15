package gcp

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	"gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	container_proto "google.golang.org/genproto/googleapis/container/v1"
)

//getNodeFQN retrive fqdn for the node
func getNodePoolFQN(projectId, location, clusterId, nodeId string) string {
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", projectId, location, clusterId, nodeId)
}

//getDefaultDiskType return the name of the disk type
func getDefaultDiskType() string {
	// Doc : https://cloud.google.com/compute/docs/disks
	return "pd-standard"
}

func getImageType() string {
	//Doc : https://cloud.google.com/kubernetes-engine/docs/concepts/node-images#available_node_images
	return "COS_CONTAINERD"
}

//getNodePool Get the NodePool config for the given NodeSpec
func getNodePool(node *spawner.NodeSpec) (*container_proto.NodePool, error) {

	nodeCount := int32(1)
	if node.Count != 0 {
		nodeCount = int32(node.Count)
	}
	//node labels
	label := make(map[string]string)
	for k, v := range labels.GetNodeLabel(node) {
		label[k] = *v
	}

	instance := ""
	if node.MachineType != "" {
		instance = common.GetInstance(constants.GcpLabel, node.MachineType)
	} else {
		instance = node.Instance
	}

	if instance == "" {
		return nil, errors.New(constants.InvalidInstanceOrMachineType)
	}
	imageType := getImageType()
	diskType := getDefaultDiskType()

	nodeConfig := &container_proto.NodeConfig{

		MachineType: instance,
		DiskSizeGb:  node.DiskSize,
		ImageType:   imageType,
		Preemptible: false,
		DiskType:    diskType,
		Labels:      label,
	}
	np := container_proto.NodePool{

		Name:             node.Name,
		Config:           nodeConfig,
		InitialNodeCount: nodeCount,
		//in case we use Zonal, if not specied gcp will create per zone node, so 1/2, resulting in 3 nodes default,
		//us-central1-a, us-central1-b, us-central1-c, us-central1-f
		//FIXME: should get the node zones from the regions
		Locations: []string{"us-central1-a"},
		Autoscaling: &container_proto.NodePoolAutoscaling{
			Enabled:      true,
			MinNodeCount: nodeCount,
			MaxNodeCount: nodeCount,
		},
		Management: &container_proto.NodeManagement{
			//Disable the auto repair for now, seems like it will start when creating nodes and wont let us do anything on the cluster
			//might be because of cluster repair is also running, but dont have badnwidth to check, so disabling noderepair too
			AutoRepair:  false,
			AutoUpgrade: false,
		},
	}
	return &np, nil
}
func (g *GCPController) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "AddNode ")
	}

	client, err := getClusterManagerClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "AddNode")
	}

	defer client.Close()

	cfqn := getClusterFQName(cred.ProjectId, req.Region, req.ClusterName)
	//get cluster id -- apparently cluster id is the cluster name,
	/*
		cluster, err := client.GetCluster(ctx, &container_proto.GetClusterRequest{
			Name: fqn,
		})

		if err != nil {

			if e, ok := err.(*apierror.APIError); ok {
				st := e.GRPCStatus()
				g.logger.Error(ctx, "cluster not found", "cluster", req.ClusterName, "code", st.Code(), "reason", st.Details())
				if st.Code() == codes.NotFound {
					return nil, errors.New("cluster not found")
				}
				//TODO: might be some other rpc error
				return nil, errors.Wrap(st.Err(), "cluster not found")
			}
			return nil, errors.Wrap(err, "getCluster:")
		}
	*/

	np, err := getNodePool(req.NodeSpec)

	if err != nil {
		g.logger.Error(ctx, "failed to get the nodepool config from the spec", "error", err)
		return nil, errors.Wrap(err, "AddNode: getNodePool returned error")
	}

	nodeReq := container_proto.CreateNodePoolRequest{
		NodePool: np,
		Parent:   cfqn,
	}

	r, err := client.CreateNodePool(ctx, &nodeReq)
	if err != nil {
		g.logger.Error(ctx, "failed to create a nodepool", "error", err)
		return nil, errors.Wrap(err, "AddNode: create node pool failed")
	}

	g.logger.Info(ctx, "nodepool created", "status", r.Status, "cluster", req.ClusterName, "nodepool", np.Name)
	return &proto.NodeSpawnResponse{}, nil
}

func (g *GCPController) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "delete node ")
	}

	client, err := getClusterManagerClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "delete node")
	}

	defer client.Close()

	r, err := client.DeleteNodePool(ctx, &container_proto.DeleteNodePoolRequest{
		Name: getNodePoolFQN(cred.ProjectId, req.Region, req.ClusterName, req.NodeGroupName),
	})

	if err != nil {
		g.logger.Error(ctx, "failed to delete the nodepool", "error", err)
		return nil, errors.Wrap(err, "failed to delete to node pool")
	}
	g.logger.Info(ctx, "deleting nodepool", "status", r.Status, "cluster", req.ClusterName, "nodepool", req.NodeGroupName)
	return &proto.NodeDeleteResponse{}, nil
}
