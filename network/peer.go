package network

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// DerivePeerID builds a transport PeerID distinct from NodeID (manifest rule).
// Format: peer:<base64url(sha256(nodeID|seed)[:16])>
func DerivePeerID(nodeID, seed string) string {
	h := sha256.Sum256([]byte(nodeID + "|" + seed))
	return fmt.Sprintf("peer:%s", base64.RawURLEncoding.EncodeToString(h[:16]))
}
