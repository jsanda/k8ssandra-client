package cleaner

import (
	"context"
	"log"

	cassdcapi "github.com/datastax/cass-operator/operator/pkg/apis/cassandra/v1beta1"
	medusa "github.com/k8ssandra/medusa-operator/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	managedLabel      = "app.kubernetes.io/managed-by"
	managedLabelValue = "Helm"
	releaseAnnotation = "meta.helm.sh/release-name"
	instanceLabel     = "app.kubernetes.io/instance"
)

// Agent is a cleaner utility for resources which helm pre-delete requires
type Agent struct {
	Client    client.Client
	Namespace string
}

// New returns a new instance of cleaning agent
func New(namespace string) (*Agent, error) {
	_ = api.AddToScheme(scheme.Scheme)
	_ = cassdcapi.AddToScheme(scheme.Scheme)

	c, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatal(err)
	}

	// Ensure all operations are targeting a single namespace
	c = client.NewNamespacedClient(c, namespace)

	return &Agent{
		Client:    c,
		Namespace: namespace,
	}, nil
}

// RemoveResources deletes all the resources with finalizers or which we want an operator to trigger a deletion
func (a *Agent) RemoveResources(releaseName string) error {
	// Remove CassandraDatacenter (cass-operator should delete all the finalizers and associated resources)
	if err := a.removeCassandraDatacenter(releaseName); err != nil {
		log.Fatalf("Failed to remove Cassandra cluster(s): %v", err)
		return err
	}
	return nil
}

func (a *Agent) removeCassandraBackups(cassdc *cassdcapi.CassandraDatacenter, releaseName string) error {
	// Should be related to a removed CassandraDatacenter.. spec.cassandraDatacenter has it
	list := &medusa.CassandraBackupList{}
	err := a.Client.List(context.Background(), list, client.InNamespace(cassdc.Namespace), client.MatchingLabels(map[string]string{instanceLabel: releaseName}))
	if err != nil {
		log.Fatalf("Failed to list CassandraBackups in namespace %s for release %s", a.Namespace, releaseName)
		return err
	}

	for _, backup := range list.Items {
		if backup.Spec.CassandraDatacenter == cassdc.Name {
			err = a.Client.Delete(context.Background(), &backup)
			if err != nil {
				log.Fatalf("Failed to delete CassandraBackup: %v\n", backup)
				return err
			}
		}
	}

	return nil
}

func (a *Agent) removeCassandraDatacenter(releaseName string) error {
	log.Printf("Removing CassandraDatacenter(s) managed in release %s from namespace %s\n", releaseName, a.Namespace)
	list := &cassdcapi.CassandraDatacenterList{}
	err := a.Client.List(context.Background(), list, client.InNamespace(a.Namespace), client.MatchingLabels(map[string]string{managedLabel: managedLabelValue}))
	if err != nil {
		log.Fatalf("Failed to list CassandraDatacenters in namespace: %s", a.Namespace)
		return err
	}

	for _, cassdc := range list.Items {
		if release, found := cassdc.Annotations[releaseAnnotation]; found {
			if release == releaseName {
				err = a.Client.Delete(context.Background(), &cassdc)
				if err != nil {
					log.Fatalf("Failed to delete CassandraDatacenter: %v\n", cassdc)
					return err
				}
			}
		}
	}

	return nil
}
