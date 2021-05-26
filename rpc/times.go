package rpc

import "time"

const cfgPageSize = 200000
const cfgSlotsPerEpoch = 32
const cfgSecondsPerSlot = 12
const cfgGenesisTimestamp = 1573489682

// EpochOfSlot will return the corresponding epoch of a slot
func EpochOfSlot(slot uint64) uint64 {
	return slot / cfgSlotsPerEpoch
}

// SlotToTime will return a time.Time to slot
func SlotToTime(slot uint64) time.Time {
	return time.Unix(int64(cfgGenesisTimestamp+slot*cfgSecondsPerSlot), 0)
}

// TimeToSlot will return time to slot in seconds
func TimeToSlot(timestamp uint64) uint64 {
	if cfgGenesisTimestamp > timestamp {
		return 0
	}
	return (timestamp - cfgGenesisTimestamp) / cfgSecondsPerSlot
}

// EpochToTime will return a time.Time for an epoch
func EpochToTime(epoch uint64) time.Time {
	return time.Unix(int64(cfgGenesisTimestamp+epoch*cfgSecondsPerSlot*cfgSlotsPerEpoch), 0)
}

// EpochToTime will return a time.Time for an epoch
func DayToTime(day uint64) time.Time {
	return time.Unix(int64(cfgGenesisTimestamp), 0).Add(time.Hour * time.Duration(24*int(day)))
}

// TimeToEpoch will return an epoch for a given time
func TimeToEpoch(ts time.Time) int64 {
	if int64(cfgGenesisTimestamp) > ts.Unix() {
		return 0
	}
	return (ts.Unix() - int64(cfgGenesisTimestamp)) / int64(cfgSecondsPerSlot) / int64(cfgSlotsPerEpoch)
}
