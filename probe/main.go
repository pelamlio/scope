package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/tag"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

var version = "dev" // set at build time

const linux = "linux" // runtime.GOOS

func main() {
	var (
		httpListen         = flag.String("http.listen", "", "listen address for HTTP profiling and instrumentation server")
		publishInterval    = flag.Duration("publish.interval", 3*time.Second, "publish (output) interval")
		spyInterval        = flag.Duration("spy.interval", time.Second, "spy (scan) interval")
		listen             = flag.String("listen", ":"+strconv.Itoa(xfer.ProbePort), "listen address")
		prometheusEndpoint = flag.String("prometheus.endpoint", "/metrics", "Prometheus metrics exposition endpoint (requires -http.listen)")
		spyProcs           = flag.Bool("processes", true, "report processes (needs root)")
		dockerEnabled      = flag.Bool("docker", true, "collect Docker-related attributes for processes")
		dockerInterval     = flag.Duration("docker.interval", 10*time.Second, "how often to update Docker attributes")
		weaveRouterAddr    = flag.String("weave.router.addr", "", "IP address or FQDN of the Weave router")
		procRoot           = flag.String("proc.root", "/proc", "location of the proc filesystem")
	)
	flag.Parse()

	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("probe version %s", version)

	procspy.SetProcRoot(*procRoot)

	if *httpListen != "" {
		log.Printf("profiling data being exported to %s", *httpListen)
		log.Printf("go tool pprof http://%s/debug/pprof/{profile,heap,block}", *httpListen)
		if *prometheusEndpoint != "" {
			log.Printf("exposing Prometheus endpoint at %s%s", *httpListen, *prometheusEndpoint)
			http.Handle(*prometheusEndpoint, makePrometheusHandler())
		}
		go func(err error) { log.Print(err) }(http.ListenAndServe(*httpListen, nil))
	}

	if *spyProcs && os.Getegid() != 0 {
		log.Printf("warning: process reporting enabled, but that requires root to find everything")
	}

	publisher, err := xfer.NewTCPPublisher(*listen)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	var (
		hostName = hostname()
		hostID   = hostName // TODO: we should sanitize the hostname
	)

	var (
		weaveTagger *tag.WeaveTagger
	)

	taggers := []tag.Tagger{tag.NewTopologyTagger(), tag.NewOriginHostTagger(hostID)}
	reporters := []tag.Reporter{}

	if *dockerEnabled && runtime.GOOS == linux {
		dockerRegistry, err := docker.NewRegistry(*dockerInterval)
		if err != nil {
			log.Fatalf("failed to start docker registry: %v", err)
		}
		defer dockerRegistry.Stop()

		taggers = append(taggers, docker.NewTagger(dockerRegistry, *procRoot))
		reporters = append(reporters, docker.NewReporter(dockerRegistry, hostID))
	}

	if *weaveRouterAddr != "" {
		var err error
		weaveTagger, err = tag.NewWeaveTagger(*weaveRouterAddr)
		if err != nil {
			log.Fatalf("failed to start Weave tagger: %v", err)
		}
		taggers = append(taggers, weaveTagger)
	}

	log.Printf("listening on %s", *listen)

	quit := make(chan struct{})
	defer close(quit)
	go func() {
		var (
			pubTick = time.Tick(*publishInterval)
			spyTick = time.Tick(*spyInterval)
			r       = report.MakeReport()
		)

		for {
			select {
			case <-pubTick:
				publishTicks.WithLabelValues().Add(1)
				publisher.Publish(r)
				r = report.MakeReport()

			case <-spyTick:
				r.Merge(spy(hostID, hostName, *spyProcs))

				// Do this every tick so it gets tagged by the OriginHostTagger
				r.Host = hostTopology(hostID, hostName)

				// TODO abstract PIDTree to a process provider, and provide an
				// alternate implementation for Darwin.
				if runtime.GOOS == linux {
					if pidTree, err := tag.NewPIDTree(*procRoot); err == nil {
						r.Process.Merge(pidTree.ProcessTopology(hostID))
					} else {
						log.Printf("PIDTree: %v", err)
					}
				}

				for _, reporter := range reporters {
					r.Merge(reporter.Report())
				}

				if weaveTagger != nil {
					r.Overlay.Merge(weaveTagger.OverlayTopology())
				}

				r = tag.Apply(r, taggers)

			case <-quit:
				return
			}
		}
	}()

	log.Printf("%s", <-interrupt())
}

// hostTopology produces a host topology for this host. No need to do this
// more than once per published report.
func hostTopology(hostID, hostName string) report.Topology {
	var localCIDRs []string
	if localNets, err := net.InterfaceAddrs(); err == nil {
		// Not all networks are IP networks.
		for _, localNet := range localNets {
			if ipNet, ok := localNet.(*net.IPNet); ok {
				localCIDRs = append(localCIDRs, ipNet.String())
			}
		}
	}
	t := report.NewTopology()
	t.NodeMetadatas[report.MakeHostNodeID(hostID)] = report.NodeMetadata{
		"ts":             time.Now().UTC().Format(time.RFC3339Nano),
		"host_name":      hostName,
		"local_networks": strings.Join(localCIDRs, " "),
		"os":             runtime.GOOS,
		"load":           getLoad(),
	}
	return t
}

func interrupt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}
