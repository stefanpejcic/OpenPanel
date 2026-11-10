package imunifyav

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"
	"time"
)

// socketPath is imunify's root rpc socket, bind-mounted from the host since the agent and its cli only live there
var socketPath = "/run/defence360agent/simple_rpc.sock"

var errNotRunning = errors.New("ImunifyAV is not running on this server")

// available is false when imunify isn't installed or the agent's socket isn't there
var available = func() bool {
	info, err := os.Stat(socketPath)
	return err == nil && info.Mode()&os.ModeSocket != 0
}

// call sends one request the way imunify360-agent's own cli does, a json line in and a json line back
var call = func(command []string, params map[string]any) (json.RawMessage, error) {
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		return nil, errNotRunning
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(20 * time.Second))

	req, err := json.Marshal(map[string]any{"command": command, "params": params})
	if err != nil {
		return nil, err
	}
	if _, err := conn.Write(append(req, '\n')); err != nil {
		return nil, err
	}
	line, err := bufio.NewReaderSize(conn, 64*1024).ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return decodeReply(line)
}

// decodeReply unwraps data from a reply, or turns an error reply into its messages
func decodeReply(line []byte) (json.RawMessage, error) {
	var reply struct {
		Result   string          `json:"result"`
		Messages json.RawMessage `json:"messages"`
		Data     json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(line, &reply); err != nil {
		return nil, err
	}
	if reply.Result != "success" {
		return nil, errors.New(replyMessages(reply.Messages))
	}
	return reply.Data, nil
}

// replyMessages flattens messages, which imunify sends as a list of strings or as one string
func replyMessages(raw json.RawMessage) string {
	var list []string
	if json.Unmarshal(raw, &list) == nil && len(list) > 0 {
		return strings.Join(list, "; ")
	}
	var one string
	if json.Unmarshal(raw, &one) == nil && one != "" {
		return one
	}
	return "ImunifyAV returned an error"
}

// items decodes the items list most replies carry
func items(data json.RawMessage, out any) error {
	var body struct {
		Items json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return err
	}
	if len(body.Items) == 0 || string(body.Items) == "null" {
		return nil
	}
	return json.Unmarshal(body.Items, out)
}
