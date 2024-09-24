package kubeController

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"wormtail/pkg/utils"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	cmclientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
)

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

func (ctrl *K8Controller) buildIngress(routes []utils.RouteConfig) networkingv1.Ingress {
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctrl.ingressName,
			Namespace: ctrl.namespace,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ctrl.ingressClass,
			Rules:            []networkingv1.IngressRule{},
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
										Number: 80, // Update with your service's port
									},
								},
							},
						},
					},
				},
			},
		}
		ingress.Spec.Rules = append(ingress.Spec.Rules, rule)
	}
	// TLS: []networkingv1.IngressTLS{
	// 	{
	// 		Hosts:      []string{host},
	// 		SecretName: secretName, // The Secret that will be populated by cert-manager with the certificate
	// 	},
	// },
	return ingress
}

func (ctrl *K8Controller) CreateIngress(routes []utils.RouteConfig) error {
	ingress := ctrl.buildIngress(routes)

	// Check if the Ingress already exists
	existingIngress, err := ctrl.k8Client.NetworkingV1().Ingresses(ctrl.namespace).Get(context.TODO(), ctrl.ingressName, metav1.GetOptions{})
	if err != nil {
		// If the Ingress does not exist, create it
		fmt.Println("Ingress does not exist, creating a new one...")
		_, err := ctrl.k8Client.NetworkingV1().Ingresses(ctrl.namespace).Create(context.TODO(), &ingress, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Ingress: %v", err)
		}
		return nil
	}
	// If the Ingress exists, update it
	fmt.Println("Ingress exists, updating it...")
	existingIngress.ObjectMeta.Annotations = ingress.ObjectMeta.Annotations
	existingIngress.Spec = ingress.Spec
	_, err = ctrl.k8Client.NetworkingV1().Ingresses(ctrl.namespace).Update(context.TODO(), existingIngress, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Ingress: %v", err)
	}
	return nil
}

// createCertificate creates a cert-manager Certificate resource
func createCertificate(cmclient *cmclientset.Clientset, namespace string, certName string, secretName string, host string, issuerName string) (*certmanagerv1.Certificate, error) {
	// Define the Certificate object
	certificate := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: namespace,
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: secretName, // The Secret where cert-manager will store the certificate
			DNSNames:   []string{host},
			IssuerRef: cmmeta.ObjectReference{
				Name: issuerName,      // Name of the cert-manager Issuer or ClusterIssuer
				Kind: "ClusterIssuer", // ClusterIssuer or Issuer depending on your setup
			},
		},
	}

	// Create the Certificate resource
	createdCert, err := cmclient.CertmanagerV1().Certificates(namespace).Create(context.TODO(), certificate, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Certificate: %v", err)
	}

	return createdCert, nil
}
