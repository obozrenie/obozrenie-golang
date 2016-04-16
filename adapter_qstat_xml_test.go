package main

import (
	"testing"

	"github.com/skybon/goutil"
)

func TestLoadQStatXML(t *testing.T) {
	input := `
<?xml version="1.0" encoding="UTF-8"?>
<qstat>
	<server type="Q3M" address="master3.idsoftware.com" status="UP" servers="206">
	</server>
	<server type="Q3S" address="79.142.106.99:27963" status="UP">
		<hostname>79.142.106.99:27963</hostname>
		<name>Мой сервер</name>
		<gametype>osp</gametype>
		<map>ztn3tourney1</map>
		<numplayers>0</numplayers>
		<maxplayers>16</maxplayers>
		<numspectators>0</numspectators>
		<maxspectators>0</maxspectators>
		<ping>7</ping>
		<retries>0</retries>
	</server>
</qstat>
`

	fixture := qStatOutput{[]qStatServer{}}
	aST := "Q3M"
	aH := "master3.idsoftware.com"
	aS := "UP"
	a := qStatServer{ServerType: &aST, Hostname: &aH, Status: &aS}
	fixture.Servers = append(fixture.Servers, a)

	bST := "Q3S"
	bH := "79.142.106.99:27963"
	bS := "UP"
	bN := "Мой сервер"

	bGT := "osp"
	bMap := "ztn3tourney1"
	bNumP := 0
	bMaxP := 16
	bNumSpec := 0
	bMaxSpec := 0
	bPing := 7
	bRetries := 0

	b := qStatServer{&bST, &bH, &bS, &bN, &bGT, &bMap, &bNumP, &bMaxP, &bNumSpec, &bMaxSpec, &bPing, &bRetries, nil, nil}
	fixture.Servers = append(fixture.Servers, b)

	result, resultErr := loadQStatXML(input)

	if resultErr != nil {
		t.Error(goutil.ErrorOutJSON(resultErr, fixture, result))
		return
	}

	if !result.Equal(fixture) {
		t.Error(goutil.ErrorOutJSON(goutil.ErrMismatch, fixture, result))
		return
	}
}
