
CLUSTER_NAME        = exploring-cloud-native
CLUSTER_CONFIG_PATH	= cluster-config.yaml

.PHONY: cluster cluster-rebuild port-forward clean

cluster:
	kind create cluster --name $(CLUSTER_NAME) --config $(CLUSTER_CONFIG_PATH)
	helm upgrade --install ingress-nginx ingress-nginx --repo https://kubernetes.github.io/ingress-nginx 
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.6.3/cert-manager.yaml
	sleep 10s
	kubectl apply -f ./cluster/observability/namespace.yaml
	kubectl create -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.47.0/jaeger-operator.yaml -n observability
	sleep 10s
	kubectl apply -R -f ./cluster/

cluster-rebuild: clean cluster

port-forward:
	kubectl port-forward service/ingress-nginx-controller 8080:80

clean:
	kind delete cluster --name $(CLUSTER_NAME)
