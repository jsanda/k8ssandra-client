/*
	// TODO As kubectl plugin, we also need context and namespace (from the kubectl command itself..)

	kubectl k8ssandra command <node>/<cluster>/<release> <parameters>
	If no node / cluster / release is required, then the parameters is after command

	command selection, such as:
	remove <release> 				=> uninstall CassandraDatacenter for release X, but nothing else (cleaner)
	shutdown <cluster>				=> shutdown cluster X (but do not delete it)
	start <cluster>					=> resume operation of shutdown cluster
	nodetool <node>					=> use nodetool on node X
	restart <cluster>				=> issue rolling restart for cluster X
	cqlsh <node>					=> exec cqlsh in the node
	backup 							=> fetch backup information? *
	restore <cluster> <backup>		=> initiate a restore of cluster X to a version Y. What about to a new cluster?
	repair <cluster>				=> repair cluster X now
	get all							=> get all k8ssandra resources (more than kubectl get all)

	*
		* backup list <cluster>
		* backup create <cluster>
	--wait (wait for some process to finish)
*/
package main

import (
	"os"

	"github.com/spf13/pflag"

	"github.com/burmanm/k8ssandra-client/cmd/kubectl-k8ssandra/k8ssandra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-k8ssandra", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := k8ssandra.NewCmd(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}