package kubeController

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"warptail/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	cmclientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
)

const SecretName = "warptail-certificate"

type K8Controller struct {
	k8Client     *kubernetes.Clientset
	cmclient     *cmclientset.Clientset
	namespace    string
	ingressName  string
	serviceName  string
	ingressClass string
}

func getK8Config() (*rest.Config, error) {
	if config, err := rest.InClusterConfig(); err == nil {
		return config, err
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user home dir: %v", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to find kubernetes config: %v", err)
	}
	return kubeConfig, nil
}

func getCurrentNamespace() (string, error) {
	namespaceFilePath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	// Check if the file exists
	if _, err := os.Stat(namespaceFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("namespace file does not exist: %v", err)
	}

	// Read the namespace from the file
	namespaceBytes, err := os.ReadFile(namespaceFilePath)
	if err != nil {
		return "", fmt.Errorf("error reading namespace file: %v", err)
	}

	return string(namespaceBytes), nil
}

func NewK8Controller(cfg utils.K8Config) (*K8Controller, error) {
	config, err := getK8Config()
	if err != nil {
		return nil, err
	}
	k8client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	cmclient, err := cmclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	namespace, err := getCurrentNamespace()
	if err != nil {
		namespace = cfg.Namespace
	}

	return &K8Controller{
		k8Client:     k8client,
		cmclient:     cmclient,
		namespace:    namespace,
		ingressName:  cfg.IngressName,
		serviceName:  cfg.ServiceName,
		ingressClass: cfg.IngressClass,
	}, nil
}

// func getSecretName(host string) string {
// 	host = strings.TrimSpace(host)
// 	hostParts := strings.Split(host, ".")
// 	if hostParts[0] == "www" {
// 		host = hostParts[1]
// 	} else {
// 		host = hostParts[0]
// 	}
// 	return fmt.Sprintf("warptail-%s-certificate", host)
// }

func (ctrl *K8Controller) buildService(routes []utils.RouteConfig) corev1.Service {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctrl.ingressName,
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
			Port:     int32(route.Port),
			NodePort: int32(route.Port),
		}

		service.Spec.Ports = append(service.Spec.Ports, port)
	}
	return service
}

func (ctrl *K8Controller) createService(routes []utils.RouteConfig) error {
	service := ctrl.buildService(routes)
	existingService, err := ctrl.k8Client.CoreV1().Services(ctrl.namespace).Get(context.TODO(), ctrl.ingressName, metav1.GetOptions{})
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

func (ctrl *K8Controller) buildIngress(routes []utils.RouteConfig) networkingv1.Ingress {
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctrl.ingressName,
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
func (ctrl *K8Controller) Update(routes []utils.RouteConfig) error {
	if err := ctrl.createService(routes); err != nil {
		return err
	}
	if err := ctrl.createIngress(routes); err != nil {
		return err
	}
	if err := ctrl.createCertificate(routes); err != nil {
		return err
	}
	return nil
}

func (ctrl *K8Controller) createIngress(routes []utils.RouteConfig) error {
	ingress := ctrl.buildIngress(routes)
	existingIngress, err := ctrl.k8Client.NetworkingV1().Ingresses(ctrl.namespace).Get(context.TODO(), ctrl.ingressName, metav1.GetOptions{})
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

func (ctrl *K8Controller) buildCertificate(routes []utils.RouteConfig) certmanagerv1.Certificate {
	DNSNames := []string{}
	for _, route := range routes {
		DNSNames = append(DNSNames, route.Name)
	}

	return certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-cert", ctrl.ingressName),
			Namespace: ctrl.namespace,
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: SecretName,
			DNSNames:   DNSNames,
			IssuerRef: cmmeta.ObjectReference{
				Name: "letsencrypt-prod",
				Kind: "ClusterIssuer",
			},
		},
	}
}

func (ctrl *K8Controller) createCertificate(routes []utils.RouteConfig) error {
	certificate := ctrl.buildCertificate(routes)
	existingCertificate, err := ctrl.cmclient.CertmanagerV1().Certificates(ctrl.namespace).Get(context.TODO(), fmt.Sprintf("%s-cert", ctrl.ingressName), metav1.GetOptions{})
	if err != nil {
		fmt.Println("Certficate does not exist, creating a new one...")
		_, err := ctrl.cmclient.CertmanagerV1().Certificates(ctrl.namespace).Create(context.TODO(), &certificate, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Certficate: %v", err)
		}
		return nil
	}
	fmt.Println("Certficate exists, updating it...")
	existingCertificate.Spec = certificate.Spec
	_, err = ctrl.cmclient.CertmanagerV1().Certificates(ctrl.namespace).Update(context.TODO(), existingCertificate, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Ingress: %v", err)
	}
	return nil
}
