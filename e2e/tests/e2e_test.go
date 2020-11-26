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
	&AttestationFatFutureSigning{},
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
	&ProposalFatFutureSigning{},

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
