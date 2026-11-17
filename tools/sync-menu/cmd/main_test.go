package main

import "testing"

// 栄養値の比較は nil 安全でなければならない。
// 片方だけ nil のときに panic したり、誤って「同じ」と判定してはいけない。
func TestSameInt(t *testing.T) {
	three := 3
	other := 5

	if !sameInt(nil, nil) {
		t.Error("nil, nil should be equal")
	}
	if sameInt(&three, nil) {
		t.Error("3 and nil should differ")
	}
	if sameInt(nil, &three) {
		t.Error("nil and 3 should differ")
	}
	if !sameInt(&three, &three) {
		t.Error("3 and 3 should be equal")
	}
	if sameInt(&three, &other) {
		t.Error("3 and 5 should differ")
	}
}

func TestSameFloat(t *testing.T) {
	one := 1.4
	other := 0.9

	if !sameFloat(nil, nil) {
		t.Error("nil, nil should be equal")
	}
	if sameFloat(&one, nil) {
		t.Error("1.4 and nil should differ")
	}
	if !sameFloat(&one, &one) {
		t.Error("1.4 and 1.4 should be equal")
	}
	if sameFloat(&one, &other) {
		t.Error("1.4 and 0.9 should differ")
	}
}
