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

// verifyOutbox is a testing helper function to verify outbox data.
func verifyOutbox(t *testing.T, store api.VulcanitoStore, expOp string, expDTO interface{}, ignoreFields map[string][]string) {
	t.Helper()

	var outbox cdc.Outbox

	err := store.(vulcanitoStore).Conn.Raw(`
		SELECT * FROM outbox
		ORDER BY created_at DESC
		LIMIT 1`,
	).Scan(&outbox).Error
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
