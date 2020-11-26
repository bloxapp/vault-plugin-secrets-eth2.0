package tests

import (
	"testing"
)

type E2E interface {
	Name() string
	Run(t *testing.T)
}

var tests = []E2E{
	// Attestation signing
	&AttestationSigning{},
	&AttestationSigningAccountNotFound{},
	&AttestationDoubleSigning{},
	&AttestationConcurrentSigning{},
	&AttestationFarFutureSigning{},
	&AttestationNoSlashingDataSigning{},

	// Aggregation signing
	&AggregationSigning{},
	&AggregationDoubleSigning{},
	&AggregationConcurrentSigning{},
	&AggregationSigningAccountNotFound{},

	//Proposal signing
	&ProposalSigning{},
	&ProposalDoubleSigning{},
	&ProposalConcurrentSigning{},
	&ProposalSigningAccountNotFound{},
	&ProposalFarFutureSigning{},

	// Accounts tests
	&AccountsList{},

	// Config tests
	&ConfigRead{},

	// Storage tests
	&SlashingStorageRead{},
}

func TestE2E(t *testing.T) {
	for _, tst := range tests {
		t.Run(tst.Name(), func(t *testing.T) {
			tst.Run(t)
		})
	}
}
