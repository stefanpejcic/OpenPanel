package mongomanager

import (
	"context"
	"math"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// maxCommandLen caps how much of an op's command document is shown in the processlist
const maxCommandLen = 500

// Op is one in-progress operation from $currentOp.
type Op struct {
	OpID           int64
	Desc           string
	Op             string
	NS             string
	Client         string
	AppName        string
	User           string
	Command        string
	SecsRunning    int64
	Active         bool
	WaitingForLock bool
	// IsClient is false for mongod's own background threads (TTLMonitor, Checkpointer, ...), which have no connectionId
	IsClient bool
}

func toInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case int32:
		return int64(t), true
	case int64:
		return t, true
	case float64:
		return int64(t), true
	default:
		return 0, false
	}
}

// heartbeatCommands are drivers' awaitable topology checks, they sit "running" on purpose so they're left out
var heartbeatCommands = map[string]bool{"hello": true, "isMaster": true, "ismaster": true}

// sessionFields are driver bookkeeping that bury the actual query in the command column
var sessionFields = map[string]bool{"lsid": true, "$clusterTime": true, "$db": true, "$readPreference": true, "txnNumber": true, "apiVersion": true}

func stripSessionFields(cmd bson.D) bson.D {
	out := make(bson.D, 0, len(cmd))
	for _, e := range cmd {
		if !sessionFields[e.Key] {
			out = append(out, e)
		}
	}
	return out
}

func str(m bson.M, key string) string {
	s, _ := m[key].(string)
	return s
}

// CurrentOps lists active operations for every user, minus the $currentOp call that's asking
func CurrentOps(ctx context.Context, userContext string) ([]Op, error) {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return nil, err
	}
	pipeline := mongo.Pipeline{{{Key: "$currentOp", Value: bson.D{{Key: "allUsers", Value: true}, {Key: "idleConnections", Value: false}}}}}
	cur, err := client.Database("admin").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var docs []bson.M
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}

	ops := make([]Op, 0, len(docs))
	for _, d := range docs {
		opid, ok := toInt64(d["opid"])
		if !ok {
			continue
		}
		cmd, _ := d["command"].(bson.D)
		if len(cmd) > 0 && heartbeatCommands[cmd[0].Key] {
			continue
		}
		command := ""
		if len(cmd) > 0 {
			if b, err := bson.MarshalExtJSON(stripSessionFields(cmd), false, false); err == nil {
				command = string(b)
			}
		}
		if strings.Contains(command, `"$currentOp"`) {
			continue
		}
		if len(command) > maxCommandLen {
			command = command[:maxCommandLen] + "…"
		}

		op := Op{
			OpID: opid, Desc: str(d, "desc"), Op: str(d, "op"), NS: str(d, "ns"),
			Client: str(d, "client"), AppName: str(d, "appName"), Command: command,
		}
		op.SecsRunning, _ = toInt64(d["secs_running"])
		op.Active, _ = d["active"].(bool)
		op.WaitingForLock, _ = d["waitingForLock"].(bool)
		_, op.IsClient = d["connectionId"]
		if users, ok := d["effectiveUsers"].(bson.A); ok && len(users) > 0 {
			switch u := users[0].(type) {
			case bson.D:
				for _, e := range u {
					if e.Key == "user" {
						op.User, _ = e.Value.(string)
					}
				}
			case bson.M:
				op.User = str(u, "user")
			}
		}
		ops = append(ops, op)
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].OpID < ops[j].OpID })
	return ops, nil
}

// KillOp asks mongod to terminate one operation, the connection itself stays open
func KillOp(ctx context.Context, userContext string, opid int64) error {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return err
	}
	var op any = opid
	if opid <= math.MaxInt32 {
		op = int32(opid)
	}
	cmd := bson.D{{Key: "killOp", Value: 1}, {Key: "op", Value: op}}
	return client.Database("admin").RunCommand(ctx, cmd).Err()
}
