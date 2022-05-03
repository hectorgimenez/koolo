package api

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var GRPCClient MapAssistApiClient

type MapAssistClient struct {
	logger *zap.Logger
	cmd    *exec.Cmd
}

func NewMapAssistClient(logger *zap.Logger) *MapAssistClient {
	return &MapAssistClient{logger: logger}
}

func (ma *MapAssistClient) StartAndConfigure(ctx context.Context) error {
	err := os.Chdir("MapAssist")
	if err != nil {
		return err
	}

	ma.cmd = exec.CommandContext(ctx, "KooloMA.exe")
	err = ma.cmd.Start()
	if err != nil {
		return err
	}

	err = os.Chdir("../")
	if err != nil {
		return err
	}

	grpcClient, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("error dialing MapAssist: %w", err)
	}
	GRPCClient = NewMapAssistApiClient(grpcClient)

	for i := 0; i < 10; i++ {
		_, err = GRPCClient.GetData(ctx, &R{})
		if err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}

	return errors.New("error connecting mapassist")
}

// Stop ensures MapAssist is closed and if not it will try to force shutdown/kill
func (ma *MapAssistClient) Stop() {
	ma.logger.Info("Closing MapAssist...")
	if ma.cmd.ProcessState == nil || ma.cmd.ProcessState.Exited() || ma.cmd.ProcessState.Success() {
		return
	}

	err := ma.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		err = ma.cmd.Process.Kill()
		if err != nil {
			ma.logger.Error("Error closing MapAssist", zap.Error(err))
			return
		}
	}
	ma.logger.Debug("MapAssist closed successfully")
}
