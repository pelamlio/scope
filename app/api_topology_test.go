package main

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

func TestAPITopologyApplications(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()
	is404(t, ts, "/api/topology/applications/foobar")
	{
		body := getRawJSON(t, ts, "/api/topology/applications")
		var topo APITopology
		if err := json.Unmarshal(body, &topo); err != nil {
			t.Fatal(err)
		}
		equals(t, 4, len(topo.Nodes))
		node, ok := topo.Nodes["pid:hostA:23128"]
		if !ok {
			t.Errorf("missing curl node")
		}
		equals(t, 1, len(node.Adjacency))
		equals(t, report.MakeIDList("pid:hostB:215"), node.Adjacency)
		equals(t, report.MakeIDList(
			report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345"),
			report.MakeEndpointNodeID("hostA", "192.168.1.1", "12346"),
			report.MakeProcessNodeID("hostA", "23128"),
			report.MakeHostNodeID("hostA"),
		), node.Origins)
		equals(t, "curl", node.LabelMajor)
		equals(t, "hostA (23128)", node.LabelMinor)
		equals(t, "23128", node.Rank)
		equals(t, false, node.Pseudo)
	}
	{
		body := getRawJSON(t, ts, "/api/topology/applications/pid:hostA:23128")
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatal(err)
		}
		equals(t, "pid:hostA:23128", node.Node.ID)
		equals(t, "curl", node.Node.LabelMajor)
		equals(t, "hostA (23128)", node.Node.LabelMinor)
		equals(t, false, node.Node.Pseudo)
		// Let's not unit-test the specific content of the detail tables
	}
	{
		body := getRawJSON(t, ts, "/api/topology/applications/pid:hostA:23128/pid:hostB:215")
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := render.AggregateMetadata{
			"egress_bytes":       24,
			"ingress_bytes":      0,
			"max_conn_count_tcp": 401,
		}
		if !reflect.DeepEqual(want, edge.Metadata) {
			t.Errorf("Edge metadata error. Want %v, have %v", want, edge)
		}
	}
}

func TestAPITopologyHosts(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()
	is404(t, ts, "/api/topology/hosts/foobar")
	{
		body := getRawJSON(t, ts, "/api/topology/hosts")
		var topo APITopology
		if err := json.Unmarshal(body, &topo); err != nil {
			t.Fatal(err)
		}
		equals(t, 3, len(topo.Nodes))
		node, ok := topo.Nodes["host:hostB"]
		if !ok {
			t.Errorf("missing host:hostB node")
			t.Errorf("%+v", topo)
		}
		equals(t, report.MakeIDList("host:hostA"), node.Adjacency)
		equals(t, report.MakeIDList(
			report.MakeAddressNodeID("hostB", "192.168.1.2"),
			report.MakeHostNodeID("hostB"),
		), node.Origins)
		equals(t, "node-b", node.LabelMajor)
		equals(t, "local", node.LabelMinor)
		equals(t, "local", node.Rank)
		equals(t, false, node.Pseudo)
	}
	{
		body := getRawJSON(t, ts, "/api/topology/hosts/host:hostB")
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatal(err)
		}
		equals(t, "host:hostB", node.Node.ID)
		equals(t, "node-b", node.Node.LabelMajor)
		equals(t, "local", node.Node.LabelMinor)
		equals(t, false, node.Node.Pseudo)
		// Let's not unit-test the specific content of the detail tables
	}
	{
		body := getRawJSON(t, ts, "/api/topology/hosts/host:hostB/host:hostA")
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := render.AggregateMetadata{
			"egress_bytes":       0,
			"ingress_bytes":      12,
			"max_conn_count_tcp": 16,
		}
		if !reflect.DeepEqual(want, edge.Metadata) {
			t.Errorf("Edge metadata error. Want %v, have %v", want, edge)
		}
	}
}

// Basic websocket test
func TestAPITopologyWebsocket(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()
	url := "/api/topology/applications/ws"

	// Not a websocket request
	res, _ := checkGet(t, ts, url)
	if have := res.StatusCode; have != 400 {
		t.Fatalf("Expected status %d, got %d.", 400, have)
	}

	// Proper websocket request
	ts.URL = "ws" + ts.URL[len("http"):]
	dialer := &websocket.Dialer{}
	ws, res, err := dialer.Dial(ts.URL+url, nil)
	ok(t, err)
	defer ws.Close()

	if want, have := 101, res.StatusCode; want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	_, p, err := ws.ReadMessage()
	ok(t, err)
	var d render.Diff
	if err := json.Unmarshal(p, &d); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 4, len(d.Add))
	equals(t, 0, len(d.Update))
	equals(t, 0, len(d.Remove))
}
