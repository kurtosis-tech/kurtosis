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
	ClusterRolesKubernetesResource           = "clusterroles"
	ClusterRoleBindingsKubernetesResource    = "clusterrolebindings"
	RolesKubernetesResource                  = "roles"
	RoleBindingsKubernetesResource           = "rolebindings"
	PodsKubernetesResource                   = "pods"
	PodExecsKubernetesResource               = "pods/exec"
	PodLogsKubernetesResource                = "pods/log"
	ServicesKubernetesResource               = "services"
	JobsKubernetesResource                   = "jobs"
	NodesKubernetesResource                  = "nodes"
	PersistentVolumesKubernetesResource      = "persistentvolumes"
	PersistentVolumeClaimsKubernetesResource = "persistentvolumeclaims"
	IngressesKubernetesResource              = "ingresses"

	ClusterRoleKubernetesResourceType = "ClusterRole"
	RoleKubernetesResourceType        = "Role"

	RbacAuthorizationApiGroup = "rbac.authorization.k8s.io"
)
