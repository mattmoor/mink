/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package postgressource

import (
	"context"
	"errors"
	"fmt"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"

	"database/sql"

	"github.com/vaikas/postgressource/pkg/apis/sources/v1alpha1"

	reconcilerpostgressource "github.com/vaikas/postgressource/pkg/client/injection/reconciler/sources/v1alpha1/postgressource"
	"github.com/vaikas/postgressource/pkg/reconciler"
	"github.com/vaikas/postgressource/pkg/reconciler/postgressource/resources"

	// Needed in case we need to open a db connection
	_ "github.com/lib/pq"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason SampleSourceReconciled.
func newReconciledNormal(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "PostgresSourceReconciled", "PostgresSource reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler reconciles a PostgresSource object
type Reconciler struct {
	ReceiveAdapterImage string `envconfig:"POSTGRES_SOURCE_RA_IMAGE" required:"true"`

	dr           *reconciler.DeploymentReconciler
	sbr          *reconciler.SinkBindingReconciler
	db           *sql.DB
	secretLister corev1listers.SecretLister
}

// Check that our Reconciler implements Interface
var _ reconcilerpostgressource.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.PostgresSource) pkgreconciler.Event {
	src.Status.InitializeConditions()
	src.Status.ObservedGeneration = src.Generation

	db, err := r.getDB(ctx, src)
	if err != nil {
		src.Status.PropagateFunctionCreated(false, err)
		return err
	}

	err = r.reconcileDBFunction(ctx, db, src)
	if err != nil {
		src.Status.PropagateFunctionCreated(false, err)
		logging.FromContext(ctx).Warnf("Failed to create function: %w", err)
		return err
	}
	src.Status.PropagateFunctionCreated(true, nil)

	for _, table := range src.Spec.Tables {
		logging.FromContext(ctx).Infof("Reconciling table: %q", table.Name)

		// Check that the table exists
		tableExists, err := r.checkTable(ctx, db, table.Name)
		if err != nil {
			src.Status.PropagateTriggersCreated(false, err)
			logging.FromContext(ctx).Warnf("Couldn't check the existence of table %q: %w", table.Name, err)
			return err
		}
		if !tableExists {
			src.Status.PropagateTriggersCreated(false, fmt.Errorf("Table %q does not exist", table.Name))
			logging.FromContext(ctx).Warnf("Table %q doesn't exist", table.Name)
			return err
		}

		err = r.reconcileDBTrigger(ctx, db, src, table.Name)
		if err != nil {
			src.Status.PropagateTriggersCreated(false, err)
			logging.FromContext(ctx).Warnf("Failed to reconcile triggers on table %q: %w", table.Name, err)
			return err
		}
		src.Status.PropagateTriggersCreated(true, nil)
	}

	ra, event := r.dr.ReconcileDeployment(ctx, src, resources.MakeReceiveAdapter(&resources.ReceiveAdapterArgs{
		EventSource:         src.Namespace + "/" + src.Name,
		Image:               r.ReceiveAdapterImage,
		NotificationChannel: resources.MakePostgresName(src),
		Source:              src,
		Labels:              resources.Labels(src.Name),
	}))
	if ra != nil {
		src.Status.PropagateDeploymentAvailability(ra)
	}
	if event != nil {
		logging.FromContext(ctx).Infof("returning because event from ReconcileDeployment")
		return event
	}

	if ra != nil {
		logging.FromContext(ctx).Info("going to ReconcileSinkBinding")
		sb, event := r.sbr.ReconcileSinkBinding(ctx, src, src.Spec.SourceSpec, tracker.Reference{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
			Namespace:  ra.Namespace,
			Name:       ra.Name,
		})
		logging.FromContext(ctx).Infof("ReconcileSinkBinding returned %#v", sb)
		if sb != nil {
			src.Status.MarkSink(sb.Status.SinkURI)
		}
		if event != nil {
			return event
		}
	}

	return newReconciledNormal(src.Namespace, src.Name)
}

// FinalizeKind removes the triggers and functions.
func (r *Reconciler) FinalizeKind(ctx context.Context, src *v1alpha1.PostgresSource) pkgreconciler.Event {
	db, err := r.getDB(ctx, src)
	if err != nil {
		return err
	}
	logging.FromContext(ctx).Infof("IN FINALIZE FOR \"%s/%s\"", src.Namespace, src.Name)
	for _, table := range src.Spec.Tables {
		logging.FromContext(ctx).Infof("Dropping triggers on table: %q", table.Name)
		err := r.dropTriggers(ctx, db, src, table.Name)
		if err != nil {
			return err
		}
	}
	err = r.dropFunction(ctx, db, src)
	if err != nil {
		return err
	}
	return newReconciledNormal(src.Namespace, src.Name)
}

// TODO: Diff the function in case it has changed and update it.
func (r *Reconciler) reconcileDBFunction(ctx context.Context, db *sql.DB, s *v1alpha1.PostgresSource) error {
	exists, err := r.checkFunction(ctx, db, s)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Found existing function")
		return nil
	}

	f := resources.MakeFunction(s)
	_, err = r.db.Exec(f)
	if err != nil {
		log.Printf("Failed to create function\n%s\nerr: %v", f, err)
		return err
	}
	return nil

}

