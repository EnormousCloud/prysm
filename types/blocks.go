package types

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

// ChainHead is a struct to hold chain head data
type ChainHead struct {
	HeadSlot                   uint64
	HeadEpoch                  uint64
	HeadBlockRoot              []byte
	FinalizedSlot              uint64
	FinalizedEpoch             uint64
	FinalizedBlockRoot         []byte
	JustifiedSlot              uint64
	JustifiedEpoch             uint64
	JustifiedBlockRoot         []byte
	PreviousJustifiedSlot      uint64
	PreviousJustifiedEpoch     uint64
	PreviousJustifiedBlockRoot []byte
}

// EpochData is a struct to hold epoch data
type EpochData struct {
	Epoch                   uint64
	Validators              []*Validator
	ValidatorAssignments    *Assignments
	Blocks                  map[uint64]map[string]*Block
	EpochParticipationStats *ValidatorParticipation
}

// ValidatorParticipation is a struct to hold validator participation data
type ValidatorParticipation struct {
	Epoch                   uint64
	Finalized               bool
	GlobalParticipationRate float32
	VotedEther              uint64
	EligibleEther           uint64
}

// BeaconCommitteItem is a struct to hold beacon committee data
type BeaconCommitteItem struct {
	ValidatorIndices []uint64
}

// Validator is a struct to hold validator data
type Validator struct {
	Index                      uint64 `db:"validatorindex"`
	PublicKey                  []byte `db:"pubkey"`
	Balance                    uint64 `db:"balance"`
	EffectiveBalance           uint64 `db:"effectivebalance"`
	Slashed                    bool   `db:"slashed"`
	ActivationEligibilityEpoch uint64 `db:"activationeligibilityepoch"`
	ActivationEpoch            uint64 `db:"activationepoch"`
	ExitEpoch                  uint64 `db:"exitepoch"`
	WithdrawableEpoch          uint64 `db:"withdrawableepoch"`
	WithdrawalCredentials      []byte `db:"withdrawalcredentials"`

	BalanceActivation uint64 `db:"balanceactivation"`
	Balance1d         uint64 `db:"balance1d"`
	Balance7d         uint64 `db:"balance7d"`
	Balance31d        uint64 `db:"balance31d"`
	Status            string `db:"status"`
}

// ValidatorF is a cachable fixed size struct to hold validator data
type ValidatorF struct {
	Index                      uint64   `db:"validatorindex" json:"i"`
	PublicKey                  [48]byte `db:"pubkey" json:"pk"`
	Balance                    uint64   `db:"balance" json:"b"`
	EffectiveBalance           uint64   `db:"effectivebalance" json:"eb"`
	Slashed                    bool     `db:"slashed" json:"slashed"`
	ActivationEligibilityEpoch uint64   `db:"activationeligibilityepoch" json:"aee"`
	ActivationEpoch            uint64   `db:"activationepoch" json:"ae"`
	ExitEpoch                  uint64   `db:"exitepoch" json:"x"`
	WithdrawableEpoch          uint64   `db:"withdrawableepoch" json:"w"`
	WithdrawalCredentials      [32]byte `db:"withdrawalcredentials" json:"wc"`
	BalanceActivation          uint64   `db:"balanceactivation" json:"ba"`
	Balance1d                  uint64   `db:"balance1d" json:"b1"`
	Balance7d                  uint64   `db:"balance7d" json:"b7"`
	Balance31d                 uint64   `db:"balance31d" json:"b32"`
}

func (src *ValidatorF) ToValidator() *Validator {
	res := &Validator{
		Index:                      src.Index,
		Balance:                    src.Balance,
		EffectiveBalance:           src.EffectiveBalance,
		Slashed:                    src.Slashed,
		ActivationEligibilityEpoch: src.ActivationEligibilityEpoch,
		ActivationEpoch:            src.ActivationEpoch,
		ExitEpoch:                  src.ExitEpoch,
		WithdrawableEpoch:          src.WithdrawableEpoch,
		BalanceActivation:          src.BalanceActivation,
		Balance1d:                  src.Balance1d,
		Balance7d:                  src.Balance7d,
		Balance31d:                 src.Balance31d,
		Status:                     "",
	}
	copy(res.PublicKey, src.PublicKey[:])
	copy(res.WithdrawalCredentials, src.WithdrawalCredentials[:])
	return res
}

