package proposal

import (
	"fmt"
	"log"

	"github.com/ZeraVision/ZeraBot/telegram"
	"github.com/ZeraVision/ZeraBot/txnstatus"
	"github.com/ZeraVision/ZeraBot/util"
	zera_protobuf "github.com/ZeraVision/go-zera-network/grpc/protobuf"
	"github.com/ZeraVision/zera-go-sdk/transcode"
)

// ProcessProposals processes new governance proposals and notifies subscribers
func ProcessProposals(block *zera_protobuf.Block) error {
	bot := telegram.GetBot()
	if bot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}

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

		// Extract the contract ID (symbol) from the proposal
		symbol := proposal.ContractId

		// Format the proposal message
		message := formatProposalMessage(proposal)

		// Notify subscribers
		if err := bot.NotifySubscribers(symbol, message); err != nil {
			log.Printf("Failed to notify subscribers for proposal %s: %v", proposal.Base.Hash, err)
		}
	}

	return nil
}

// formatProposalMessage formats a proposal into a user-friendly message
func formatProposalMessage(proposal *zera_protobuf.GovernanceProposal) string {

	proposalID := transcode.HexEncode(proposal.Base.Hash)

	return fmt.Sprintf(`🗳️ *New Proposal* 🗳️

*Symbol:* %s

*Title:* %s

*Synopsis:* %s

*Proposal ID:* %s

[View on Explorer](https://explorer.zera.vision/proposal/%s)`,
		proposal.ContractId,
		util.Truncate(proposal.Title, 50),
		util.Truncate(proposal.Synopsis, 200),
		proposalID,
		proposalID,
	)
}
