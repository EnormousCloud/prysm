package rpc

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var logbalances = logrus.New().WithField("module", "balances")

func FnBalances(epoch int64) string {
	return fmt.Sprintf("/cache/%d.balances.gz", epoch)
}

func HasBalances(epoch int64) bool {
	file, err := os.Open(FnBalances(epoch))
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

func LoadBalances(epoch int64) (map[uint64]uint64, error) {
	file, err := os.Open(FnBalances(epoch))
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
	file, err = os.Open(FnBalances(epoch))
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
	sz, err := io.Copy(bb2, zr)
	logbalances.Printf("balances of %d validators from cache of epoch %d\n", sz/8, epoch)
	ints := make([]uint64, sz/8)
	err = binary.Read(bytes.NewReader(bb2.Bytes()), binary.LittleEndian, ints)
	if err != nil {
		return nil, err
	}
	mapres := make(map[uint64]uint64, len(ints))
	for k, v := range ints {
		mapres[uint64(k)] = v
	}
	return mapres, nil
}

func SaveBalances(epoch int64, src map[uint64]uint64) error {
	if len(src) == 0 || epoch <= 0 {
		return nil
	}
	buf := make([]uint64, len(src))
	for k, v := range src {
		buf[k] = v
	}
	file, err := os.Create(FnBalances(epoch))
	if err != nil {
		return err
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	if err := binary.Write(gz, binary.LittleEndian, buf); err != nil {
		return err
	}
	return gz.Close()
}
