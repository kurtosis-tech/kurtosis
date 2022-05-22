package consts

const (
	CreateKubernetesVerb = "create"
	UpdateKubernetesVerb = "update"
	PatchKubernetesVerb  = "patch"
	DeleteKubernetesVerb = "delete"
	GetKubernetesVerb    = "get"
	ListKubernetesVerb   = "list"
	WatchKubernetesVerb  = "watch"

	NamespacesKubernetesResource             = "namespaces"
	ServiceAccountsKubernetesResource        = "serviceaccounts"
	RolesKubernetesResource                  = "roles"
	RoleBindingsKubernetesResource           = "rolebindings"
	PodsKubernetesResource                   = "pods"
	PodExecsKubernetesResource                   = "pods/exec"
	PodLogsKubernetesResource                   = "pods/log"
	ServicesKubernetesResource               = "services"
	PersistentVolumeClaimsKubernetesResource = "persistentvolumeclaims"
	JobsKubernetesResource = "jobs"

	ClusterRoleKubernetesResourceType = "ClusterRole"
	RoleKubernetesResourceType = "Role"

	RbacAuthorizationApiGroup = "rbac.authorization.k8s.io"
)
