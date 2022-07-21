package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	clusterName = "gcp-cluster-test-3"
	region      = "us-west-2" //"us-central1" //"eastus2"
	provider    = "aws"
	nodeName    = "add-node-2"
	instance    = "e2-medium"
	volumeName  = "vol-50-20220607164454"
	accountName = "netbook-aws"
)

func main() {

	var logger, _ = zap.NewDevelopment()
	var sugar = logger.Sugar()
	defer sugar.Sync()

	fs := flag.NewFlagSet("testclient", flag.ExitOnError)
	grpcAddr := fs.String("grpc-addr", ":8083", "gRPC address of spawner")
	method := fs.String("method", "HealthCheck", "default HealthCheck")
	fs.Usage = usageFor(fs, os.Args[0]+" [flags] <a> <b>")
	fs.Parse(os.Args[1:])

	if *grpcAddr == "" {
		sugar.Errorf("host address is empty '%s'", *grpcAddr)
		return
	}
	conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
	if err != nil {
		sugar.Errorw("error connecting to remote", "error", err)
		return
	}
	defer conn.Close()
	client := proto.NewSpawnerServiceClient(conn)

	if err != nil {
		sugar.Errorw("error connecting to remote", "error", err)
		return
	}

	node := &proto.NodeSpec{
		Name:     nodeName,
		Instance: instance,
		DiskSize: 30,
	}
	createClusterReq := &proto.ClusterRequest{
		Provider: provider,
		Region:   region,
		Node:     node,
		Labels: map[string]string{
			"user": "dev-tester",
		},
		ClusterName: clusterName,
		AccountName: accountName,
	}

	addTokenReq := &proto.AddTokenRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
	}

	getTokenReq := &proto.GetTokenRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}

	addRoute53RecordReq := &proto.AddRoute53RecordRequest{
		DnsName:    "20.85.85.202",
		RecordName: "*.1117907260.eastus2.azure.app.dev.netbook.ai",
		// Region:      region,
		Provider:    provider,
		AccountName: accountName,
		// RegionIdentifier: "Oregon region",
	}

	getRoute53Req := &proto.GetRoute53TXTRecordsRequest{}

	deleteRoute53Req := &proto.DeleteRoute53RecordsRequest{
		Records: []*proto.Route53ResourceRecordSet{
			{
				Type: "TXT",
				Name: "ash1234.app.dev.netbook.ai",
				ResourceRecords: []*proto.Route53ResourceRecord{
					{
						Value: "test1",
					},
					{
						Value: "test2",
					},
				},
				TtlInSeconds: 250,
			},
		},
	}

	createRoute53RecordsReq := &proto.CreateRoute53RecordsRequest{
		Records: []*proto.Route53ResourceRecordSet{
			{
				Type: "TXT",
				Name: "ash1234.app.dev.netbook.ai",
				ResourceRecords: []*proto.Route53ResourceRecord{
					{
						Value: "test1",
					},
					{
						Value: "test2",
					},
				},
				TtlInSeconds: 250,
			},
		},
	}

	clusterStatusReq := &proto.ClusterStatusRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}

	getClustersReq := &proto.GetClustersRequest{
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}

	getClusterReq := &proto.GetClusterRequest{
		ClusterName: clusterName,
		Provider:    provider,
		Region:      region,
		AccountName: accountName,
	}

	addNode := &proto.NodeSpec{
		Name:          nodeName,
		Count:         5,
		Instance:      instance,
		MigProfile:    proto.MIGProfile_MIG3g,
		CapacityType:  proto.CapacityType_ONDEMAND,
		MachineType:   "m",
		SpotInstances: []string{"t2.small", "t3.small"},
		DiskSize:      20,
		GpuEnabled:    false,
		Labels: map[string]string{"cluster-name": clusterName,
			"node-name":   nodeName,
			"user":        "dev-tester",
			"workspaceid": "dev-tester",
		},
	}

	addNodeReq := &proto.NodeSpawnRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
		NodeSpec:    addNode,
		AccountName: accountName,
	}

	deleteClusterReq := &proto.ClusterDeleteRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
		ForceDelete: true,
	}

	deleteNodeReq := &proto.NodeDeleteRequest{
		ClusterName:   clusterName,
		NodeGroupName: nodeName,
		Region:        region,
		Provider:      provider,
		AccountName:   accountName,
	}

	createVolumeReq := &proto.CreateVolumeRequest{
		Availabilityzone: region,
		Volumetype:       "StandardSSD_LRS", //"gp2",
		Size:             50,
		Snapshotid:       "vol-50-20220607152827-snapshot",
		//SnapshotUri: "snapshot-uri",
		Region:         region,
		Provider:       provider,
		AccountName:    accountName,
		DeleteSnapshot: true,
	}

	deleteVolumeReq := &proto.DeleteVolumeRequest{
		//		Volumeid:    "vol-eastus2-1-20220323121600",
		Volumeid:    volumeName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}

	createSnapshotReq := &proto.CreateSnapshotRequest{
		Volumeid:    volumeName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}
	createSnapshotAndDeleteReq := &proto.CreateSnapshotAndDeleteRequest{
		Volumeid:    volumeName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}

	getWorkspacesCost := &proto.GetWorkspacesCostRequest{
		WorkspaceIds: []string{"d1411352-c14a-4a78-a1d6-44d4c199ba3a", "18638c97-7352-426e-a79e-241956188fed", "dceaf501-1775-4339-ba7b-ec6d98569d11"},
		Provider:     "aws",
		AccountName:  "netbook-aws-dev",
		StartDate:    "2022-04-01",
		EndDate:      "2022-05-01",
		Granularity:  "DAILY",
		CostType:     "BlendedCost",
		GroupBy: &proto.GroupBy{
			Type: "TAG",
			Key:  "workspaceid",
		},
	}

	getApplicationsCost := &proto.GetApplicationsCostRequest{
		ApplicationIds: []string{"d1411352-c14a-4a78-a1d6-44d4c199ba3a", "18638c97-7352-426e-a79e-241956188fed", "dceaf501-1775-4339-ba7b-ec6d98569d11"},
		Provider:       "aws",
		AccountName:    "netbook-aws-dev",
		StartDate:      "2022-04-01",
		EndDate:        "2022-05-01",
		Granularity:    "DAILY",
		CostType:       "BlendedCost",
		GroupBy: &proto.GroupBy{
			Type: "TAG",
			Key:  "workspaceid",
		},
	}

	//AWS cost request
	getCostByTime := &proto.GetCostByTimeRequest{
		Ids:         []string{"d1411352-c14a-4a78-a1d6-44d4c199ba3a", "18638c97-7352-426e-a79e-241956188fed", "dceaf501-1775-4339-ba7b-ec6d98569d11"},
		Provider:    "aws",
		AccountName: "netbook-aws-dev",
		StartDate:   "2022-04-01",
		EndDate:     "2022-05-01",
		Granularity: "DAILY",
		GroupBy: &proto.GroupBy{
			Type: "TAG",
			Key:  "workspaceid",
		},
	}

	//Azure Cost Req
	// getCostByTime := &proto.GetCostByTimeRequest{
	// 	Ids:         []string{"24522d72-9b86-48c4-b66a-521a2f202413", "testid", "5d4eb7d8-9289-4740-a7f8-a9bfbdf06a16", "b5fbc7b6-e502-4093-81aa-d3efdce80afc"},
	// 	Provider:    "azure",
	// 	AccountName: "netbook-azure-dev",
	// 	StartDate:   "2022-04-01",
	// 	EndDate:     "2022-05-17",
	// 	Granularity: "DAILY",
	// 	GroupBy: &proto.GroupBy{
	// 		Type: "TAG",
	// 		Key:  "workspaceid",
	// 	},
	// }

	switch *method {
	case "Echo":
		v, err := client.Echo(context.Background(), &proto.EchoRequest{Msg: "hello spawner"})

		if err != nil {
			sugar.Errorw("Echo", "error", err)
			return
		}
		sugar.Infow("Echo", "response", v)

	case "HealthCheck":
		v, err := client.HealthCheck(context.Background(), &proto.Empty{})

		if err != nil {
			sugar.Errorw("HealthCheck", "error", err)
			return
		}
		sugar.Infow("HealthCheck", "response", v)

	case "CreateCluster":
		v, err := client.CreateCluster(context.Background(), createClusterReq)
		if err != nil {
			sugar.Errorw("error creating cluster", "error", err)
			return
		}
		sugar.Infow("CreateCluster method", "response", v)
	case "AddToken":
		v, err := client.AddToken(context.Background(), addTokenReq)
		if err != nil {
			sugar.Errorw("error adding token", "error", err)
			return
		}
		sugar.Infow("AddToken method", "reponse", v)
	case "GetToken":
		v, err := client.GetToken(context.Background(), getTokenReq)
		if err != nil {
			sugar.Errorw("error getting token", "error", err)
			return
		}
		base64Ca := base64.StdEncoding.EncodeToString([]byte(v.CaData))
		sugar.Infow("base64 token", "Ca", base64Ca)
		sugar.Infow("GetToken method", "response", v)
	case "AddRoute53Record":
		v, err := client.AddRoute53Record(context.Background(), addRoute53RecordReq)
		if err != nil {
			sugar.Errorw("error creating Alias record", "error", err)
			return
		}
		sugar.Infow("AddRoute53Record method", "response", v)
	case "GetCluster":
		v, err := client.GetCluster(context.Background(), getClusterReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		sugar.Infow("GetCluster method", "response", v)
	case "GetClusters":
		v, err := client.GetClusters(context.Background(), getClustersReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		sugar.Infow("GetClusters method", "response", v)
	case "ClusterStatus":
		ctx := context.Background()

		ctx = metadata.AppendToOutgoingContext(ctx, "trace-id", "cafebabe-345678-xcvbn-345678-QWDFVBNJI")
		v, err := client.ClusterStatus(ctx, clusterStatusReq)
		if err != nil {
			sugar.Errorw("error fetching cluster status", "error", err)
			return
		}
		sugar.Infow("ClusterStatus method", "response", v)
	case "AddNode":
		v, err := client.AddNode(context.Background(), addNodeReq)
		if err != nil {
			sugar.Errorw("error adding node", "error", err)
			return
		}
		sugar.Infow("AddNode method", "response", v)
	case "DeleteCluster":
		v, err := client.DeleteCluster(context.Background(), deleteClusterReq)
		if err != nil {
			sugar.Errorw("error deleting cluster", "error", err)
			return
		}
		sugar.Infow("DeleteCluster method", "response", v)
	case "DeleteAllClustersInRegion":
		v, err := client.GetClusters(context.Background(), getClustersReq)
		if err != nil {
			sugar.Errorw("error getting clusters", "account", getClustersReq.AccountName, "provider", getClustersReq.Provider, "region", getClustersReq.Region, "error", err)
			return
		}
		clusters := make([]string, 0)
		for _, cluster := range v.Clusters {
			clusters = append(clusters, cluster.Name)
		}
		logger.Sugar().Infow("deleting following clusters", "clusters", clusters)

		for _, cluster := range clusters {
			sugar.Infow("deleting cluster", "cluster", cluster)
			req := &proto.ClusterDeleteRequest{
				Provider:    provider,
				Region:      region,
				AccountName: accountName,
				ClusterName: cluster,
				ForceDelete: true,
			}
			v, err := client.DeleteCluster(context.Background(), req)
			if err != nil {
				sugar.Errorw("error deleting cluster", "cluster", cluster, "error", err)
			}
			sugar.Infow("DeleteCluster method", "cluster", cluster, "response", v)
		}

		sugar.Infow("Deleted all clusters", "account", getClustersReq.AccountName, "provider", getClustersReq.Provider, "region", getClustersReq.Region)
	case "DeleteNode":
		v, err := client.DeleteNode(context.Background(), deleteNodeReq)
		if err != nil {
			sugar.Errorw("error deleting node", "error", err)
			return
		}
		sugar.Infow("DeleteNode method", "response", v)

	case "CreateVolume":
		v, err := client.CreateVolume(context.Background(), createVolumeReq)
		if err != nil {
			sugar.Errorw("error creating volume", "error", err)
			return
		}
		sugar.Infow("CreateVolume method", "response", v)

	case "DeleteVolume":
		v, err := client.DeleteVolume(context.Background(), deleteVolumeReq)
		if err != nil {
			sugar.Errorw("error deleting volume", "error", err)
			return
		}
		sugar.Infow("DeleteVolume method", "response", v)

	case "CreateSnapshot":
		v, err := client.CreateSnapshot(context.Background(), createSnapshotReq)
		if err != nil {
			sugar.Errorw("error creating snapshot", "error", err)
			return
		}
		sugar.Infow("CreateSnapshot method", "response", v)

	case "DeleteSnapshot":
		v, err := client.DeleteSnapshot(context.Background(), &proto.DeleteSnapshotRequest{
			Provider:    provider,
			Region:      region,
			AccountName: accountName,
			SnapshotId:  fmt.Sprintf("%s-snapshot", volumeName),
		})

		if err != nil {
			sugar.Errorw("error deleting snapshot", "error", err)
			return
		}
		sugar.Info("snapshot deleted", "response", v)

	case "CreateSnapshotAndDelete":
		v, err := client.CreateSnapshotAndDelete(context.Background(), createSnapshotAndDeleteReq)
		if err != nil {
			sugar.Errorw("error creating snapshot and deleting volume", "error", err)
			return
		}
		sugar.Infow("CreateSnapshotAndDelete method", "response", v)

	case "RegisterWithRancher":
		v, err := client.RegisterWithRancher(context.Background(), &proto.RancherRegistrationRequest{
			ClusterName: clusterName,
		})
		if err != nil {
			sugar.Errorw("error registering cluster with rancher", "error", err)
			return
		}
		sugar.Infow("RegisterWithRancher method", "response", v)
	case "GetWorkspacesCost":
		v, err := client.GetWorkspacesCost(context.Background(), getWorkspacesCost)
		if err != nil {
			sugar.Errorw("error getting workspaces cost", "error", err)
			return
		}
		sugar.Infow("GetWorkspaceCost method", "response", v)
	case "GetApplicationsCost":
		v, err := client.GetApplicationsCost(context.Background(), getApplicationsCost)
		if err != nil {
			sugar.Errorw("error getting applications cost", "error", err)
			return
		}
		sugar.Infow("GetApplicationsCost method", "response", v)
	case "ReadCredentialAws":
		v, err := client.ReadCredential(context.Background(), &proto.ReadCredentialRequest{
			Account: "alexis",
			Type:    "aws",
		})
		if err != nil {
			sugar.Errorw("error reading credentials", "error", err)
			return
		}
		sugar.Infow("ReadCredential", "response", v)

	case "WriteCredentialAws":
		v, err := client.WriteCredential(context.Background(), &proto.WriteCredentialRequest{
			Account: "alexis",
			Type:    "aws",
			Cred: &proto.WriteCredentialRequest_AwsCred{
				AwsCred: &proto.AwsCredentials{
					AccessKeyID:     "access_id",
					SecretAccessKey: "secret_key",
					Token:           "token",
				},
			},
		})
		if err != nil {
			sugar.Errorw("error writing credentials", "error", err)
			return
		}
		sugar.Infow("WriteCredentialAws", "response", v)
	case "ReadCredentialAzure":
		v, err := client.ReadCredential(context.Background(), &proto.ReadCredentialRequest{
			Account: "netbook-azure-dev",
			Type:    "azure",
		})
		if err != nil {
			sugar.Errorw("error reading credentials", "error", err)
			return
		}
		sugar.Infow("ReadCredential", "response", v)

	case "WriteCredentialAzure":
		v, err := client.WriteCredential(context.Background(), &proto.WriteCredentialRequest{
			Account: "alex",
			Type:    "azure",
			Cred: &proto.WriteCredentialRequest_AzureCred{
				AzureCred: &proto.AzureCredentials{
					SubscriptionID: "subscription",
					TenantID:       "tenant_id",
					ClientID:       "client_id",
					ClientSecret:   "client_secret",
					ResourceGroup:  "resource_group",
				},
			},
		})
		if err != nil {
			sugar.Errorw("error writing credentials", "error", err)
			return
		}
		sugar.Infow("WriteCredentialAws", "response", v)

	case "ReadCredentialGitPAT":
		v, err := client.ReadCredential(context.Background(), &proto.ReadCredentialRequest{
			Account: "nsp-dev",
			Type:    "git-pat",
		})

		if err != nil {
			sugar.Errorw("error reading Git PAT ", err)
			return
		}
		sugar.Infow("ReadCredentialResponse_GitPat", "response", v)
	case "WriteCredentialGitPAT":
		v, err := client.WriteCredential(context.Background(), &proto.WriteCredentialRequest{
			Account: "nsp-dev",
			Type:    "git-pat",
			Cred: &proto.WriteCredentialRequest_GitPat{
				GitPat: &proto.GithubPersonalAccessToken{
					Token: "this-is-very-secret-token-thats-why-you-see-this-message-when-reading",
				},
			},
		})

		if err != nil {
			sugar.Errorw("error writing Git PAT ", err)
			return
		}
		sugar.Infow("WriteCredentialResponse_GitPat", "response", v)
	case "AddTag":
		v, err := client.TagNodeInstance(context.Background(), &proto.TagNodeInstanceRequest{
			Provider:    provider,
			Region:      region,
			AccountName: accountName,
			ClusterName: clusterName,
			Labels: map[string]string{
				"label1": "valuelabel1",
			},
		})

		if err != nil {
			sugar.Errorw("error adding tags to node", "error", err)
		}
		sugar.Infow("TagNodeInstane", "response", v)

	case "GetCostByTime":
		v, err := client.GetCostByTime(context.Background(), getCostByTime)
		if err != nil {
			sugar.Errorw("error getting cost by time", "error", err)
			return
		}
		sugar.Infow("GetWorkspaceCost method", "response", v)
	case "GetContainerRegistryAuth":
		v, err := client.GetContainerRegistryAuth(context.Background(), &proto.GetContainerRegistryAuthRequest{
			Provider:    "aws",
			Region:      "us-west-2",
			AccountName: accountName,
		})
		if err != nil {
			sugar.Errorw("error getting ecr auth details", "error", err)
			return
		}
		sugar.Infow("GetContainerRegistryAuth method", "response", v)
	case "CreateContainerRegistryRepo":
		v, err := client.CreateContainerRegistryRepo(context.Background(), &proto.CreateContainerRegistryRepoRequest{
			Provider:    "aws",
			Region:      "us-west-2",
			AccountName: accountName,
			Name:        "nsp-test-2",
		})
		if err != nil {
			sugar.Errorw("error creating container repo", "error", err)
			return
		}
		sugar.Infow("CreateContainerRegistryRepo: created repo", "response", v)

	case "RegisterClusterOIDC":
		v, err := client.RegisterClusterOIDC(context.Background(), &proto.RegisterClusterOIDCRequest{
			Provider:    "aws",
			Region:      region,
			AccountName: accountName,
			ClusterName: clusterName,
		})

		if err != nil {
			sugar.Errorw("error connecting cluster oidc to policy", "error", err)
			return
		}
		sugar.Infow("ConnectClusterOIDCToTrustPolicy : connect cluster to oidc", "response", v)

	case "GetRoute53TXTRecords":
		v, err := client.GetRoute53TXTRecords(context.Background(), getRoute53Req)
		if err != nil {
			sugar.Errorw("error getting route53 records", "error", err)
			return
		}
		sugar.Infow("Route53 Records", "response", v)

	case "DeleteRoute53Records":
		v, err := client.DeleteRoute53Records(context.Background(), deleteRoute53Req)
		if err != nil {
			sugar.Errorw("error deleting route53 records", "error", err)
			return
		}
		sugar.Infow("Route53 records deleted successfully", "response", v)

	case "CreateRoute53Records":
		v, err := client.CreateRoute53Records(context.Background(), createRoute53RecordsReq)
		if err != nil {
			sugar.Errorw("error appending route53 records", "error", err)
			return
		}
		sugar.Infow("Route53 records created successfully", "response", v)
	case "CopySnapshot":
		v, err := client.CopySnapshot(context.Background(), &proto.CopySnapshotRequest{
			Provider:    provider,
			Region:      region,
			AccountName: accountName,
			SnapshotId:  "snap-001c9501528bc1a33",
		})
		if err != nil {
			sugar.Errorw("error copying snapshot", "error", err)
			return
		}
		sugar.Infow("snapshot copied", "response", v)

	case "PresignS3":
		v, err := client.PresignS3Url(context.Background(), &proto.PresignS3UrlRequest{
			Region:      "us-west-2",
			AccountName: accountName,
			Bucket:      "nishanth-test",
			File:        "/hello.txt",
		})
		if err != nil {
			sugar.Errorw("erro presigning s3 url", "error", err)
			return
		}
		sugar.Infow("successfully signed s3", "signed-url", v.SignedUrl)
	default:
		sugar.Errorw("error: invalid method", "method", *method)
		return
	}
}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")

		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