func (r *Reconciler) reconcileDBTrigger(ctx context.Context, db *sql.DB, s *v1alpha1.PostgresSource, table string) error {
	exists, err := r.checkTriggers(ctx, db, s, table)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Found existing triggers on table: %q\n", table)
		return nil
	}
	t := resources.MakeTrigger(s, table)
	_, err = r.db.Exec(t)
	if err != nil {
		log.Printf("Failed to create trigger on table: %q\n%s\nerr: %v", table, t, err)
		return err
	}
	return nil

}

// Just make sure the table we're trying to create triggers against table that actually exists.
func (r *Reconciler) checkTable(ctx context.Context, db *sql.DB, table string) (bool, error) {
	rows, err := db.Query(resources.GetTableQuery, table)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			log.Fatal(err)
		}
		if tableName == table {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (r *Reconciler) checkTriggers(ctx context.Context, db *sql.DB, src *v1alpha1.PostgresSource, table string) (bool, error) {
	tName := resources.MakePostgresName(src)
	rows, err := db.Query(resources.GetTriggersQuery, table, tName)
	var insert, update, delete bool
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var triggerName, cond, table string
		err = rows.Scan(&triggerName, &cond, &table)
		if err != nil {
			return false, err
		}
		switch cond {
		case "INSERT":
			insert = true
		case "DELETE":
			delete = true
		case "UPDATE":
			update = true
		default:
			fmt.Printf("Found unknown action %q in table %q", cond, table)
		}
		fmt.Printf("%q %q %q\n", triggerName, cond, table)
	}
	return insert == true && update == true && delete == true, rows.Err()
}

func (r *Reconciler) checkFunction(ctx context.Context, db *sql.DB, src *v1alpha1.PostgresSource) (bool, error) {
	fName := resources.MakePostgresName(src)
	rows, err := db.Query(resources.GetFunctionQuery, fName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var functionName string
		err = rows.Scan(&functionName)
		if err != nil {
			return false, err
		}
		fmt.Printf("%q\n", functionName)
		if fName == functionName {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (r *Reconciler) dropFunction(ctx context.Context, db *sql.DB, src *v1alpha1.PostgresSource) error {
	_, err := db.Exec(resources.MakeDropFunction(src))
	return err
}

func (r *Reconciler) dropTriggers(ctx context.Context, db *sql.DB, src *v1alpha1.PostgresSource, table string) error {
	_, err := db.Exec(resources.MakeDropTrigger(src, table))
	return err
}

func (r *Reconciler) getDB(ctx context.Context, src *v1alpha1.PostgresSource) (*sql.DB, error) {
	// If there's only one db, use that.
	if r.db != nil && src.Spec.Secret == nil {
		return r.db, nil
	}

	// If there's no global one and not one specified in the spec, we can't
	// move forward, bail.
	if r.db == nil && src.Spec.Secret == nil {
		src.Status.PropagateFunctionCreated(false, errors.New("Database credentials not specified, can not proceed"))
		return nil, errors.New("Database credentials not specified, can not proceed")
	}

	// If the source specified the credentials to use, use them.
	secret := src.Spec.Secret
	if secret != nil {
		s, err := r.secretLister.Secrets(secret.Namespace).Get(secret.Name)
		if err != nil {
			return nil, err
		}
		if connstr, exists := s.Data["connectionstr"]; exists {
			logging.FromContext(ctx).Infof("GOT CONN STR AS %q", connstr)
			return sql.Open("postgres", string(connstr))
		}
	}
	return nil, errors.New("Failed to get a usable db connection")
}
