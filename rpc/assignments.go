package rpc

import (
	"beaconchain/types"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/sirupsen/logrus"
)

var logassignments = logrus.New().WithField("module", "assignments")

func FnAssignments(epoch uint64) string {
	return fmt.Sprintf("/cache/%d.assign.gz", epoch)
}

func HasAssignments(epoch uint64) bool {
	file, err := os.Open(FnAssignments(epoch))
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

func NewAssignmentsFromPB(epoch uint64, src []*ethpb.ValidatorAssignments) *types.Assignments {
	since := time.Now()
	proposers := map[uint64]uint64{}
	slotsz := map[uint64]uint64{}
	firstSlot := uint64(0)

	numAssignments := 0
	// loop 1 - define sizes for allocations
	for ai := 0; ai < len(src); ai++ {
		numAssignments += len(src[ai].Assignments)
		log.Printf("%d assignments batch", len(src[ai].Assignments))
		for i := 0; i < len(src[ai].Assignments); i++ {
			assignment := src[ai].Assignments[i]

			if len(assignment.ProposerSlots) > 0 {
				slot := uint64(assignment.ProposerSlots[0])
				if slot < firstSlot || firstSlot == 0 {
					firstSlot = slot
				}
				proposer := uint64(assignment.ValidatorIndex)
				proposers[uint64(slot)] = proposer
				// log.Printf("%d ProposerSlots: %v %v", i, slot, proposer)
			}
			slot := uint64(assignment.AttesterSlot)
			if slot < firstSlot || firstSlot == 0 {
				firstSlot = slot
			}

			if val, ok := slotsz[slot]; ok {
				if val < uint64(assignment.CommitteeIndex) {
					slotsz[slot] = uint64(assignment.CommitteeIndex)
				}
			} else {
				slotsz[slot] = uint64(assignment.CommitteeIndex)
			}
		}
		// fmt.Printf("slotsz %v proposers: %v\n", slotsz, proposers)
	}
	// step 2 - allocation
	assignments := make([]types.AssignmentSlot, uint32(len(slotsz)))
	for slot, maxCommitteeIndex := range slotsz {
		slotIndex := uint64(slot) - firstSlot
		assignments[slotIndex] = types.AssignmentSlot{
			Proposer:   proposers[uint64(slot)],
			Committees: make([][]uint64, uint64(maxCommitteeIndex)+1),
		}
	}

	for ai := 0; ai < len(src); ai++ {
		for i := 0; i < len(src[ai].Assignments); i++ {
			assignment := src[ai].Assignments[i]
			slotIndex := uint64(assignment.AttesterSlot) - firstSlot
			m := make([]uint64, len(assignment.BeaconCommittees))
			for k := 0; k < len(assignment.BeaconCommittees); k++ {
				m[k] = uint64(assignment.BeaconCommittees[k])
			}
			assignments[slotIndex].Committees[assignment.CommitteeIndex] = m
		}
	}

	logassignments.Printf("encoding %v assignements from PB for epoch %v starting from slot %v took %v\n",
		numAssignments, epoch, firstSlot, time.Since(since))
	// log.Printf("max committee index: %v", assignments)

	return &types.Assignments{
		Epoch:          uint32(epoch),
		FirstSlot:      firstSlot,
		NumSlots:       uint32(len(proposers)),
		NumAssignments: uint64(numAssignments),
		Assignments:    assignments,
	}
}

func LoadAssignments(epoch uint64) (*types.Assignments, error) {
	file, err := os.Open(FnAssignments(epoch))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	if size <= 255 {
		return nil, errors.New("empty storage")
	}
	file, err = os.Open(FnAssignments(epoch))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	zr, err := gzip.NewReader(bufio.NewReader(file))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	bb2 := new(bytes.Buffer)
	_, err = io.Copy(bb2, zr)
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(bytes.NewReader(bb2.Bytes())) // Will read
	var out types.Assignments
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func SaveAssignments(epoch uint64, src *types.Assignments) error {
	// start := time.Now()

	if src == nil || epoch <= 0 {
		return nil
	}

	var bb bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&bb) // Will write to network.
	err := enc.Encode(src)
	if err != nil {
		return err
	}

	file, err := os.Create(FnAssignments(epoch))
	if err != nil {
		return err
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	if err := binary.Write(gz, binary.LittleEndian, bb.Bytes()); err != nil {
		return err
	}
	defer gz.Close()
	// logassignments.Printf("saving of epoch %v took %v", epoch, time.Since(start))
	return nil
}
