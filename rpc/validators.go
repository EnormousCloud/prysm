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
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var logvalidators = logrus.New().WithField("module", "validators")

func FnValidators(epoch uint64) string {
	return fmt.Sprintf("/cache/%d.validators.gz", epoch)
}

func HasValidators(epoch uint64) bool {
	file, err := os.Open(FnValidators(epoch))
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

func LoadValidators(epoch uint64) ([]types.ValidatorF, error) {
	start := time.Now()
	file, err := os.Open(FnValidators(epoch))
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
	file, err = os.Open(FnValidators(epoch))
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

	var out []types.ValidatorF
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	logvalidators.Infof("%d validators loaded from cache of epoch %d within %v", len(out), epoch, time.Since(start))
	return out, nil
}

func SaveValidators(epoch uint64, src []types.ValidatorF) error {
	if len(src) == 0 || epoch <= 0 {
		return nil
	}
	file, err := os.Create(FnValidators(epoch))
	if err != nil {
		return err
	}
	defer file.Close()

	var bb bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&bb) // Will write to network.
	err = enc.Encode(src)
	if err != nil {
		return err
	}

	gz := gzip.NewWriter(file)
	if err := binary.Write(gz, binary.LittleEndian, bb.Bytes()); err != nil {
		return err
	}
	defer gz.Close()
	return nil
}
