package tushare

import "testing"

func TestCBBasicFromRow(t *testing.T) {
	row := Row{
		"ts_code":         "110059.SH",
		"bond_short_name": "浦发转债",
		"stk_code":        "600000.SH",
		"cb_type":         "CB",
		"par":             "100",
		"coupon_rate":     "0.2",
		"list_date":       "20200101",
	}
	it, ok := cbBasicFromRow(row)
	if !ok {
		t.Fatal("expected ok")
	}
	if it.TSCode != "110059.SH" || it.BondShortName != "浦发转债" || it.StkCode != "600000.SH" {
		t.Fatalf("%+v", it)
	}
	if it.Par == nil || *it.Par != 100 {
		t.Fatalf("par=%v", it.Par)
	}
	if it.ListDate == nil {
		t.Fatal("list_date")
	}
}

func TestCBTopHolderFromRow(t *testing.T) {
	row := Row{
		"ts_code":     "110059.SH",
		"end_date":    "20250630",
		"holder_rank": "1",
		"holder_name": "测试持有人",
		"hold_amount": "100.5",
		"hold_ratio":  "2.3",
	}
	it, ok := cbTopHolderFromRow(row)
	if !ok || it.HolderRank != 1 || it.HolderName != "测试持有人" {
		t.Fatalf("%+v ok=%v", it, ok)
	}
	if it.HoldAmount == nil || *it.HoldAmount != 100.5 {
		t.Fatalf("amount=%v", it.HoldAmount)
	}
}
