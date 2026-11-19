# A FULL backup of an entire managed-database cluster.
resource "clo_dbaas_backup" "cluster_backup" {
  cluster_id = clo_dbaas_cluster.cluster_1.id
  name       = "nightly-full"
}

# A PARTIAL backup of a single database. Set force_delete to remove it even when
# it is the parent of later incremental backups.
resource "clo_dbaas_backup" "database_backup" {
  database_id  = clo_dbaas_database.app.id
  name         = "app-only"
  force_delete = true
}
