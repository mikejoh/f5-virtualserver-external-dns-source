package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	f5 "github.com/F5Networks/k8s-bigip-ctlr/v2/config/apis/cis/v1"
)

type ipamControllerOptions struct {
	vsAddress   string
	vsName      string
	namespace   string
	status      string
	interval    int
	keepRunning bool
}

func main() {
	var opts ipamControllerOptions
	flag.StringVar(&opts.vsAddress, "vs-address", "192.168.1.101", "VirtualServer address")
	flag.StringVar(&opts.vsName, "vs-name", "example-vs-ipam", "VirtualServer name")
	flag.StringVar(&opts.namespace, "namespace", "default", "Namespace")
	flag.StringVar(&opts.status, "status", "Ok", "Status")
	flag.IntVar(&opts.interval, "interval", 15, "Interval in seconds to update status")
	flag.BoolVar(&opts.keepRunning, "keep-running", false, "Run as daemon that updates status every interval seconds")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		slog.Error("error building kubeconfig", "error", err)
		os.Exit(1)
	}

	scheme := scheme.Scheme
	if err := f5.AddToScheme(scheme); err != nil {
		slog.Error("error adding CRD to scheme", "error", err)
		os.Exit(1)
	}

	c, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		slog.Error("error creating client", "error", err)
		os.Exit(1)
	}

	statuses := []string{"Ok", "ERROR", ""}
	vsAddresses := []string{"192.168.1.102", ""}

	if opts.keepRunning {
		logger.Info("starting", "app", "vs-status-updater", "interval", opts.interval)
		go func() {
			ticker := time.NewTicker(time.Duration(opts.interval) * time.Second)
			for {
				select {
				case <-ticker.C:
					randomStatus := statuses[time.Now().Nanosecond()%len(statuses)]
					randomVSAddress := vsAddresses[time.Now().Nanosecond()%len(vsAddresses)]
					err = updateStatus(context.Background(), c, randomVSAddress, opts.vsName, opts.namespace, randomStatus)
					if err != nil {
						slog.Error("error updating status of F5 VirtualServer CRD", "error", err)
						os.Exit(1)
					}

					slog.Info("updated status of F5 VirtualServer CRD", "status", opts.status, "vsAddress", opts.vsAddress, "vsName", opts.vsName, "namespace", opts.namespace)
				}
			}
		}()

		select {}
	} else {
		err = updateStatus(context.Background(), c, opts.vsAddress, opts.vsName, opts.namespace, opts.status)
		if err != nil {
			slog.Error("error updating status of F5 VirtualServer CRD", "error", err)
			os.Exit(1)
		}

		slog.Info("updated status of F5 VirtualServer CRD", "status", opts.status, "vsAddress", opts.vsAddress, "vsName", opts.vsName, "namespace", opts.namespace)
	}
}

func updateStatus(ctx context.Context, c client.Client, vsAddress, vsName, namespace, status string) error {
	vs := f5.VirtualServer{}

	err := c.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      vsName,
	}, &vs)
	if err != nil {
		return fmt.Errorf("error getting CRD: %v", err)
	}

	vsStatus := f5.VirtualServerStatus{
		VSAddress: vsAddress,
		Status:    status,
		Error:     "",
	}

	vs.Status = vsStatus

	err = c.Status().Update(context.TODO(), &vs, &client.SubResourceUpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating CRD: %v", err)
	}

	return nil
}
