# CLO Terraform provider coverage

Reviewed on 2026-11-19 against release **v2.8.0**, commit `11744b8f4a86a6081acbe99c6a525c4ef8741ae8`. The provider registers 17 resources and 27 data sources. Of 129 operations in the public v2 OpenAPI document, **91 have a Terraform call path, 1 has an unused SDK adapter, and 37 have no provider call**. A call path may serve only refresh, polling or deletion. It does not imply a standalone data source or full control of the request. [Registry release](https://registry.terraform.io/providers/clo-ru/clo/2.8.0) · [Registration](clo/provider.go) · [CLO guide](https://clo.ru/api) · [Live OpenAPI](https://api.clo.ru/openapi.json)

## Fork implementation status

On 2026-11-19 all 129 live OpenAPI operation IDs matched production SDK calls in this fork. The inventory below records the original upstream v2.8.0 comparison. Endpoint coverage alone does not establish field, import, pagination, lifecycle or live-service parity.

| Formerly missing operation group | Terraform entry point in this fork |
|---|---|
| Project create, update metadata, start, stop, delete | [Project resource](docs/resources/project.md) |
| Account-wide and single-project limit updates | [Account limits](docs/resources/account_limits.md), [project limits](docs/resources/project_limits.md) |
| Server rescue and reboot | [Instance action receipt](docs/resources/compute_instance_action.md) |
| Server license create, update and delete | [Instance license](docs/resources/compute_instance_license.md) |
| Volume metadata update | [Volume](docs/resources/disks_volume.md) |
| Address bandwidth update | [Address](docs/resources/network_ip.md), [attachment](docs/resources/network_ip_attach.md) |
| S3 suspend and unsuspend | [S3 user](docs/resources/storage_s3_user.md) |
| Project-wide DBaaS database listing | [Database lookup](docs/data-sources/dbaas_databases.md) |
| 19 GET endpoints, guest-agent check and console session | [Registered data sources](clo/provider.go), [API lookup schemas](clo/data_source_api_schemas.go) |

## Upstream gaps that affect Terraform plans

| Area | Effect | Source |
|---|---|---|
| Server destruction | Destroying an instance or snapshot-restored server requests deletion of its reported addresses and every attached disk with storage type `volume`. The selection uses current attachments, including volumes attached after creation. HCL cannot select which to retain or set the API’s `clear_fstab` option. | [Instance](clo/resource_instance.go), [restore](clo/resource_snapshot_restore.go), [delete request](internal/cloapi/server.go) |
| Server creation | The API accepts an existing boot `volume`, `user_data` and `flavor.cpu_type`. HCL requires an image and block devices and exposes none of those three inputs. HCL accepts `licenses.value`, but the builder discards it. The API license item has only name and addon. | [Resource](clo/resource_instance.go), [request](internal/cloapi/server.go) |
| Instance updates and refresh | Keypairs and address bandwidth feed creation. Editing those attributes alone has no corresponding update call. Refresh reads ID, name, status and switch status, leaving flavor changes made outside Terraform undetected by the resource. The instance data sources return a wider set of details. | [Resource](clo/resource_instance.go), [data source](clo/data_source_instance.go) |
| Volumes | Changing the name forces replacement despite the API supporting in-place name and description updates. HCL has no description input. Omitting the name selects autorename implicitly. Extend cannot request a server reboot. Destroy sends an empty body, so force and fstab cleanup cannot be selected. Detach always requests force and omits fstab cleanup. Resource refresh skips size and name. Attachment creation stores the device path, but refresh overwrites it with the attached server ID. | [Resource](clo/resource_volume.go), [attachment](clo/resource_volume_attach.go), [requests](internal/cloapi/volume.go) |
| IP addresses | Standalone attachment omits `bandwidth_max_mbps`. Bandwidth is read-only on the address resource. Changing `is_primary` from true to false invokes the same make-primary action as changing it to true. Attachment refresh reads primary/status/address but does not verify the attached target. Address refresh skips the configured PTR. | [Attachment](clo/resource_ip_attach.go), [address](clo/resource_ip.go), [requests](internal/cloapi/ip.go) |
| Load balancers | The API accepts initial listener rules. The provider sends an empty rule list and creates rules through separate resources. Project-wide rule discovery and balancer statistics are unavailable. | [Request builder](internal/cloapi/loadbalancer.go), [rule resource](clo/resource_loadbalancer_rule.go) |
| DBaaS | The API accepts an initial database in the cluster-create request. Terraform creates databases separately. Database listing requires a cluster ID despite an implemented project-wide list adapter. | [Cluster builder](internal/cloapi/dbaas_cluster.go), [database lookup](clo/data_source_dbaas_databases.go) |
| Collection queries | The guide documents filters, ordering, limit and offset where accepted. List adapters issue one request and read its result. HCL has no general server-side query or page traversal facility. Collections can be incomplete if the API paginates them. | [Project list](internal/cloapi/project.go), [volume list](internal/cloapi/volume.go), [guide](https://clo.ru/api) |
| Adoption | None of the registered resources declares a Terraform importer. Reading an existing object through a data source does not adopt its lifecycle. | [Provider resources](clo) |

## Read-side differences

| API result | Information lost or transformed |
|---|---|
| Project | The list drops display name, description and desired switch state. Creation time is exposed as `created_in`. [Project adapter](internal/cloapi/project.go) |
| Server | Drops admin-key status, datacenter, primary address, snapshot IDs, flavor disk and CPU type. Image becomes a distribution/version string. Image requirements and OS family are lost. Recipe becomes its name, losing status and minimum requirements. The instance data sources omit switch status even though the managed resource exposes it. [Adapter](internal/cloapi/server.go), [lookup](clo/data_source_instance.go), [list](clo/data_source_instances.go) |
| Volume | Drops filesystem, image, recipe, snapshots and licenses. Attachment is reduced to device and server ID. [Adapter](internal/cloapi/volume.go) |
| Address | Drops updated timestamp, external flag, MAC, mask, gateway, IP version and billing item ID. [Adapter](internal/cloapi/ip.go) |
| S3 user | Drops API-reported project and switch state. The API supports suspend and unsuspend independently of quota and name updates. Optional quota fields guarded by positive-value checks cannot transmit explicit zero through those paths. [Adapter](internal/cloapi/s3.go) |
| DBaaS backup | Drops the datastore ID and the backed-up databases collection, including database names and admin usernames. [Adapter](internal/cloapi/dbaas_backup.go) |

The DBaaS config data source converts API config values to strings. It renders null as an empty string and serializes lists and objects as JSON. Consumers lose the original scalar types. [Config data source](clo/data_source_dbaas_cluster_config.go)

Creating the S3 keys resource rotates credentials. Refresh updates only the access key, so external rotation can leave a stale secret paired with the new access key in state. Destroy removes the state entry without revoking the credentials. [S3 keys lifecycle](clo/resource_s3_user_keys.go)

## Operation inventory

The v2 guide contains 118 method/path pairs. The live specification contains those 118 plus 11 marked “OpenAPI only” below. Source comparison matched verb and path after normalizing path parameter names, then matched operation IDs to SDK calls in production Go files. The SDK module major version is v3, while the HTTP routes are v2. This is a static comparison. No cloud credentials or live resource operations were used.


### Missing Terraform operations (38)

| Service | Method and path | API operation | Guide | Implementation |
|---|---|---|---|---|
| account | `GET /v2/balance` | AccountBalance | v2 | No provider call |
| account | `GET /v2/limits` | AccountLimits | v2 | No provider call |
| account | `PATCH /v2/limits` | AccountPatchLimits | v2 | No provider call |
| account | `GET /v2/limits/projects` | AccountProjectsLimits | v2 | No provider call |
| account | `GET /v2/stat` | AccountStat | v2 | No provider call |
| project | `POST /v2/projects` | ProjectCreate | v2 | No provider call |
| project | `GET /v2/projects/{object_id}/detail` | ProjectDetail | OpenAPI only | No provider call |
| project | `GET /v2/projects/{object_id}/limits` | ProjectLimitsList | v2 | No provider call |
| project | `PATCH /v2/projects/{object_id}/limits` | ProjectPatchLimits | v2 | No provider call |
| server | `GET /v2/projects/{object_id}/servers/related-resources` | ProjectServerConfigDep | v2 | No provider call |
| dbaas | `GET /v2/projects/{object_id}/dbaas/related-resources` | ProjectDbaasConfigDep | v2 | No provider call |
| infrastructure_module | `GET /v2/projects/{object_id}/params` | ProjectInfrastructureModuleConstants | OpenAPI only | No provider call |
| project | `DELETE /v2/projects/{object_id}` | ProjectDelete | OpenAPI only | No provider call |
| project | `PATCH /v2/projects/{object_id}` | ProjectPatchDisplayNameDescription | OpenAPI only | No provider call |
| project | `POST /v2/projects/{object_id}/stop` | ProjectStop | OpenAPI only | No provider call |
| project | `POST /v2/projects/{object_id}/start` | ProjectStart | OpenAPI only | No provider call |
| project | `GET /v2/projects/{object_id}/consumption` | ProjectConsumption | v2 | No provider call |
| network | `GET /v2/projects/{object_id}/networks` | NetworksList | v2 | No provider call |
| server | `POST /v2/servers/{object_id}/rescue` | ServerRescue | v2 | No provider call |
| server | `POST /v2/servers/{object_id}/reboot` | ServerReboot | v2 | No provider call |
| server | `POST /v2/servers/{object_id}/check-guest-agent` | ServerCheckGuestAgent | OpenAPI only | No provider call |
| server | `POST /v2/servers/{object_id}/console` | ServerConsole | v2 | No provider call |
| local disk | `GET /v2/projects/{object_id}/local-disks` | ProjectLocalDisksList | v2 | No provider call |
| local disk | `GET /v2/local-disks/{object_id}/detail` | LocalDiskDetail | v2 | No provider call |
| volume | `PATCH /v2/volumes/{object_id}` | VolumeUpdate | v2 | No provider call |
| address | `POST /v2/addresses/{object_id}/bandwidth` | AddressChangeBandwidth | OpenAPI only | No provider call |
| license | `GET /v2/licenses` | AvailableLicensesList | v2 | No provider call |
| server | `GET /v2/servers/{object_id}/licenses` | ServerLicenses | v2 | No provider call |
| license | `POST /v2/servers/{object_id}/licenses` | ServerAddLicense | v2 | No provider call |
| license | `GET /v2/licenses/{object_id}` | LicenseDetails | v2 | No provider call |
| license | `DELETE /v2/licenses/{object_id}` | LicenseDelete | v2 | No provider call |
| license | `PATCH /v2/licenses/{object_id}` | LicenseUpdate | v2 | No provider call |
| s3 | `POST /v2/s3/users/{object_id}/suspend` | S3UserSuspend | v2 | No provider call |
| s3 | `POST /v2/s3/users/{object_id}/unsuspend` | S3UserUnsuspend | v2 | No provider call |
| loadbalancer | `GET /v2/projects/{object_id}/loadbalancers/rules` | ProjectRuleList | v2 | No provider call |
| loadbalancer | `GET /v2/loadbalancers/{object_id}/stat` | LoadBalancerStat | OpenAPI only | No provider call |
| dbaas | `GET /v2/projects/{object_id}/dbaas/databases` | ProjectDbaasDatabasesList | v2 | [ListDatabasesByProject](internal/cloapi/dbaas_database.go) (no Terraform caller) |
| maintenance | `GET /v2/maintenance-mode` | MaintenanceModeStatus | OpenAPI only | No provider call |

### Operations with a Terraform call path (91)

| Service | Method and path | API operation | Guide | Implementation |
|---|---|---|---|---|
| account | `GET /v2/projects` | ProjectList | v2 | [ListProjects](internal/cloapi/project.go) · [data_source_projects](clo/data_source_projects.go) |
| project | `GET /v2/projects/{object_id}/images` | ProjectImagesList | v2 | [ListImages](internal/cloapi/image.go) · [data_source_image](clo/data_source_image.go), [data_source_images](clo/data_source_images.go) |
| project | `GET /v2/projects/{object_id}/recipes` | ProjectRecipes | v2 | [ListRecipes](internal/cloapi/recipe.go) · [data_source_recipe](clo/data_source_recipe.go), [data_source_recipes](clo/data_source_recipes.go) |
| server | `GET /v2/projects/{object_id}/servers` | ProjectServerList | v2 | [ListServers](internal/cloapi/server.go) · [data_source_instances](clo/data_source_instances.go) |
| server | `POST /v2/projects/{object_id}/servers` | ServerCreate | v2 | [CreateServer](internal/cloapi/server.go) · [resource_instance](clo/resource_instance.go) |
| server | `DELETE /v2/servers/{object_id}` | ServerDelete | v2 | [DeleteServer](internal/cloapi/server.go) · [resource_instance](clo/resource_instance.go), [resource_snapshot_restore](clo/resource_snapshot_restore.go) |
| server | `PATCH /v2/servers/{object_id}` | ServerUpdate | v2 | [RenameServer](internal/cloapi/server.go) · [resource_instance](clo/resource_instance.go) |
| server | `POST /v2/servers/{object_id}/start` | ServerStart | v2 | [StartServer](internal/cloapi/server.go) · [resource_instance_power](clo/resource_instance_power.go) |
| server | `POST /v2/servers/{object_id}/stop` | ServerStop | v2 | [StopServer](internal/cloapi/server.go) · [resource_instance_power](clo/resource_instance_power.go) |
| server | `POST /v2/servers/{object_id}/resize` | ServerResize | v2 | [ResizeServer](internal/cloapi/server.go) · [resource_instance](clo/resource_instance.go) |
| server | `GET /v2/servers/{object_id}/detail` | ServerDetail | v2 | [GetServer](internal/cloapi/server.go) · [data_source_instance](clo/data_source_instance.go), [resource_instance](clo/resource_instance.go), [resource_instance_power](clo/resource_instance_power.go), [resource_snapshot_restore](clo/resource_snapshot_restore.go) |
| server | `POST /v2/servers/{object_id}/password` | ServerChangePassword | v2 | [ChangeServerPassword](internal/cloapi/server.go) · [resource_instance](clo/resource_instance.go) |
| volume | `GET /v2/projects/{object_id}/volumes` | ProjectVolumesList | v2 | [ListVolumes](internal/cloapi/volume.go) · [data_source_volumes](clo/data_source_volumes.go) |
| volume | `POST /v2/projects/{object_id}/volumes` | VolumeCreate | v2 | [CreateVolume](internal/cloapi/volume.go) · [resource_volume](clo/resource_volume.go) |
| volume | `GET /v2/volumes/{object_id}/detail` | VolumeDetail | v2 | [GetVolume](internal/cloapi/volume.go) · [data_source_volume](clo/data_source_volume.go), [resource_volume](clo/resource_volume.go), [resource_volume_attach](clo/resource_volume_attach.go) |
| volume | `DELETE /v2/volumes/{object_id}` | VolumeDelete | v2 | [DeleteVolume](internal/cloapi/volume.go) · [resource_volume](clo/resource_volume.go) |
| volume | `POST /v2/volumes/{object_id}/detach` | VolumeDetach | v2 | [DetachVolume](internal/cloapi/volume.go) · [resource_volume_attach](clo/resource_volume_attach.go) |
| volume | `POST /v2/volumes/{object_id}/attach` | VolumeAttach | v2 | [AttachVolume](internal/cloapi/volume.go) · [resource_volume_attach](clo/resource_volume_attach.go) |
| volume | `POST /v2/volumes/{object_id}/extend` | VolumeExtend | v2 | [ExtendVolume](internal/cloapi/volume.go) · [resource_volume](clo/resource_volume.go) |
| address | `GET /v2/projects/{object_id}/addresses` | ProjectAddressesList | v2 | [ListAddresses](internal/cloapi/ip.go) · [data_source_ips](clo/data_source_ips.go) |
| address | `POST /v2/projects/{object_id}/addresses` | AddressCreate | v2 | [CreateAddress](internal/cloapi/ip.go) · [resource_ip](clo/resource_ip.go) |
| address | `GET /v2/addresses/{object_id}/detail` | AddressDetail | v2 | [GetAddress](internal/cloapi/ip.go) · [data_source_ip](clo/data_source_ip.go), [resource_ip](clo/resource_ip.go), [resource_ip_attach](clo/resource_ip_attach.go) |
| address | `DELETE /v2/addresses/{object_id}` | AddressDelete | v2 | [DeleteAddress](internal/cloapi/ip.go) · [resource_ip](clo/resource_ip.go) |
| address | `POST /v2/addresses/{object_id}/attach` | AddressAttach | v2 | [AttachAddress](internal/cloapi/ip.go) · [resource_ip_attach](clo/resource_ip_attach.go) |
| address | `POST /v2/addresses/{object_id}/detach` | AddressDetach | v2 | [DetachAddress](internal/cloapi/ip.go) · [resource_ip_attach](clo/resource_ip_attach.go) |
| address | `POST /v2/addresses/{object_id}/primary` | AddressSetPrimary | v2 | [SetAddressPrimary](internal/cloapi/ip.go) · [resource_ip_attach](clo/resource_ip_attach.go) |
| address | `PUT /v2/addresses/{object_id}/ptr` | AddressEditPtr | v2 | [ChangeAddressPtr](internal/cloapi/ip.go) · [resource_ip](clo/resource_ip.go) |
| snapshot | `DELETE /v2/snapshots/{object_id}` | SnapshotDelete | v2 | [DeleteSnapshot](internal/cloapi/snapshot.go) · [resource_snapshot](clo/resource_snapshot.go) |
| snapshot | `GET /v2/projects/{object_id}/snapshots` | SnapshotsList | v2 | [ListSnapshots](internal/cloapi/snapshot.go) · [data_source_snapshots](clo/data_source_snapshots.go) |
| snapshot | `GET /v2/snapshots/{object_id}/detail` | SnapshotDetails | v2 | [GetSnapshot](internal/cloapi/snapshot.go) · [resource_snapshot](clo/resource_snapshot.go) |
| snapshot | `POST /v2/snapshots/{object_id}/restore` | SnapshotRestore | v2 | [RestoreSnapshot](internal/cloapi/snapshot.go) · [resource_snapshot_restore](clo/resource_snapshot_restore.go) |
| snapshot | `POST /v2/servers/{object_id}/snapshot` | CreateServerSnapshot | v2 | [CreateSnapshot](internal/cloapi/snapshot.go) · [resource_snapshot](clo/resource_snapshot.go) |
| keypair | `GET /v2/projects/{object_id}/keypairs` | KeyPairsList | v2 | [ListKeypairs](internal/cloapi/keypair.go) · [data_source_keypair](clo/data_source_keypair.go), [data_source_keypairs](clo/data_source_keypairs.go) |
| keypair | `POST /v2/projects/{object_id}/keypairs` | ImportKeypair | v2 | [ImportKeypair](internal/cloapi/keypair.go) · [resource_keypair](clo/resource_keypair.go) |
| keypair | `GET /v2/keypairs/{object_id}/detail` | KeypairDetail | v2 | [GetKeypair](internal/cloapi/keypair.go) · [resource_keypair](clo/resource_keypair.go) |
| keypair | `DELETE /v2/keypairs/{object_id}` | KeypairDelete | v2 | [DeleteKeypair](internal/cloapi/keypair.go) · [resource_keypair](clo/resource_keypair.go) |
| keypair | `POST /v2/projects/{object_id}/keypairs/generate` | GenerateKeypair | v2 | [GenerateKeypair](internal/cloapi/keypair.go) · [resource_keypair](clo/resource_keypair.go) |
| s3 | `GET /v2/s3/users/{object_id}/credentials` | S3GetUserKeys | v2 | [GetS3UserAccessKey](internal/cloapi/s3.go) · [data_source_s3_user_keys](clo/data_source_s3_user_keys.go), [resource_s3_user_keys](clo/resource_s3_user_keys.go) |
| s3 | `POST /v2/s3/users/{object_id}/credentials` | S3GenUserKeys | v2 | [GenS3UserKeys](internal/cloapi/s3.go) · [resource_s3_user_keys](clo/resource_s3_user_keys.go) |
| s3 | `GET /v2/projects/{object_id}/s3/users` | S3UsersList | v2 | [ListS3Users](internal/cloapi/s3.go) · [data_source_s3_users](clo/data_source_s3_users.go) |
| s3 | `POST /v2/projects/{object_id}/s3/users` | S3UserCreate | v2 | [CreateS3User](internal/cloapi/s3.go) · [resource_s3_user](clo/resource_s3_user.go) |
| s3 | `DELETE /v2/s3/users/{object_id}` | S3UserDelete | v2 | [DeleteS3User](internal/cloapi/s3.go) · [resource_s3_user](clo/resource_s3_user.go) |
| s3 | `PATCH /v2/s3/users/{object_id}` | S3UserUpdate | v2 | [UpdateS3UserName](internal/cloapi/s3.go) · [resource_s3_user](clo/resource_s3_user.go) |
| s3 | `GET /v2/s3/users/{object_id}/detail` | S3UserDetails | v2 | [GetS3User](internal/cloapi/s3.go) · [data_source_s3_user](clo/data_source_s3_user.go), [resource_s3_user](clo/resource_s3_user.go) |
| s3 | `PUT /v2/s3/users/{object_id}/quotas` | S3UserUpdateQuota | v2 | [UpdateS3UserQuota](internal/cloapi/s3.go) · [resource_s3_user](clo/resource_s3_user.go) |
| loadbalancer | `GET /v2/projects/{object_id}/loadbalancers` | LoadBalancerList | v2 | [ListLoadBalancers](internal/cloapi/loadbalancer.go) · [data_source_loadbalancers](clo/data_source_loadbalancers.go) |
| loadbalancer | `POST /v2/projects/{object_id}/loadbalancers` | LoadBalancerCreate | v2 | [CreateLoadBalancer](internal/cloapi/loadbalancer.go) · [resource_loadbalancer](clo/resource_loadbalancer.go) |
| loadbalancer | `DELETE /v2/loadbalancers/{object_id}` | LoadBalancerDelete | v2 | [DeleteLoadBalancer](internal/cloapi/loadbalancer.go) · [resource_loadbalancer](clo/resource_loadbalancer.go) |
| loadbalancer | `PATCH /v2/loadbalancers/{object_id}` | LoadBalancerRename | v2 | [RenameLoadBalancer](internal/cloapi/loadbalancer.go) · [resource_loadbalancer](clo/resource_loadbalancer.go) |
| loadbalancer | `GET /v2/loadbalancers/{object_id}/detail` | LoadBalancerDetail | v2 | [GetLoadBalancer](internal/cloapi/loadbalancer.go) · [resource_loadbalancer](clo/resource_loadbalancer.go) |
| loadbalancer | `POST /v2/loadbalancers/{object_id}/algorithm` | LoadBalancerUpdate | v2 | [UpdateLoadBalancer](internal/cloapi/loadbalancer.go) · [resource_loadbalancer](clo/resource_loadbalancer.go) |
| loadbalancer | `PUT /v2/loadbalancers/{object_id}/healthmonitor` | LoadBalancerUpdateHealthmonitor | v2 | [UpdateHealthmonitor](internal/cloapi/loadbalancer.go) · [resource_loadbalancer](clo/resource_loadbalancer.go) |
| loadbalancer | `POST /v2/loadbalancers/{object_id}/start` | LoadBalancerEnable | v2 | [EnableLoadBalancer](internal/cloapi/loadbalancer.go) · [resource_loadbalancer](clo/resource_loadbalancer.go) |
| loadbalancer | `POST /v2/loadbalancers/{object_id}/stop` | LoadBalancerStop | v2 | [StopLoadBalancer](internal/cloapi/loadbalancer.go) · [resource_loadbalancer](clo/resource_loadbalancer.go) |
| loadbalancer | `GET /v2/loadbalancers/{object_id}/rules` | RuleList | v2 | [ListRules](internal/cloapi/loadbalancer.go) · [data_source_loadbalancer_rules](clo/data_source_loadbalancer_rules.go) |
| loadbalancer | `POST /v2/loadbalancers/{object_id}/rules` | RuleCreate | v2 | [CreateRule](internal/cloapi/loadbalancer.go) · [resource_loadbalancer_rule](clo/resource_loadbalancer_rule.go) |
| loadbalancer | `GET /v2/loadbalancers/rules/{object_id}` | RuleDetail | v2 | [GetRule](internal/cloapi/loadbalancer.go) · [resource_loadbalancer_rule](clo/resource_loadbalancer_rule.go) |
| loadbalancer | `DELETE /v2/loadbalancers/rules/{object_id}` | RuleDelete | v2 | [DeleteRule](internal/cloapi/loadbalancer.go) · [resource_loadbalancer_rule](clo/resource_loadbalancer_rule.go) |
| dbaas | `GET /v2/projects/{object_id}/dbaas/datastores` | ProjectDbaasDatastores | v2 | [ListDatastores](internal/cloapi/dbaas_cluster.go) · [data_source_dbaas_datastores](clo/data_source_dbaas_datastores.go) |
| dbaas | `GET /v2/projects/{object_id}/dbaas/clusters` | DbaasClustersList | v2 | [ListClusters](internal/cloapi/dbaas_cluster.go) · [data_source_dbaas_clusters](clo/data_source_dbaas_clusters.go) |
| dbaas | `POST /v2/projects/{object_id}/dbaas/clusters` | DbaasClusterCreate | v2 | [CreateCluster](internal/cloapi/dbaas_cluster.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `GET /v2/dbaas/clusters/{object_id}` | DbaasClusterDetail | v2 | [GetCluster](internal/cloapi/dbaas_cluster.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `DELETE /v2/dbaas/clusters/{object_id}` | DbaasClusterDelete | v2 | [DeleteCluster](internal/cloapi/dbaas_cluster.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `PATCH /v2/dbaas/clusters/{object_id}` | DbaasClusterUpdate | v2 | [RenameCluster](internal/cloapi/dbaas_cluster.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `GET /v2/dbaas/clusters/{object_id}/configuration` | DbaasClusterConfig | OpenAPI only | [GetClusterConfig](internal/cloapi/dbaas_cluster.go) · [data_source_dbaas_cluster_config](clo/data_source_dbaas_cluster_config.go) |
| dbaas | `POST /v2/dbaas/clusters/{object_id}/stop` | DbaasClusterStop | v2 | [StopCluster](internal/cloapi/dbaas_cluster.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `POST /v2/dbaas/clusters/{object_id}/start` | DbaasClusterStart | v2 | [StartCluster](internal/cloapi/dbaas_cluster.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `POST /v2/dbaas/clusters/{object_id}/resize/resources` | DbaasClusterResize | v2 | [ResizeCluster](internal/cloapi/dbaas_cluster.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `POST /v2/dbaas/clusters/{object_id}/resize/storage` | DbaasClusterResizeStorage | v2 | [ResizeClusterStorage](internal/cloapi/dbaas_cluster.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `GET /v2/dbaas/clusters/{object_id}/databases` | ClusterDbaasDatabasesList | v2 | [ListDatabasesByCluster](internal/cloapi/dbaas_database.go) · [data_source_dbaas_databases](clo/data_source_dbaas_databases.go) |
| dbaas | `POST /v2/dbaas/clusters/{object_id}/databases` | ClusterAddDatabase | v2 | [CreateDatabase](internal/cloapi/dbaas_database.go) · [resource_dbaas_database](clo/resource_dbaas_database.go) |
| dbaas | `GET /v2/dbaas/clusters/{object_id}/nodes` | ClusterDbaasNodesList | v2 | [ListNodes](internal/cloapi/dbaas_cluster.go) · [data_source_dbaas_nodes](clo/data_source_dbaas_nodes.go) |
| dbaas | `POST /v2/dbaas/clusters/{object_id}/backup/enable` | DbaasClusterBackupEnable | v2 | [EnableClusterBackup](internal/cloapi/dbaas_backup.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `POST /v2/dbaas/clusters/{object_id}/backup/disable` | DbaasClusterBackupDisable | v2 | [DisableClusterBackup](internal/cloapi/dbaas_backup.go) · [resource_dbaas_cluster](clo/resource_dbaas_cluster.go) |
| dbaas | `POST /v2/dbaas/databases/{object_id}/backup/enable` | DbaasDatabaseBackupEnable | v2 | [EnableDatabaseBackup](internal/cloapi/dbaas_backup.go) · [resource_dbaas_database](clo/resource_dbaas_database.go) |
| dbaas | `POST /v2/dbaas/databases/{object_id}/backup/disable` | DbaasDatabaseBackupDisable | v2 | [DisableDatabaseBackup](internal/cloapi/dbaas_backup.go) · [resource_dbaas_database](clo/resource_dbaas_database.go) |
| dbaas | `POST /v2/dbaas/databases/{object_id}/backup/create` | ClusterDatabaseBackup | v2 | [CreateDatabaseBackup](internal/cloapi/dbaas_backup.go) · [resource_dbaas_backup](clo/resource_dbaas_backup.go) |
| dbaas | `GET /v2/dbaas/databases/{object_id}` | DbaasDatabaseDetail | v2 | [GetDatabase](internal/cloapi/dbaas_database.go) · [resource_dbaas_database](clo/resource_dbaas_database.go) |
| dbaas | `DELETE /v2/dbaas/databases/{object_id}` | DbaasDatabaseDelete | v2 | [DeleteDatabase](internal/cloapi/dbaas_database.go) · [resource_dbaas_database](clo/resource_dbaas_database.go) |
| dbaas | `POST /v2/dbaas/databases/{object_id}/restore` | DbaasRestoreAdminPassword | v2 | [RestoreAdminPassword](internal/cloapi/dbaas_database.go) · [resource_dbaas_database](clo/resource_dbaas_database.go) |
| dbaas | `GET /v2/projects/{object_id}/dbaas/backups` | ProjectBackupList | v2 | [ListBackups](internal/cloapi/dbaas_backup.go) · [data_source_dbaas_backups](clo/data_source_dbaas_backups.go) |
| dbaas | `GET /v2/dbaas/backups/{object_id}` | DbaasBackupDetail | v2 | [GetBackup](internal/cloapi/dbaas_backup.go) · [resource_dbaas_backup](clo/resource_dbaas_backup.go) |
| dbaas | `DELETE /v2/dbaas/backups/{object_id}` | DbaasBackupDelete | v2 | [DeleteBackup](internal/cloapi/dbaas_backup.go) · [resource_dbaas_backup](clo/resource_dbaas_backup.go) |
| dbaas | `POST /v2/dbaas/backups/{object_id}/download` | DbaasBackupDownload | v2 | [DownloadBackup](internal/cloapi/dbaas_backup.go) · [data_source_dbaas_backup_download](clo/data_source_dbaas_backup_download.go) |
| dbaas | `POST /v2/dbaas/clusters/{object_id}/backup/create` | DbaasClusterBackup | v2 | [CreateClusterBackup](internal/cloapi/dbaas_backup.go) · [resource_dbaas_backup](clo/resource_dbaas_backup.go) |
| vrouter | `GET /v2/projects/{object_id}/vrouters` | ProjectVrouterList | v2 | [ListVrouters](internal/cloapi/vrouter.go) · [data_source_vrouters](clo/data_source_vrouters.go) |
| vrouter | `POST /v2/projects/{object_id}/vrouters` | VrouterCreate | v2 | [CreateVrouter](internal/cloapi/vrouter.go) · [resource_vrouter](clo/resource_vrouter.go) |
| vrouter | `POST /v2/vrouters/{object_id}/stop` | VrouterStop | v2 | [StopVrouter](internal/cloapi/vrouter.go) · [resource_vrouter](clo/resource_vrouter.go) |
| vrouter | `POST /v2/vrouters/{object_id}/start` | VrouterStart | v2 | [StartVrouter](internal/cloapi/vrouter.go) · [resource_vrouter](clo/resource_vrouter.go) |
| vrouter | `GET /v2/vrouters/{object_id}` | VrouterDetail | v2 | [GetVrouter](internal/cloapi/vrouter.go) · [resource_vrouter](clo/resource_vrouter.go) |
| vrouter | `DELETE /v2/vrouters/{object_id}` | VrouterDelete | v2 | [DeleteVrouter](internal/cloapi/vrouter.go) · [resource_vrouter](clo/resource_vrouter.go) |

## Legacy guide

The guide also publishes 111 v1 operations. The provider targets v2 and has no v1 mode. Most legacy capabilities map onto the operation inventory above. Recipes and keypairs now use project-scoped discovery. Floating IP operations map to address operations, including PTR updates. Snapshot and keypair deletion use the v2 delete routes. S3 user operations moved to the v2 S3 paths, with suspend and unsuspend still missing. DBaaS enable/disable and node resize correspond to cluster power and resize operations. Node detail by ID has no dedicated Terraform lookup. Nodes are listed by cluster. The legacy server-scoped PTR action has no server-scoped Terraform equivalent.
