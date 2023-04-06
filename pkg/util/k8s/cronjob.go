package k8s

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CronJob interface {
	GetCronJob(namespace, name string) (*batchv1.CronJob, error)
	CreateCronJob(namespace string, cronJob *batchv1.CronJob) error
	UpdateCronJob(namespace string, cronJob *batchv1.CronJob) error
	CreateOrUpdateCronJob(namespace string, cronJob *batchv1.CronJob) error
	DeleteCronJob(namespace, name string) error
	ListCronJobs(namespace string) (*batchv1.CronJobList, error)
}

type CronJobResource struct {
	clientset *kubernetes.Clientset
}

func NewCronJobResource(clientset *kubernetes.Clientset) *CronJobResource {
	return &CronJobResource{
		clientset: clientset,
	}
}

func (c *CronJobResource) GetCronJob(namespace, name string) (*batchv1.CronJob, error) {
	return c.clientset.BatchV1().CronJobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (c *CronJobResource) CreateCronJob(namespace string, cronJob *batchv1.CronJob) error {
	_, err := c.clientset.BatchV1().CronJobs(namespace).Create(context.TODO(), cronJob, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *CronJobResource) UpdateCronJob(namespace string, cronJob *batchv1.CronJob) error {
	_, err := c.clientset.BatchV1().CronJobs(namespace).Update(context.TODO(), cronJob, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *CronJobResource) CreateOrUpdateCronJob(namespace string, cronJob *batchv1.CronJob) error {
	storedCronJob, err := c.GetCronJob(namespace, cronJob.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.CreateCronJob(namespace, cronJob)
		}
	}

	cronJob.ResourceVersion = storedCronJob.ResourceVersion
	return c.UpdateCronJob(namespace, cronJob)
}

func (c *CronJobResource) DeleteCronJob(namespace, name string) error {
	return c.clientset.BatchV1().CronJobs(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (c *CronJobResource) ListCronJobs(namespace string) (*batchv1.CronJobList, error) {
	return c.clientset.BatchV1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{})
}
