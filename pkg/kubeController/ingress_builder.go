package kubeController

import (
	"context"
	"fmt"
	"warptail/pkg/utils"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const INGRESS_NAME = "warptail-route-ingress"

func (ctrl *K8Controller) buildIngress(routes []utils.RouteConfig) networkingv1.Ingress {
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      INGRESS_NAME,
			Namespace: ctrl.namespace,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ctrl.ingressClass,
			Rules:            []networkingv1.IngressRule{},
			TLS:              []networkingv1.IngressTLS{},
		},
	}

	for _, route := range routes {
		if route.Type != utils.HTTP {
			continue
		}
		rule := networkingv1.IngressRule{
			Host: route.Name,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: func() *networkingv1.PathType { pathType := networkingv1.PathTypePrefix; return &pathType }(),
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: ctrl.serviceName,
									Port: networkingv1.ServiceBackendPort{
										Number: 80,
									},
								},
							},
						},
					},
				},
			},
		}
		tlsRule := networkingv1.IngressTLS{
			Hosts:      []string{route.Name},
			SecretName: SecretName,
		}
		ingress.Spec.Rules = append(ingress.Spec.Rules, rule)
		ingress.Spec.TLS = append(ingress.Spec.TLS, tlsRule)
	}
	return ingress
}

func (ctrl *K8Controller) getIngress() (*networkingv1.Ingress, error) {
	return ctrl.k8Client.NetworkingV1().Ingresses(ctrl.namespace).Get(context.TODO(), INGRESS_NAME, metav1.GetOptions{})
}

func (ctrl *K8Controller) deleteIngress() error {
	if _, err := ctrl.getIngress(); err == nil {
		return nil
	}
	return ctrl.k8Client.NetworkingV1().Ingresses(ctrl.namespace).Delete(context.TODO(), INGRESS_NAME, metav1.DeleteOptions{})
}

func (ctrl *K8Controller) createIngress(routes []utils.RouteConfig) error {
	ingress := ctrl.buildIngress(routes)
	if len(ingress.Spec.Rules) == 0 {
		fmt.Println("Ingress exists, deleting it...")
		return ctrl.deleteIngress()
	}
	existingIngress, err := ctrl.getIngress()
	if err != nil {
		fmt.Println("Ingress does not exist, creating a new one...")
		_, err := ctrl.k8Client.NetworkingV1().Ingresses(ctrl.namespace).Create(context.TODO(), &ingress, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Ingress: %v", err)
		}
		return nil
	}
	fmt.Println("Ingress exists, updating it...")
	existingIngress.Spec = ingress.Spec
	_, err = ctrl.k8Client.NetworkingV1().Ingresses(ctrl.namespace).Update(context.TODO(), existingIngress, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Ingress: %v", err)
	}
	return nil
}
