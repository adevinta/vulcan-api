/*
Copyright 2021 Adevinta
*/

package store

import (
	"encoding/json"
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

// verifyOutbox is a testing helper function to verify outbox data.
func verifyOutbox(t *testing.T, store api.VulcanitoStore, exp expOutbox, ignoreFields map[string][]string) {
	t.Helper()

	expOp := exp.action
	expDTO := exp.dto
	expNotPresent := exp.notPresent
	var outbox cdc.Outbox
	db := store.(vulcanitoStore)
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
