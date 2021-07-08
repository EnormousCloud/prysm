package rpc

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
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

func LoadAssignmentsPB(epoch uint64) (*ethpb.ValidatorAssignments, error) {
	start := time.Now()
	data, err := ioutil.ReadFile(FnAssignmentsPB(epoch))
	if err != nil {
		return nil, err
	}
	var message ethpb.ValidatorAssignments
	err = proto.Unmarshal(data, &message)
	if err != nil {
		return nil, err
	}
	logassignments.Printf("loading from PB took %v", time.Since(start))
	return &message, nil
}

func SaveAssignmentsPB(epoch uint64, src *ethpb.ValidatorAssignments) error {
	start := time.Now()
	if src == nil || epoch <= 0 {
		return nil
	}
	data, err := proto.Marshal(src)
	if err != nil {
		return fmt.Errorf("cannot marshal proto message to binary: %w", err)
	}
	err = ioutil.WriteFile(FnAssignmentsPB(epoch), data, 0666)
	if err != nil {
		return err
	}
	logassignments.Printf("saving PB took %v", time.Since(start))
	return err
}
