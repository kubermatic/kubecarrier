# KubeCarrier

<p align="center">
  <img src="docs/img/KubeCarrier.png" width="700px" />
</p>

<p align="center">
  <img src="https://img.shields.io/github/license/kubermatic/kubecarrier"/>
  <img src="https://img.shields.io/github/go-mod/go-version/kubermatic/kubecarrier"/>
  <a href="https://github.com/kubermatic/kubecarrier/releases">
    <img src="https://img.shields.io/github/v/release/kubermatic/kubecarrier"/>
  </a>
  <a href="https://docs.kubermatic.com/kubecarrier">
    <img src="https://img.shields.io/badge/documentation-docs.kubermatic.io-blue"/>
  </a>
</p>

---

KubeCarrier is an open source system for managing applications and services across multiple Kubernetes Clusters; providing a framework to centralize the management of services and provide these services with external users in a self service hub.

---

- [Project Status](#project-status)
- [Features](#features)
- [Documentation](#documentation)
- [Contributing](#contributing)
  - [Before you start](#before-you-start)
  - [Pull Requests](#pull-requests)
- [FAQ](#faq)
- [Changelog](#changelog)

---


## Project Status

KubeCarrier is currently in early development and is not ready for production use, the APIs are not yet final and breaking changes might be introduced in every release.

## Features

- Cross Cluster Management of CRD instances
- Service Hub
- Multi Tenancy
- Account Management
- Integration with any existing operator

## Documentation

KubeCarrier is documented on [docs.kubermatic.io](https://docs.kubermatic.io) alongside our other open source projects.

## Troubleshooting

If you encounter issues [file an issue][1] or talk to us on the [#kubecarrier channel][12] on the [Kubermatic Slack][15].

## Contributing

Thanks for taking the time to join our community and start contributing!
Feedback and discussion are available on [the mailing list][11].

### Before you start

* Please familiarize yourself with the [Code of Conduct][4] before contributing.
* See [CONTRIBUTING.md][2] for instructions on the developer certificate of origin that we require.

### Pull requests

* We welcome pull requests. Feel free to dig through the [issues][1] and jump in.

## FAQ

### What`s the difference to OLM / Crossplane?

The [Operator Lifecycle Manager](https://github.com/operator-framework/operator-lifecycle-manager) from RedHat and [Crossplane](https://crossplane.io/) are both projects that manage installation, upgrade and deletion of Operators and their CustomResourceDefinitions in a Kubernetes cluster.

KubeCarrier on the other hand is just working with existing CustomResourceDefinitions and already installed Operators.
As both OLM and Crossplane are driven by CRDs, they can be combined with KubeCarrier to manage their configuration across clusters.

### What`s the difference to KubeFed - Kubernetes Federation?

The [Kubernetes Federation Project](https://github.com/kubernetes-sigs/kubefed) was created to distribute Workload across Kubernetes Clusters for e.g. geo-replication and disaster recovery.
It's intentionally low-level to work for generic workload to be spread across clusters.

While KubeCarrier is also operating on multiple clusters, KubeCarrier operates on a higher abstraction level.
KubeCarrier assigns applications onto single pre-determined Kubernetes clusters. Kubernetes Operators that enable these applications, may still use KubeFed underneath to spread the workload across clusters.

## Changelog

See [the list of releases][3] to find out about feature changes.

[1]: https://github.com/kubermatic/kubecarrier/issues
[2]: https://github.com/kubermatic/kubecarrier/blob/master/CONTRIBUTING.md
[3]: https://github.com/kubermatic/kubecarrier/releases
[4]: https://github.com/kubermatic/kubecarrier/blob/master/CODE_OF_CONDUCT.md

[11]: https://groups.google.com/forum/#!forum/loodse-dev
[12]: https://kubermatic.slack.com/messages/kubecarrier
[15]: http://slack.kubermatic.io/
