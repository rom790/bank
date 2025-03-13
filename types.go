package main

import (
	"math/rand"
)

type Account struct {
	ID        int
	FirstName string
	LastName  string
	Number    int64
	Balance   int64
}

func NewAccount(fName, lName string) *Account {
	return &Account{
		ID:        rand.Intn(100000),
		FirstName: fName,
		LastName:  lName,
		Number:    int64(rand.Intn(10000000)),
	}
}
