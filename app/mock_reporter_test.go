package main

import (
	"net"

	"github.com/weaveworks/scope/report"
)

// StaticReport is used as know test data in api tests.
type StaticReport struct{}

func (s StaticReport) Report() report.Report {
	_, localNet, err := net.ParseCIDR("192.168.1.1/24")
	if err != nil {
		panic(err.Error())
	}

	var testReport = report.Report{
		Endpoint: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345")): report.MakeIDList(report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")),
				report.MakeAdjacencyID(report.MakeEndpointNodeID("hostA", "192.168.1.1", "12346")): report.MakeIDList(report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")),
				report.MakeAdjacencyID(report.MakeEndpointNodeID("hostA", "192.168.1.1", "8888")):  report.MakeIDList(report.MakeEndpointNodeID("", "1.2.3.4", "22")),
				report.MakeAdjacencyID(report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")):    report.MakeIDList(report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345")),
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345"), report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  200,
				},
				report.MakeEdgeID(report.MakeEndpointNodeID("hostA", "192.168.1.1", "12346"), report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  201,
				},
				report.MakeEdgeID(report.MakeEndpointNodeID("hostA", "192.168.1.1", "8888"), report.MakeEndpointNodeID("", "1.2.3.4", "80")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      200,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  202,
				},
				report.MakeEdgeID(report.MakeEndpointNodeID("hostB", "192.168.1.2", "80"), report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      0,
					BytesIngress:     12,
					WithConnCountTCP: true,
					MaxConnCountTCP:  203,
				},
			},
			NodeMetadatas: report.NodeMetadatas{
				report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345"): report.NodeMetadata{
					"addr":            "192.168.1.1",
					"port":            "12345",
					"pid":             "23128",
					report.HostNodeID: report.MakeHostNodeID("hostA"),
				},
				report.MakeEndpointNodeID("hostA", "192.168.1.1", "12346"): report.NodeMetadata{ // <-- same as :12345
					"addr":            "192.168.1.1",
					"port":            "12346",
					"pid":             "23128",
					report.HostNodeID: report.MakeHostNodeID("hostA"),
				},
				report.MakeEndpointNodeID("hostA", "192.168.1.1", "8888"): report.NodeMetadata{
					"addr":            "192.168.1.1",
					"port":            "8888",
					"pid":             "55100",
					report.HostNodeID: report.MakeHostNodeID("hostA"),
				},
				report.MakeEndpointNodeID("hostB", "192.168.1.2", "80"): report.NodeMetadata{
					"addr":            "192.168.1.2",
					"port":            "80",
					"pid":             "215",
					report.HostNodeID: report.MakeHostNodeID("hostB"),
				},
			},
		},

		Process: report.Topology{
			NodeMetadatas: report.NodeMetadatas{
				report.MakeProcessNodeID("hostA", "23128"): report.NodeMetadata{
					"pid":             "23128",
					"comm":            "curl",
					report.HostNodeID: report.MakeHostNodeID("hostA"),
				},
				report.MakeProcessNodeID("hostA", "8888"): report.NodeMetadata{
					"pid":             "8888",
					"comm":            "ssh",
					report.HostNodeID: report.MakeHostNodeID("hostA"),
				},
				report.MakeProcessNodeID("hostB", "80"): report.NodeMetadata{
					"pid":                 "80",
					"comm":                "apache",
					"docker_container_id": "abcdefg",
					report.HostNodeID:     report.MakeHostNodeID("hostB"),
				},
			},
		},

		Container: report.Topology{
			NodeMetadatas: report.NodeMetadatas{
				report.MakeContainerNodeID("hostB", "abcdefg"): report.NodeMetadata{
					"docker_container_id":   "abcdefg",
					"docker_container_name": "server",
					report.HostNodeID:       report.MakeHostNodeID("hostB"),
				},
			},
		},

		Address: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(report.MakeAddressNodeID("hostA", "192.168.1.1")): report.MakeIDList(report.MakeAddressNodeID("hostB", "192.168.1.2"), report.MakeAddressNodeID("", "1.2.3.4")),
				report.MakeAdjacencyID(report.MakeAddressNodeID("hostB", "192.168.1.2")): report.MakeIDList(report.MakeAddressNodeID("hostA", "192.168.1.1")),
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(report.MakeAddressNodeID("hostA", "192.168.1.1"), report.MakeAddressNodeID("hostB", "192.168.1.2")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  14,
				},
				report.MakeEdgeID(report.MakeAddressNodeID("hostA", "192.168.1.1"), report.MakeAddressNodeID("", "1.2.3.4")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      200,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  15,
				},
				report.MakeEdgeID(report.MakeAddressNodeID("hostB", "192.168.1.2"), report.MakeAddressNodeID("hostA", "192.168.1.1")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      0,
					BytesIngress:     12,
					WithConnCountTCP: true,
					MaxConnCountTCP:  16,
				},
			},
			NodeMetadatas: report.NodeMetadatas{
				report.MakeAddressNodeID("hostA", "192.168.1.1"): report.NodeMetadata{
					"addr":            "192.168.1.1",
					report.HostNodeID: report.MakeHostNodeID("hostA"),
				},
				report.MakeAddressNodeID("hostB", "192.168.1.2"): report.NodeMetadata{
					"addr":            "192.168.1.2",
					report.HostNodeID: report.MakeHostNodeID("hostB"),
				},
			},
		},

		Host: report.Topology{
			Adjacency:     report.Adjacency{},
			EdgeMetadatas: report.EdgeMetadatas{},
			NodeMetadatas: report.NodeMetadatas{
				report.MakeHostNodeID("hostA"): report.NodeMetadata{
					"host_name":       "node-a.local",
					"os":              "Linux",
					"local_networks":  localNet.String(),
					"load":            "3.14 2.71 1.61",
					report.HostNodeID: report.MakeHostNodeID("hostA"),
				},
				report.MakeHostNodeID("hostB"): report.NodeMetadata{
					"host_name":       "node-b.local",
					"os":              "Linux",
					"local_networks":  localNet.String(),
					report.HostNodeID: report.MakeHostNodeID("hostB"),
				},
			},
		},
	}
	return testReport
}
