package grpc

import (
	grpc_network_listener "github.com/ZeraVision/go-zera-network/grpc/listener"
)

func InitialHookups() {
	validatorServer := grpc_network_listener.NewValidatorService()
	validatorServer.HandleBroadcast = Broadcast

	validatorServer.StartService()

}
