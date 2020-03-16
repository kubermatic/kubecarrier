# Kubecarrier multi-region proposal

## Motivation

Currently the KubeCarrier can only effectively operate in a single region, e.g. eu-central. Due to its cross-cluster syncing nature a high latency between management cluster and service cluster would impact the quality of service.

For truly global scoped KubeCarrier installation current architecture isn't satisfactory. Here are some of the problems honest and hard-working system operator could encounter:

* latency --> It's as simple as managing `eu-central` from a `east-asia` region would take between 200 and 400ms. This may or may not be satisfactory for the system requirement, but it's a matter of life. And speed of light. Certain operations require multiple round trips, increasing the overall latency.
* Failure domains --> Failure of management cluster, or its unreachability would make the whole system immutable. However unlikely the whole management cluster becomes unavailable, it's a possible failure mode
* Legalese & policy --> for some legal requirements/policy certain things need be separate.

## Proposed solution -- introducing region concept

To solve the above mentioned issues, we'll introduce managed kubecarrier federation. The basic concepts would be:

* ServiceCluster --> stays the same as before; its akin to availability zone.
* Region --> a group of multiple providers & service clusters regionally grouped with reasonable low-latency network connection is a region. The region has a single KubeCarrier management cluster

This would provide the logical, geographical, and failure domain grouping to acchive the initial motivational goals. How to coordinate and configure the global KubeCarrier solution from single place (or to the end user appearing as a single place) is still an open problem.

Between regions the following things should be synced, and should be globally unique:
* `(ServiceCluster, Account)`

Additionally the following objects should be replicated across every region:
* `(Account)`

Thus there must exist global consensus what the global KubeCarrier configuration should be. The authorization shall be handled the same as in the single region case --> management cluster shall create the necessary roles and rolebinding for the account.

This is enough to handle the authorization via the native k8s mechanism. For Authentication see the dedicated section.

## related work:

* https://github.com/kubermatic/kubecarrier/pull/291/files

This solutions revolves around having a single superCluster managing multiple regions with, and each region being managed by region-local installation.

Pros:
* the user authentication happens at the superCluster level, whilst their actions is propagated to region-local management cluster via the `Impersonate` headers
* installation is handled via the superCluster operator

Cons:
* the latency issue still remains. (e.g. request for creating new service instance would propagate from user -- superClusterManager -- region -- region local serviceCluster)
* the superCluster becomes Single point of Failure
* the superCluster has admin privileges in the region's management cluster. A exploit in our API server implementation could lead to unwanted privilege escalation.

## Proposed solution:

### Having no superCluster/Federations

* The latency and SPOF issue shall still reamain.
* We build API server accepting user's kubeconfig/token
    * This way we're preventing privilege escalation since API server doesen't have cluster-admin privileges.
* it's simpler approach

### build truly decentralized KubeCarrier installation with helper API/library for cross-region management

The way cross-region requires consists of two operation types:
* routing
* fan-out

Also there are two ways for keeping configuration in sync:
* static configuration
* central configuration plane (akin to superCluster in the original @nico proposal)

The routing operations require the client be routed to the right region executing its request. (e.g. create new serviceCluster, tenant consuming new service, etc.). The typical workflow would work like this:

* client looking up the cache does he have a route in place
    * if not ask the central configuration plane
    * alternatively fail since the region are hard-coded for the client
* handle the authentication. The authorization is already handled by the account, and general KubeCarrier installation. The authentication could work in multiple ways:
    * each kubecarrier region management cluster having the same credentials (e.g. same OIDC settings, or same client CA, or whatever)
* perform the required operation

The fan-out operations are more complicated, but rarer. They include the following:
* adding/removing regions
* adding/removing accounts
* adding/removing serviceCluster (if we want pair account-serviceCluster globally unique)

Their application differ depending on the central configuration store presence. In case the central configuration store is present, add the fan-out operation to the write-ahead log, perform it client side, and report the result. Upon next fan-out operation check the write-ahead log for any possible failure and required clean ups.

In case no central configuration store is present, due to limited operation types that can be done using the known distributed system techniques, it's also a non-issue. For service cluster, for example could work like this:

* append this operation to write-ahead log
* create ServiceCluster reservation in every cluster
* convert the target ServiceCluster from reservation to the configuration
* delete the write-ahead log entry

This approach has the following pros at the extra complexity:

* latency --> is minimized since there's no superCluster all requests must go through
* no single point of failure --> whole region failure doesn't impact other regions one tiny bit
* synchronicity for superCluster configuration --> there's no async waiting if the operation would succeed.

Cons:
* complexity

Many operations would be packed into a single library, which then can be exposed as an API server, CLI tool, etc. It's the most scalable, performant and resilient solution out of currently proposed, but required more rigorous modeling for correctness. Other solutions might appear simpler at a surface, but could hide unwanted skeletons in their closets.
