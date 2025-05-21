package grpc

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/time/rate"

	zera_protobuf "github.com/ZeraVision/go-zera-network/grpc/protobuf"
	"github.com/jfederk/ZeraBot/proposal"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	DB_ATTEMPTS_BEFORE_SATISFIED_DONE = 5
	DB_TIME_BETWEEN_ATTEMPTS          = 1 * time.Second
	DB_TOTAL_TIMEOUT                  = 1 * time.Minute
)

// Create a global rate limiter
var limiter = rate.NewLimiter(rate.Every(3*time.Second), 1) // 1 request per 3 seconds with a burst of 1

func Broadcast(ctx context.Context, block *zera_protobuf.Block) (*emptypb.Empty, error) {

	// Rate limit the broadcasts
	if !limiter.Allow() {
		log.Println("Broadcast rate limit exceeded (1 per 3 seconds), rejecting broadcast")
		return &emptypb.Empty{}, nil
	}

	if !isSenderFromDomain(ctx) {
		return &emptypb.Empty{}, nil
	}

	go func(block *zera_protobuf.Block) {
		proposal.ProcessProposals(block)
	}(block)

	return &emptypb.Empty{}, nil // awk

}

// isSenderFromDomain checks if the sender's IP matches a trusted source
func isSenderFromDomain(ctx context.Context) bool {

	if os.Getenv("ENVIRONMENT") == "development" {
		return true
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		log.Println("No peer information found in context")
		return false
	}

	// Get the sender's IP address
	senderIP, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		log.Printf("Failed to parse sender IP: %v", err)
		return false
	}

	// Compare the sender's IP against a trusted domain
	expectedDomain := os.Getenv("GRPC_ADDR")
	ips, err := net.LookupIP(expectedDomain)
	if err != nil {
		log.Printf("Failed to resolve domain %s: %v", expectedDomain, err)
		return false
	}

	for _, ip := range ips {
		if ip.String() == senderIP {
			return true
		}
	}

	log.Println("Sender IP does not match the domain" + expectedDomain + "(" + senderIP + ")")
	return false
}
