# problem statement:

- service/managed-tenants pipeline reconciliation scope is too broad
  - when one addon is changed, checks (and some deploy tasks) run for all addons -> this takes time
  - when one addon is changed, all indexes and bundles are rebuilt for all addons that still have their bundles in service/managed-tenants
- gitops does not provide status reporting
- gitops does provide a nice history of things though (do we need to keep that?)
  - we could keep that with argocd BUT then visibility on status-reporting is lost again

# resolution draft:

- have addonMetada

# TODOS:

- validate AddonMetadataVersion

  current validations: https://gitlab.cee.redhat.com/service/managed-tenants/-/tree/main/tasks/check
  check|statically verifiable
  ---|---
  Check if the operatorName in the addon.yaml file matches the currentCSV name|yes
  Checks if the metadata fields that should be unique are actually unique|yes
  Checks if the namespaces are defined as expected.|yes
  Checks if the testHarness image defined in the addon metadata is available.|yes
  Testing if the addon icon is valid.|yes
  Checks whether the ocmQuotaName defined in the addon metadata has an addon SKU in the OCM.|no
  Validates the default_value defined in the addon parameter metadata is valid for the parameters validation regex or the options defined|yes

- apply current AddonMetadataVersion to hive




Requirements:
- all OLM resources MUST stay intact
- switch must happen from _SSS that directly owns resources_ to _SSS that owns Addon object that owns resources_
- we must know that the switch happened for everything

proposal:
- phased approach
- scoped per: _addon, hive-shard_
- phase 0: ensure addon-operator is rolled out to all clusters in the shard
- phase 1: switch SSS.spec.resourceApplyMode to `Upsert` to prevent object cleanup on deletion
- phase 2: take backup of SSS, then delete
- phase 3: create new SSS with Addon object (TODO: addon.spec.resourceAdoptionStrategy.type = AdoptAll)
- phase 4: measure rollout progress in telemeter (TODO: metrics in addon operator)

questions:
- how do we coordinate the switch? doing it over gitops will be really hard
- should we go into the meta-hive sphere?


a central addon-metadata-operator deployment could:
- manage SSS objects in hive-shards via something like Nico's fleet-operator approach
- observe metrics from telemeter (TODO: clarify how programmatic access to telemeter is possible)
