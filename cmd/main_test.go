package main

import (
	"testing"
	"os"
)

func TestMain(t *testing.T) {
	x := os.Getenv("MYSQL_DNS")
	if x != "root:root@tcp(localhost:3310)/ephemeris?charset=utf8&parseTime=True&loc=Local" {
		t.Errorf("TestMain failed: %s", x)
	}
}