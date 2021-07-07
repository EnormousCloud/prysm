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
	"strconv"
	"strings"
	"time"
)

func FnAssignments(epoch uint64) string {
	return fmt.Sprintf("/cache/%d.assign", epoch)
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

type AssignmentSlot struct {
	Proposer  uint64
	Commitees [][]uint64
}

type Assignments struct {
	// Number of epoch
	Epoch uint32
	// typically 32
	NumSlots uint32
	// First slot in the epoch
	FirstSlot   uint64
	Assignments []AssignmentSlot
}

func (a *Assignments) ToMaps() *types.EpochAssignments {
	proposers := map[uint64]uint64{}
	attestors := map[string]uint64{}
	for slotIndex, assignment := range a.Assignments {
		proposers[uint64(slotIndex)+a.FirstSlot] = assignment.Proposer
		for ci, members := range assignment.Commitees {
			for mi, member := range members {
				key := fmt.Sprintf("%d-%d-%d", uint64(slotIndex)+a.FirstSlot, ci, mi)
				attestors[key] = member
			}
		}
	}
	return &types.EpochAssignments{
		ProposerAssignments: proposers,
		AttestorAssignments: attestors,
	}
}

func unpackAssignmentKey(str string) (AttesterSlot uint64, CommitteeIndex uint64, MemberIndex uint64) {
	x := strings.Split(str, "-")
	AttesterSlot, _ = strconv.ParseUint(x[0], 10, 64)
	CommitteeIndex, _ = strconv.ParseUint(x[1], 10, 64)
	MemberIndex, _ = strconv.ParseUint(x[2], 10, 64)
	return
}

func NewAssignmentsFromMaps(epoch uint64, src *types.EpochAssignments) Assignments {
	since := time.Now()
	firstSlot := uint64(0)
	for k, _ := range src.ProposerAssignments {
		if firstSlot == 0 || k < firstSlot {
			firstSlot = k
		}
	}
	assignments := make([]AssignmentSlot, len(src.ProposerAssignments), len(src.ProposerAssignments))
	for slot, proposer := range src.ProposerAssignments {
		slotIndex := slot - firstSlot
		slotStart := time.Now()

		numCommitee := uint64(0)
		for key, _ := range src.AttestorAssignments {
			AttesterSlot, CommiteeIndex, _ := unpackAssignmentKey(key)
			if AttesterSlot == slot && numCommitee < CommiteeIndex+1 {
				numCommitee = CommiteeIndex + 1
			}
		}
		// size of commitee was defined
		commitees := make([][]uint64, numCommitee)
		for ci := uint64(0); ci < numCommitee; ci++ {
			// log.Printf("Encoding assignments slot %v commitee %v out of %v", slot, ci, numCommitee)
			numMembers := uint64(0)
			for key, _ := range src.AttestorAssignments {
				AttesterSlot, CommiteeIndex, MemberIndex := unpackAssignmentKey(key)
				if AttesterSlot == slot && CommiteeIndex == ci && numMembers < MemberIndex+1 {
					numMembers = MemberIndex + 1
				}
			}
			members := make([]uint64, numMembers)
			for key, member := range src.AttestorAssignments {
				AttesterSlot, CommiteeIndex, MemberIndex := unpackAssignmentKey(key)
				if AttesterSlot == slot && CommiteeIndex == ci {
					members[MemberIndex] = member
				}
			}
			commitees[ci] = members
		}

		assignments[slotIndex] = AssignmentSlot{
			Proposer:  proposer,
			Commitees: commitees,
		}

		log.Printf("Encoding assignments slot %v took %v", slot, time.Since(slotStart))
	}

	log.Printf("Encoding assignments for epoch %v took %v", epoch, time.Since(since))
	return Assignments{
		Epoch:       uint32(epoch),
		FirstSlot:   firstSlot,
		NumSlots:    uint32(len(src.ProposerAssignments)),
		Assignments: assignments,
	}
}

func LoadAssignmentsMaps(epoch uint64) (*types.EpochAssignments, error) {
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
	var out Assignments
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out.ToMaps(), nil
}

func SaveAssignmentsMaps(epoch uint64, src *types.EpochAssignments) error {
	if src == nil || epoch <= 0 {
		return nil
	}

	var bb bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&bb) // Will write to network.
	err := enc.Encode(NewAssignmentsFromMaps(epoch, src))
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
	return gz.Close()
}
