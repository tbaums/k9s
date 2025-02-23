package dao

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const maxJobNameSize = 42

var (
	_ Accessor = (*CronJob)(nil)
	_ Runnable = (*CronJob)(nil)
)

// CronJob represents a cronjob K8s resource.
type CronJob struct {
	Generic
}

// Run a CronJob.
func (c *CronJob) Run(path string) error {
	ns, n := client.Namespaced(path)
	auth, err := c.Client().CanI(ns, "batch/v1beta1/cronjobs", []string{client.GetVerb, client.CreateVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorize to run cronjobs")
	}

	// BOZO!! Factory resource??
	ctx, cancel := context.WithTimeout(context.Background(), client.CallTimeout)
	defer cancel()
	cj, err := c.Client().DialOrDie().BatchV1beta1().CronJobs(ns).Get(ctx, n, metav1.GetOptions{})
	if err != nil {
		return err
	}

	var jobName = cj.Name
	if len(cj.Name) >= maxJobNameSize {
		jobName = cj.Name[0:maxJobNameSize]
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName + "-manual-" + rand.String(3),
			Namespace: ns,
			Labels:    cj.Spec.JobTemplate.Labels,
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}
	_, err = c.Client().DialOrDie().BatchV1().Jobs(ns).Create(ctx, job, metav1.CreateOptions{})

	return err
}
