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
