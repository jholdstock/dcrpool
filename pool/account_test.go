package pool

import (
	"errors"
	"testing"
)

func testAccount(t *testing.T) {
	simnetAddrA := "Ssj6Sd54j11JM8qpenCwfwnKD73dsjm68ru"
	simnetAddrB := "SssPc1UNr8czcP3W9hfAgpmLRa3zJPDhfSy"

	// Create some valid accounts.
	accountA := NewAccount(simnetAddrA)
	err := db.persistAccount(accountA)
	if err != nil {
		t.Fatal(err)
	}

	accountB := NewAccount(simnetAddrB)
	err = db.persistAccount(accountB)
	if err != nil {
		t.Fatal(err)
	}

	// Creating the same account twice should fail.
	err = db.persistAccount(accountA)
	if !errors.Is(err, ErrValueFound) {
		t.Fatal("expected value found error")
	}

	// Fetch an account with its id.
	fetchedAccount, err := db.fetchAccount(accountA.UUID)
	if err != nil {
		t.Fatalf("fetchAccount error: %v", err)
	}

	if fetchedAccount.Address != accountA.Address {
		t.Fatalf("expected %v as fetched account address, got %v",
			accountA.Address, fetchedAccount.Address)
	}

	if fetchedAccount.UUID != accountA.UUID {
		t.Fatalf("expected %v as fetched account id, got %v",
			accountA.UUID, fetchedAccount.UUID)
	}

	// Delete all accounts.
	err = db.deleteAccount(accountA.UUID)
	if err != nil {
		t.Fatalf("delete accountA error: %v ", err)
	}

	err = db.deleteAccount(accountB.UUID)
	if err != nil {
		t.Fatalf("delete accountB error: %v ", err)
	}

	// Ensure the accounts have both been deleted.
	_, err = db.fetchAccount(accountA.UUID)
	if !errors.Is(err, ErrValueNotFound) {
		t.Fatal("expected value not found error")
	}

	_, err = db.fetchAccount(accountB.UUID)
	if !errors.Is(err, ErrValueNotFound) {
		t.Fatal("expected value not found error")
	}
}
