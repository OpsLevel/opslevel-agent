<p align="center">
    <a href="https://github.com/OpsLevel/opslevel-agent/blob/main/LICENSE">
        <img src="https://img.shields.io/github/license/OpsLevel/opslevel-agent.svg" alt="License" /></a>
    <a href="https://GitHub.com/OpsLevel/opslevel-agent/releases/">
        <img src="https://img.shields.io/github/v/release/OpsLevel/opslevel-agent" alt="Release" /></a>
    <a href="https://masterminds.github.io/stability/experimental.html">
        <img src="https://masterminds.github.io/stability/experimental.svg" alt="Stability: Experimental" /></a>
    <a href="https://github.com/OpsLevel/opslevel-agent/graphs/contributors">
        <img src="https://img.shields.io/github/contributors/OpsLevel/opslevel-agent" alt="Contributors" /></a>
    <a href="https://github.com/OpsLevel/opslevel-agent/pulse">
        <img src="https://img.shields.io/github/commit-activity/m/OpsLevel/opslevel-agent" alt="Activity" /></a>
    <a href="https://github.com/OpsLevel/opslevel-agent/releases">
        <img src="https://img.shields.io/github/downloads/OpsLevel/opslevel-agent/total" alt="Downloads" /></a>
</p>

[![Overall](https://img.shields.io/endpoint?style=flat&url=https%3A%2F%2Fapp.opslevel.com%2Fapi%2Fservice_level%2FjcZ9Qt0e3fce3G6Xbo767Z2tXbKKKZ6qsRGzHZWwRME)](https://app.opslevel.com/services/opslevel_agent/maturity-report)

# opslevel-agent
Main repository for the OpsLevel agent

The OpsLevel agent is a multi-facet application that runs in your cloud and connects with OpsLevel to support multiple
functionalities for OpsLevel.  The current functionalities are:

- Kubernetes Service Detection

## Quickstart

<- TODO ->

### Metrics

| Name                            | Type        | Description                                                   |
|---------------------------------|-------------|---------------------------------------------------------------|






---
extractors:
- external_kind: apps_v1_Deployment
  external_id: ".metadata.uid"
---
transforms:
- external_kind: apps_v1_Deployment
  opslevel_kind: service
  opslevel_identifier: ".metadata.name"
  on_component_not_found: suggest
  properties:
  namespace: ".metadata.namespace"
  containers: ".spec.template.spec | .containers + .initContainers | map(.image)"