// ValidatorQueue is a struct to hold validator queue data
type ValidatorQueue struct {
	ChurnLimit                 uint64
	ActivationPublicKeys       [][]byte
	ExitPublicKeys             [][]byte
	ActivationValidatorIndices []uint64
	ExitValidatorIndices       []uint64
}

// Block is a struct to hold block data
type Block struct {
	Status            uint64
	Proposer          uint64
	BlockRoot         []byte
	Slot              uint64
	ParentRoot        []byte
	StateRoot         []byte
	Signature         []byte
	RandaoReveal      []byte
	Graffiti          []byte
	Eth1Data          *Eth1Data
	BodyRoot          []byte
	ProposerSlashings []*ProposerSlashing
	AttesterSlashings []*AttesterSlashing
	Attestations      []*Attestation
	Deposits          []*Deposit
	VoluntaryExits    []*VoluntaryExit
	Canonical         bool
}

// Eth1Data is a struct to hold the ETH1 data
type Eth1Data struct {
	DepositRoot  []byte
	DepositCount uint64
	BlockHash    []byte
}

// ProposerSlashing is a struct to hold proposer slashing data
type ProposerSlashing struct {
	ProposerIndex uint64
	Header1       *Block
	Header2       *Block
}

// AttesterSlashing is a struct to hold attester slashing
type AttesterSlashing struct {
	Attestation1 *IndexedAttestation
	Attestation2 *IndexedAttestation
}

// IndexedAttestation is a struct to hold indexed attestation data
type IndexedAttestation struct {
	Data             *AttestationData
	AttestingIndices []uint64
	Signature        []byte
}

// Attestation is a struct to hold attestation header data
type Attestation struct {
	AggregationBits []byte
	Attesters       []uint64
	Data            *AttestationData
	Signature       []byte
}

// AttestationData to hold attestation detail data
type AttestationData struct {
	Slot            uint64
	CommitteeIndex  uint64
	BeaconBlockRoot []byte
	Source          *Checkpoint
	Target          *Checkpoint
}

// Checkpoint is a struct to hold checkpoint data
type Checkpoint struct {
	Epoch uint64
	Root  []byte
}

// Deposit is a struct to hold deposit data
type Deposit struct {
	Proof                 [][]byte
	PublicKey             []byte
	WithdrawalCredentials []byte
	Amount                uint64
	Signature             []byte
}

// VoluntaryExit is a struct to hold voluntary exit data
type VoluntaryExit struct {
	Epoch          uint64
	ValidatorIndex uint64
	Signature      []byte
}

// BlockContainer is a struct to hold block container data
type BlockContainer struct {
	Status   uint64
	Proposer uint64

	Block *ethpb.BeaconBlockContainer
}

// MinimalBlock is a struct to hold minimal block data
type MinimalBlock struct {
	Epoch      uint64 `db:"epoch"`
	Slot       uint64 `db:"slot"`
	BlockRoot  []byte `db:"blockroot"`
	ParentRoot []byte `db:"parentroot"`
	Canonical  bool   `db:"-"`
}

// CanonBlock is a struct to hold canon block data
type CanonBlock struct {
	BlockRoot []byte `db:"blockroot"`
	Slot      uint64 `db:"slot"`
	Canonical bool   `db:"-"`
}

// BlockComparisonContainer is a struct to hold block comparison data
type BlockComparisonContainer struct {
	Epoch uint64
	Db    *MinimalBlock
	Node  *MinimalBlock
}

type AssignmentSlot struct {
	Proposer   uint64
	Committees [][]uint64
}

type Assignments struct {
	// Number of epoch
	Epoch uint32
	// typically 32
	NumSlots uint32
	// First slot in the epoch
	FirstSlot uint64
	// total assignments - helps in memory pre-allocation
	NumAssignments uint64
	// array of committees
	Assignments []AssignmentSlot
}

func (a *Assignments) ValidatorAt(slot uint64, committee uint64, index uint64) (uint64, bool) {
	return a.Assignments[slot-a.FirstSlot].Committees[committee][index], true
}
