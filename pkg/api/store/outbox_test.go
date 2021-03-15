/*
Copyright 2021 Adevinta
*/

package store

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store/cdc"
)

// verifyOutbox is a testing helper function to verify outbox data.
func verifyOutbox(t *testing.T, store api.VulcanitoStore, expOp string, expDTO interface{}) {
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

	gotMap, err := bSliceToMapIface(outbox.DTO)
	if err != nil {
		t.Fatalf("error verifying outbox, error converting gotDTO: %v", err)
	}

	if !reflect.DeepEqual(expMap, gotMap) {
		t.Fatalf("error verifying outbox, DTO's do not match:\n%v\n%v", expMap, gotMap)
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
