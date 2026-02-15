package main

import "testing"

func TestMain(t *testing.T) {
	t.Run("table", func(t *testing.T) {
		if 1 != 1 {
			t.Fatal("unreachable")
		}
	})
}
