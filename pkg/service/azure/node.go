package azure

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/containerservice/mgmt/containerservice"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func getGPUProfile(profile proto.MIGProfile) containerservice.GPUInstanceProfile {
	switch profile {
	case proto.MIGProfile_MIG1g:
		return containerservice.GPUInstanceProfileMIG1g
	case proto.MIGProfile_MIG2g:
		return containerservice.GPUInstanceProfileMIG2g
	case proto.MIGProfile_MIG3g:
		return containerservice.GPUInstanceProfileMIG3g
	case proto.MIGProfile_MIG4g:
		return containerservice.GPUInstanceProfileMIG4g
	case proto.MIGProfile_MIG7g:
		return containerservice.GPUInstanceProfileMIG7g
	}
	return containerservice.GPUInstanceProfileMIG1g
}

func (a azureController) addNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	clusterName := req.ClusterName
	groupName := cred.ResourceGroup

	aksClient, err := getAKSClient(cred)
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	a.logger.Info(ctx, "fetching cluster information", "cluster", clusterName)
	clstr, err := aksClient.Get(ctx, groupName, clusterName)
	if err != nil {
		a.logger.Error(ctx, "failed to get cluster ", "error", err)
		return nil, err
	}

	apc, err := getAgentPoolClient(cred)
	if err != nil {
		a.logger.Error(ctx, "failed to get agent pool client", "error", err)
		return nil, err
	}
	nodeName := req.NodeSpec.Name

	a.logger.Info(ctx, "cluster found, adding new node", "cluster", clusterName, "node", nodeName)

	nodeTags := labels.GetNodeLabel(req.NodeSpec)

	count := int32(1)

	if req.NodeSpec.Count != 0 {
		count = int32(req.NodeSpec.Count)
	}

	instance := ""
	if req.NodeSpec.MachineType != "" {
		instance = common.GetInstance(constants.AzureLabel, req.NodeSpec.MachineType)
	} else {
		instance = req.NodeSpec.Instance
	}

	if instance == "" {
		return nil, errors.New("must provide valid instance by specifying MachineType or Instance.")
	}

	mcappp := containerservice.ManagedClusterAgentPoolProfileProperties{

		Count:      &count,
		VMSize:     &instance,
		NodeLabels: nodeTags,
		Tags:       nodeTags,
		Mode:       containerservice.AgentPoolModeUser,
		//	OrchestratorVersion: &constants.AzureKubeVersion,
		OsDiskSizeGB: &req.NodeSpec.DiskSize,
	}

	isGpu := common.IsGPU(req.NodeSpec.MachineType) || req.NodeSpec.GpuEnabled

	if isGpu && req.NodeSpec.MigProfile != proto.MIGProfile_UNKNOWN {
		mcappp.GpuInstanceProfile = getGPUProfile(req.NodeSpec.MigProfile) // containerservice.GPUInstanceProfileMIG1g
	}

	//Doc : https://docs.microsoft.com/en-us/rest/api/aks/agent-pools/create-or-update
	future, err := apc.CreateOrUpdate(
		ctx,
		groupName,
		*clstr.Name,
		nodeName,
		containerservice.AgentPool{ManagedClusterAgentPoolProfileProperties: &mcappp},
	)

	if err != nil {
		a.logger.Error(ctx, "failed to add node", "error", err)
		return nil, errors.Wrapf(err, "failed to add node to the cluster")
	}

	a.logger.Info(ctx, "requested to add new node, waiting on completion")
	err = future.WaitForCompletionRef(ctx, aksClient.Client)
	if err != nil {
		a.logger.Error(ctx, "failed to add node", "error", err)
		return nil, errors.Wrapf(err, "failed to add node to the cluster")
	}

	return &proto.NodeSpawnResponse{}, nil
}

func (a *azureController) deleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	groupName := cred.ResourceGroup
	apc, err := getAgentPoolClient(cred)
	if err != nil {
		a.logger.Error(ctx, "failed to get agent pool client", "error", err)
		return nil, err
	}

	cluster := req.GetClusterName()
	node := req.GetNodeGroupName()

	// Doc : https://docs.microsoft.com/en-us/rest/api/aks/agent-pools/delete
	future, err := apc.Delete(ctx, groupName, cluster, node)

	if err != nil {
		a.logger.Error(ctx, "failed to delete the node pool", "error", err)
		return nil, err
	}
	err = future.WaitForCompletionRef(ctx, apc.Client)

	if err != nil {
		a.logger.Error(ctx, "failed to delete the node pool", "error", err)
		return nil, err
	}

	if future.Response().StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("request resource '%s' not found in cluster '%s'", node, cluster)
	}

	a.logger.Info(ctx, "delete node successfully", "status", future.Response().Status)
	return &proto.NodeDeleteResponse{}, nil
}
