/*
Copyright 2021 Adevinta
*/

package store

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store/cdc"
	"github.com/google/go-cmp/cmp"
)

type expOutbox struct {
	action     string
	notPresent bool
	dto        interface{}
}

// Compare compares a expected outbox structure with an outbox
// record.
func (e expOutbox) Compare(outbox cdc.Outbox, ignoreFields map[string][]string) string {
	var diff string
	if e.action != outbox.Operation {
		diff = fmt.Sprintf("error verifying outbox, expected Op to be %s but got %s",
			e.action, outbox.Operation)
		return diff
	}
	expDTO := e.dto
	expMap, err := ifaceToMapIface(expDTO)
	if err != nil {
		diff = fmt.Sprintf("error verifying outbox, error converting expDTO: %v", err)
		return diff
	}
	filterFields(expMap, ignoreFields)

	gotMap, err := bSliceToMapIface(outbox.DTO)
	if err != nil {
		diff = fmt.Sprintf("error verifying outbox, error converting gotDTO: %v", err)
		return diff
	}
	filterFields(gotMap, ignoreFields)
	diff = cmp.Diff(expMap, gotMap)
	if diff != "" {
		diff = fmt.Sprintf("error verifying outbox, DTO's do not match.\nDiff:\n%v", diff)
		return diff
	}
	return ""
}

// verifyOutbox is a testing helper function to verify outbox data.
func verifyOutbox(t *testing.T, store api.VulcanitoStore, exp expOutbox, ignoreFields map[string][]string) {
	t.Helper()

	expOp := exp.action
	expDTO := exp.dto
	expNotPresent := exp.notPresent
	var outbox cdc.Outbox
	db := store.(Store)
	err := db.Conn.Raw(`
		SELECT * FROM outbox
		ORDER BY created_at DESC
		LIMIT 1`,
	).Scan(&outbox).Error
	if expNotPresent {
		// If no outbox data should be present and we had a NotFoundError the
		// verification is okey.
		if db.NotFoundError(err) {
			return
		}
		if err != nil {
			t.Fatalf("error verifying outbox: %v", err)
		}
		t.Fatal("error verifying outbox: no records expected but some found")
	}

	if err != nil {
		t.Fatalf("error verifying outbox: %v", err)
	}

	if outbox.Operation != expOp {
		t.Fatalf("error verifying outbox, expected Op to be %s but got %s",
			expOp, outbox.Operation)
	}

	expMap, err := ifaceToMapIface(expDTO)
	if err != nil {
		t.Fatalf("error verifying outbox, error converting expDTO: %v", err)
	}
	filterFields(expMap, ignoreFields)

	gotMap, err := bSliceToMapIface(outbox.DTO)
	if err != nil {
		t.Fatalf("error verifying outbox, error converting gotDTO: %v", err)
	}
	filterFields(gotMap, ignoreFields)

	diff := cmp.Diff(expMap, gotMap)
	if diff != "" {
		t.Fatalf("error verifying outbox, DTO's do not match.\nDiff:\n%v", diff)
	}
}

type testCdCOutbox struct {
	DTO       string
	Operation string
}

func filterFields(m map[string]interface{}, ignoreFields map[string][]string) {
	for objName, objFields := range ignoreFields {
		obj := m[objName].(map[string]interface{})
		for _, f := range objFields {
			delete(obj, f)
		}
	}
}

func ifaceToMapIface(i interface{}) (map[string]interface{}, error) {
	iData, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return bSliceToMapIface(iData)
}

func bSliceToMapIface(s []byte) (map[string]interface{}, error) {
	var iMap map[string]interface{}
	err := json.Unmarshal(s, &iMap)
	if err != nil {
		return nil, err
	}
	return iMap, nil
}
