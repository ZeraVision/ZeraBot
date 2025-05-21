package governance

import (
	"log"

	zera_protobuf "github.com/ZeraVision/go-zera-network/grpc/protobuf"
	"github.com/jfederk/ZeraBot/txnstatus"
)

func ProcessProposals(block *zera_protobuf.Block) {

	for _, proposal := range block.Transactions.GovernanceProposals {
		status, err := txnstatus.GetStatus(proposal.Base.Hash, block.Transactions.TxnFeesAndStatus)

		if err != nil {
			log.Printf("Error getting status for proposal %s: %v", proposal.Base.Hash, err)
			continue
		}

		// If not STATUS_OK - ignore
		if status != zera_protobuf.TXN_STATUS_OK {
			continue
		}

	}

}
