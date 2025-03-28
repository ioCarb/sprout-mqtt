package vm

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/machinefi/sprout/task"
	"github.com/machinefi/sprout/vm/proto"
)

func create(ctx context.Context, conn *grpc.ClientConn, projectID uint64, executeBinary, expParam string) error {
	cli := proto.NewVmRuntimeClient(conn)

	req := &proto.CreateRequest{
		ProjectID: projectID,
		Content:   executeBinary,
		ExpParam:  expParam,
	}
	if _, err := cli.Create(ctx, req); err != nil {
		return errors.Wrap(err, "failed to create vm instance")
	}
	return nil
}

func execute(ctx context.Context, conn *grpc.ClientConn, task *task.Task) ([]byte, error) {
	ds := []string{}
	for _, d := range task.Data {
		ds = append(ds, string(d))
	}
	req := &proto.ExecuteRequest{
		ProjectID:          task.ProjectID,
		TaskID:             task.ID,
		ClientID:           task.ClientID,
		SequencerSignature: task.Signature,
		Datas:              ds,
	}
	cli := proto.NewVmRuntimeClient(conn)
	resp, err := cli.ExecuteOperator(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute vm instance")
	}
	return resp.Result, nil
}
