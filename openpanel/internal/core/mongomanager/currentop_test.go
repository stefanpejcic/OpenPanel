package mongomanager

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestStripSessionFields(t *testing.T) {
	cmd := bson.D{{Key: "find", Value: "orders"}, {Key: "filter", Value: bson.D{}}, {Key: "lsid", Value: bson.D{}}, {Key: "$db", Value: "app"}, {Key: "$clusterTime", Value: bson.D{}}}
	got := stripSessionFields(cmd)
	if len(got) != 2 || got[0].Key != "find" || got[1].Key != "filter" {
		t.Errorf("got %v, want only find and filter", got)
	}
}
