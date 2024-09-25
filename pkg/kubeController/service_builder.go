package kubeController

import (
	"context"
	"fmt"
	"warptail/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const SERVICE_NAME = "warptail-route-service"

func (ctrl *K8Controller) buildService(routes []utils.RouteConfig) corev1.Service {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SERVICE_NAME,
			Namespace: ctrl.namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				"app": "warptail",
			},
			Ports: []corev1.ServicePort{},
		},
	}

	for _, route := range routes {
		if route.Type != utils.TCP && route.Type != utils.UDP {
			continue
		}
		port := corev1.ServicePort{
			Port:       int32(route.Port),
			TargetPort: intstr.FromInt(route.Port),
		}
		service.Spec.Ports = append(service.Spec.Ports, port)
	}
	return service
}

func (ctrl *K8Controller) getService() (*corev1.Service, error) {
	return ctrl.k8Client.CoreV1().Services(ctrl.namespace).Get(context.TODO(), SERVICE_NAME, metav1.GetOptions{})
}

func (ctrl *K8Controller) deleteService() error {
	if _, err := ctrl.getService(); err == nil {
		return nil
	}
	return ctrl.k8Client.CoreV1().Services(ctrl.namespace).Delete(context.TODO(), SERVICE_NAME, metav1.DeleteOptions{})
}

func (ctrl *K8Controller) createService(routes []utils.RouteConfig) error {
	service := ctrl.buildService(routes)
	if len(service.Spec.Ports) == 0 {
		fmt.Println("Service exists, deleting it...")
		return ctrl.deleteService()
	}

	existingService, err := ctrl.getService()
	if err != nil {
		fmt.Println("Service does not exist, creating a new one...")
		_, err := ctrl.k8Client.CoreV1().Services(ctrl.namespace).Create(context.TODO(), &service, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Service: %v", err)
		}
		return nil
	}
	fmt.Println("Service exists, updating it...")
	existingService.Spec = service.Spec
	_, err = ctrl.k8Client.CoreV1().Services(ctrl.namespace).Update(context.TODO(), existingService, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Service: %v", err)
	}
	return nil
}
