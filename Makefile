
CLUSTER_NAME        = exploring-cloud-native
CLUSTER_CONFIG_PATH	= cluster-config.yaml

.PHONY: cluster clean

cluster:
	kind create cluster --name $(CLUSTER_NAME) --config $(CLUSTER_CONFIG_PATH)

clean:
	kind delete cluster --name $(CLUSTER_NAME)
