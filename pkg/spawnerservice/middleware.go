package spawnerservice

import (
	"context"

	pb "gitlab.com/netbook-devs/spawner-service/pb"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(ClusterController) ClusterController

// LoggingMiddleware takes a logger as a dependency
// and returns a service Middleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next ClusterController) ClusterController {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   ClusterController
}

func (mw loggingMiddleware) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (res *pb.ClusterResponse, err error) {
	defer func() {
		mw.logger.Log("method", "CreateCluster", "provider", req.Provider, "region", req.Region, "node", req.Node, "labels", req.Labels, "res", res, "err", err)
	}()
	return mw.next.CreateCluster(ctx, req)
}

func (mw loggingMiddleware) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (res *pb.ClusterDeleteResponse, err error) {
	defer func() {
		mw.logger.Log("method", "DeleteCluster", "name", req.ClusterName, "res", res, "err", err)
	}()
	return mw.next.DeleteCluster(ctx, req)
}

func (mw loggingMiddleware) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (res *pb.ClusterStatusResponse, err error) {
	defer func() {
		mw.logger.Log("method", "ClusterStatus", "name", req.ClusterName, "res", res, "err", err)
	}()
	return mw.next.ClusterStatus(ctx, req)
}

func (mw loggingMiddleware) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (res *pb.NodeSpawnResponse, err error) {
	defer func() {
		mw.logger.Log("method", "AddNode", "name", req.ClusterName, "nodespecinstance", req.NodeSpec.Instance, "res", res, "err", err)
	}()
	return mw.next.AddNode(ctx, req)
}

func (mw loggingMiddleware) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (res *pb.NodeDeleteResponse, err error) {
	defer func() {
		mw.logger.Log("method", "DeleteNode", "name", req.ClusterName, "nodegroupname", req.NodeGroupName, "res", res, "err", err)
	}()
	return mw.next.DeleteNode(ctx, req)
}

func (mw loggingMiddleware) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (res *pb.CreateVolumeResponse, err error) {
	defer func() {
		mw.logger.Log("method", "CreateVolume", "volumetype", req.Volumetype, "size", req.Size, "region", req.Region, "provider", req.Provider, "snapshotid", req.Snapshotid, "res", res, "err", err)
	}()
	return mw.next.CreateVolume(ctx, req)
}

func (mw loggingMiddleware) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (res *pb.DeleteVolumeResponse, err error) {
	defer func() {
		mw.logger.Log("method", "DeleteVolume", "volumeid", req.Volumeid, "region", req.Region, "provider", req.Provider, "res", res, "err", err)
	}()
	return mw.next.DeleteVolume(ctx, req)
}

func (mw loggingMiddleware) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (res *pb.CreateSnapshotResponse, err error) {
	defer func() {
		mw.logger.Log("method", "CreateSnapshot", "volumeid", req.Volumeid, "region", req.Region, "provider", req.Provider, "res", res, "err", err)
	}()
	return mw.next.CreateSnapshot(ctx, req)
}

func (mw loggingMiddleware) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (res *pb.CreateSnapshotAndDeleteResponse, err error) {
	defer func() {
		mw.logger.Log("method", "CreateSnapshotAndDelete", "volumeid", req.Volumeid, "region", req.Region, "provider", req.Provider, "res", res, "err", err)
	}()
	return mw.next.CreateSnapshotAndDelete(ctx, req)
}

// InstrumentingMiddleware returns a service middleware that instruments
// the number of integers summed and characters concatenated over the lifetime of
// the service.
func InstrumentingMiddleware(ints metrics.Counter) Middleware {
	return func(next ClusterController) ClusterController {
		return instrumentingMiddleware{ints, next}
	}
}

type instrumentingMiddleware struct {
	ints metrics.Counter
	next ClusterController
}

func (mw instrumentingMiddleware) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	v, err := mw.next.CreateCluster(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	v, err := mw.next.ClusterStatus(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
	v, err := mw.next.DeleteCluster(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {
	v, err := mw.next.AddNode(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	v, err := mw.next.DeleteNode(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	v, err := mw.next.CreateVolume(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	v, err := mw.next.DeleteVolume(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	v, err := mw.next.CreateSnapshot(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	v, err := mw.next.CreateSnapshotAndDelete(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}
