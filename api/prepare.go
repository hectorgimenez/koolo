package api

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/exec"
	"time"
)

var GRPCClient MapAssistApiClient

func StartAndConfigure(ctx context.Context) error {
	err := os.Chdir("MapAssist")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "KooloMA.exe")
	err = cmd.Start()
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

	go func() {
		<-ctx.Done()
		cmd.Process.Kill()
	}()

	return errors.New("error connecting mapassist")
}
