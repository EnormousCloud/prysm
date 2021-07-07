package rpc

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

func FnAssignmentsPB(epoch uint64) string {
	return fmt.Sprintf("/cache/%d.assign.pb", epoch)
}

func HasAssignmentsPB(epoch uint64) bool {
	file, err := os.Open(FnAssignmentsPB(epoch))
	if err != nil {
		return false
	}
	defer file.Close()
	stats, statsErr := file.Stat()
	if statsErr != nil {
		return false
	}
	if stats.Size() < 100 {
		return false
	}
	return true
}

func SaveAssignmentsPB(epoch uint64, src *ethpb.ValidatorAssignments) error {
	if src == nil || epoch <= 0 {
		return nil
	}

	var bb bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&bb) // Will write to network.
	err := enc.Encode(*src)
	if err != nil {
		return err
	}

	file, err := os.Create(FnAssignmentsPB(epoch))
	if err != nil {
		return err
	}
	defer file.Close()

	return binary.Write(file, binary.LittleEndian, bb.Bytes())
}
