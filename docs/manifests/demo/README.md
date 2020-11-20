# Cross-Cluster Service provision via KubeCarrier
## Install operators
### KubeMQ Operator
- Install OLM in the cluster:
`curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.17.0/install.sh | bash -s v0.17.0`
- Install the operator:
`kubectl create -f https://operatorhub.io/install/kubemq-operator.yaml`
- Verify operator is there:
`kubectl get csv -n operators`
Ref: https://operatorhub.io/operator/kubemq-operator

### Redis Operator
TODO

## Create KubeCarrier related objects
- Create two tenants in the system:
`kubectl apply -f accounts.yaml`
- Create `CatalogEntrySet` objects for both `Redis` and `KubeMQ` CRD:
`kubectl -n kubecarrier-provider apply -f ./kubemq/catalogentryset_kubemqcluster.yaml`
(TODO for redis catalogentryset)
- Create `Catalog` object:
`kubectl -n kubecarrier-provider apply -f catalog.yaml`

If everything is fine, you can verify `Offering` objects are created in the tenants' namespace:
```bash
$ get offering -A
```
```
NAMESPACE   NAME                                                  DISPLAY NAME     PROVIDER               AGE
tenant-a    kubemqclusters.service-cluster.kubecarrier-provider   KubeMQ-Cluster   kubecarrier-provider   68m
tenant-a    redis.service-cluster.kubecarrier-provider            Redis            kubecarrier-provider   68m
tenant-b    kubemqclusters.service-cluster.kubecarrier-provider   KubeMQ-Cluster   kubecarrier-provider   68m
tenant-b    redis.service-cluster.kubecarrier-provider            Redis            kubecarrier-provider   68m
```

## Create Custom Resources of Redis and KubeMQ
### KubeMQ
For provisioning an instance of KubeMQCluster for tenants, just do (don't forget to replace <LICENSE> with your license
in the yaml file!):
`kubectl -n tenant-a apply -f ./kubemq/tenant-kubemqcluster.yaml`
`kubectl -n tenant-b apply -f ./kubemq/tenant-kubemqcluster.yaml`

And you can see the workload (pods) are created in the namespaces which are assigned to tenants by kubecarrier:
`kubectl get pod -A | grep kubemq`
```
tenant-a-hpbzd         kubemq-cluster-0                                                  1/1     Running     0          71m
tenant-a-hpbzd         kubemq-cluster-1                                                  1/1     Running     0          70m
tenant-b-jr6fd         kubemq-cluster-0                                                  1/1     Running     0          70m
tenant-b-jr6fd         kubemq-cluster-1                                                  1/1     Running     0          70m
```

### Redis
TODO

## Consume services from another cluster
### KubeMQ
1. Go to the `kubemq/client` folder.
2. Modify the .kubemqctl-a.yaml file:
 - Replace <SERVICE_NAMESPACE> with the namespace of the service in the provider's cluster,
 - Replace <SERVICE_IP> with the IP which is provided by submariner.
3. Run ` NAMESPACE=tenant-a CONFIG_PATH=.kubemqctl-a.yaml make deploy`, and then a sender (crornjob) and receiver will be
created in the `tenant-a` namespace in the consumer cluster. sender will send `hello` message periodically, and receiver will
keep receiving message from the channel. (The channel name can be modified in the kubemq-client.yaml file.)
4. Use different `.kubemqctl.yaml` file and namespace do the same thing for other tenants.

### Redis
TODO


## Create another tenant
1. `kubectl apply -f tenant-c.yaml` in provider cluster.
2. Repeat the process from `Create Custom Resources of Redis and KubeMQ` to `Consume services from another cluster`
