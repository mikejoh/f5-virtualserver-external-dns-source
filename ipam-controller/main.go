package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	f5 "github.com/F5Networks/k8s-bigip-ctlr/v2/config/apis/cis/v1"
)

type ipamControllerOptions struct {
	vsAddress string
	vsName    string
	namespace string
	status    string
}

func main() {
	var opts ipamControllerOptions
	flag.StringVar(&opts.vsAddress, "vs-address", "192.168.1.101", "Virtual Server Address")
	flag.StringVar(&opts.vsName, "vs-name", "example-vs-ipam", "Virtual Server Name")
	flag.StringVar(&opts.namespace, "namespace", "default", "Namespace")
	flag.StringVar(&opts.status, "status", "Ok", "Status")
	flag.Parse()

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	scheme := scheme.Scheme
	if err := f5.AddToScheme(scheme); err != nil {
		log.Fatalf("Error adding CRD to scheme: %v\n", err)
	}

	c, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatalf("Error creating client: %v\n", err)
	}

	vs := f5.VirtualServer{}
	vsName := opts.vsName
	namespace := opts.namespace

	err = c.Get(context.TODO(), client.ObjectKey{
		Namespace: namespace,
		Name:      vsName,
	}, &vs)
	if err != nil {
		log.Fatalf("Error fetching CRD: %v\n", err)
	}

	status := f5.VirtualServerStatus{
		VSAddress: opts.vsAddress,
		Status:    opts.status,
		Error:     "",
	}

	vs.Status = status

	err = c.Status().Update(context.TODO(), &vs, &client.SubResourceUpdateOptions{})
	if err != nil {
		log.Fatalf("Error updating CRD: %v\n", err)
	}

	fmt.Printf("Updated CRD with status: %v\n", vs.Status)
}
