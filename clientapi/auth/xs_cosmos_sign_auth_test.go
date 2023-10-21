package auth

import (
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"strings"
	"testing"
)

func TestLoginTypeCosmos_Login(t1 *testing.T) {
	pub := "021f3c4a37f9ed0b0c7423c34f052dbc5c177544befdc7f812310853ce9a9e35bb"
	//pub := "024b191b7c990f7aeb6c142ed34bd3b43e1f1bbfe0ff6c75644ea275d1403b0be3"
	sig := "62c315f0b8b188b9b40b59bcf1e3e800e18f81c127a06bf31851afee81d311b03ed231ca92398dd031b56ebf6707c884df018c6da3c292f15cce2e873bb92f1200"
	//sig := "dcdaf444fe9118edb711f573ebebb25360c1cb9e1a68484298b7f0b536dc243671029e8a8c7e9da4f93244018e54f4be1b628414cdf97daecef6d35e4393896901"
	// cosmos
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		t1.Error("sig invalid")
	}
	pubKeyBytes, err := hex.DecodeString(pub)
	if err != nil {
		t1.Error("pub invalid")
	}
	pubKey := ethsecp256k1.PubKey{Key: pubKeyBytes}
	address := sdk.AccAddress(pubKey.Address())
	fmt.Println(address.String())
	addressHex := pubKey.Address().String()
	recoveredAddress := strings.ToLower(addressHex)
	t1.Logf("addr: %s", recoveredAddress)
	ok := pubKey.VerifySignature([]byte("1663753215"), sigBytes)
	fmt.Println("ok?", ok)
	// eth
	msgBytes := crypto.Keccak256([]byte("hello"))
	pubKey1, err := crypto.SigToPub(msgBytes, sigBytes)
	addressHex = crypto.PubkeyToAddress(*pubKey1).Hex()[2:]
	recoveredAddress = strings.ToLower(addressHex)

}
