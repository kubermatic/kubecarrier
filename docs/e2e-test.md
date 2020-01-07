# E2E Test

KubeCarrier e2e test structure is based on [testify suite](https://github.com/stretchr/testify#suite-package).
KubeCarrier has multiple test "phases", and these "phases" need to be executed after each other.
For the start, every "phase" can map to a TestSuite each, and if we have more and more tests we can split them up into multiple TestSuites per phase.

Here is an overview of our e2e test:
- Setup/Installation (KubeCarrier Operator/Controller Manager Installation)
- Admin Operations (Tenant/Provider Creation/Deletion)
- Provider Operations (ServiceCluster, Catalog, etc)
- Tenant Operations (CatalogEntries, creating instances)
- Remote Clients and Integrations
