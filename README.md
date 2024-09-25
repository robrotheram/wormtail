# WarpTail

WarpTail is a tool designed to simplify proxying connections from the internet to services hosted on your Tailscale tailnet. It offers secure and seamless access to private services on your tailnet using proxy techniques and supports both Docker and Kubernetes environments.

## Features
- Easy setup to expose services from your Tailscale tailnet to the internet.
- YAML-based configuration for flexibility.
- Dynamic port routing and management.
- Built-in dashboard for monitoring and control.
- Automated ingress management and traffic routing in Kubernetes.

---

## Getting Started

### Prerequisites
- A Tailscale account with a valid authentication key.
- A service running inside your tailnet that you want to expose to the internet.
- Docker or Kubernetes setup.

### Configuration

WarpTail uses a `config.yaml` file for all configuration management. The configuration covers settings for Tailscale authentication, dashboard access, and routing rules for exposing services.

#### Example `config.yaml`:
```yaml
tailscale:
  auth_key: tskey-auth-XXXXXXXXXXXXXXXXXXXXXXXXXXX
  hostname: WORMTAIL

dashboard:
  enabled: true
  username: admin
  password: mallard

# Specify Routes
routes:
    # Example HTTP Route
  - enabled: true
    name: immich.example.io
    type: http
    machine:
      address: 127.0.0.1
      port: 30041

    # Example TCP Route
  - enabled: true
    name: minecraft server
    type: tcp
    listen: 25565
    machine:
      address: 127.0.0.1
      port: 25565
      
# Optional Kubernetes-specific configuration
kubernetes:
  namespace: wormtail
  ingress_name: warptail-routes
  service_name: wormtail-service
  ingress_class: traefik


```

- **`tailscale.auth_key`**: Tailscale authentication key to authorize access to your tailnet.
- **`tailscale.hostname`**: The hostname used for your WarpTail instance on the tailnet.
- **`dashboard.enabled`**: Enables or disables the WarpTail dashboard.
- **`dashboard.username`** / **`dashboard.password`**: Credentials for accessing the WarpTail dashboard.
- **`kubernetes`**: Kubernetes-specific settings for managing ingress, services, and routing.
- **`routes`**: Define the services within your tailnet that you want to expose. Each route specifies a domain name, the protocol (`http` or `https`), and the internal machine's IP address and port.

---

## Running WarpTail on Docker

When running WarpTail in Docker, you'll need to mount the `config.yaml` to the container and decide between specifying all proxy ports upfront or using host networking for dynamic routing.

### 1. Specifying Ports Upfront

In this mode, you must specify all the ports you wish to proxy. Make sure your `config.yaml` has the correct port mappings defined in the `routes` section.

```bash
docker run -d \
  --name warptail \
  -v /path/to/config.yaml:/app/config.yaml \
  -p 80:80 \
  -p 443:443 \
  -p 30041:30041 \
  warptail:latest
```

- Mount the `config.yaml` file using `-v /path/to/config.yaml:/app/config.yaml`.
- Expose the ports defined in your configuration.

### 2. Using Host Networking

For dynamic port management, you can run WarpTail with Docker's host networking:

```bash
docker run -d \
  --name warptail \
  --network host \
  -v /path/to/config.yaml:/app/config.yaml \
  warptail:latest
```

Host networking allows WarpTail to dynamically route traffic without needing to expose individual ports.

---

## Running WarpTail on Kubernetes

WarpTail manages its own ingress and routes traffic through node-ports in Kubernetes. This requires creating a service account for it to handle ingress and service resources.

### 1. Setup Service Account

Create a service account, role, and role binding:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: warptail-sa
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: warptail-role
rules:
  - apiGroups: [""]
    resources: ["services", "pods"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: warptail-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: warptail-role
subjects:
  - kind: ServiceAccount
    name: warptail-sa
    namespace: default
```

### 2. Deploy WarpTail

Deploy WarpTail using your `config.yaml` file:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: warptail
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: warptail
  template:
    metadata:
      labels:
        app: warptail
    spec:
      serviceAccountName: warptail-sa
      containers:
        - name: warptail
          image: warptail:latest
          volumeMounts:
            - name: config-volume
              mountPath: /app/config.yaml
              subPath: config.yaml
      volumes:
        - name: config-volume
          configMap:
            name: warptail-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: warptail-config
  namespace: default
data:
  config.yaml: |
    tailscale:
      auth_key: tskey-auth-XXXXXXXXXXXXXXXXXXXXXXXXXXX
      hostname: WORMTAIL
    dashboard:
      enabled: true
      username: admin
      password: mallard
    kubernetes: 
      namespace: wormtail
      ingress_name: warptail-routes
      service_name: wormtail-service
      ingress_class: traefik
    routes:
      - enabled: true
        name: immich.exceptionerror.io
        type: http
        machine:
          address: 192.168.0.104
          port: 30041
---
apiVersion: v1
kind: Service
metadata:
  name: wormtail-service
  namespace: default
spec:
  type: NodePort
  selector:
    app: warptail
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
      nodePort: 30080
    - protocol: TCP
      port: 443
      targetPort: 443
      nodePort: 30443
```

### 3. Accessing the Service

Once deployed, WarpTail will automatically configure ingress and route traffic through node-ports. Access your exposed services through your Kubernetes cluster's external IP using the node-port (e.g., `http://<cluster-ip>:30080` for HTTP).

---

## Contributing
To contribute to WarpTail, please open an issue or submit a pull request. Contributions are always welcome!

---

## License
WarpTail is licensed under the MIT License. See `LICENSE` for more details.