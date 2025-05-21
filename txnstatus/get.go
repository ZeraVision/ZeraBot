package txnstatus

import (
	"bytes"
	"fmt"

	zera_protobuf "github.com/ZeraVision/go-zera-network/grpc/protobuf"
	"github.com/ZeraVision/zera-go-sdk/transcode"
)

func GetStatus(txnHash []byte, status []*zera_protobuf.TXNStatusFees) (zera_protobuf.TXN_STATUS, error) {

	for _, txnStatus := range status {
		if bytes.Equal(txnStatus.TxnHash, txnHash) {
			return txnStatus.Status, nil
		}
	}

	return zera_protobuf.TXN_STATUS_FAULTY_TXN, fmt.Errorf("txnStatus not found for %s", transcode.HexEncode(txnHash))
}
